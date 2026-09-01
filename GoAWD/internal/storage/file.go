package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type fileEntry struct {
	ID        string          `json:"id"`
	Data      json.RawMessage `json:"data"`
	CreatedAt int64           `json:"created_at"`
}

type File struct {
	mu       sync.RWMutex
	path     string
	data     map[string][]fileEntry
	index    map[string]map[string]int
	counts   map[string]int64
	flushMu  sync.Mutex
	dirty    bool
	stopCh   chan struct{}
}

func NewFile(path string) (*File, error) {
	if path == "" {
		path = "./goawd.json"
	}
	f := &File{
		path:   path,
		data:   make(map[string][]fileEntry),
		index:  make(map[string]map[string]int),
		counts: make(map[string]int64),
		stopCh: make(chan struct{}),
	}
	if err := f.load(); err != nil {
		return nil, err
	}
	go f.autoFlush()
	return f, nil
}

func (f *File) load() error {
	raw, err := os.ReadFile(f.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	var snapshot map[string][]fileEntry
	if err := json.Unmarshal(raw, &snapshot); err != nil {
		return err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	for coll, entries := range snapshot {
		f.data[coll] = entries
		f.index[coll] = make(map[string]int)
		for i, e := range entries {
			f.index[coll][e.ID] = i
		}
		f.counts[coll] = int64(len(entries))
	}
	return nil
}

func (f *File) autoFlush() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-f.stopCh:
			f.flush()
			return
		case <-ticker.C:
			f.flush()
		}
	}
}

func (f *File) flush() {
	f.flushMu.Lock()
	defer f.flushMu.Unlock()

	f.mu.RLock()
	if !f.dirty {
		f.mu.RUnlock()
		return
	}
	snapshot := make(map[string][]fileEntry, len(f.data))
	for coll, entries := range f.data {
		cp := make([]fileEntry, len(entries))
		copy(cp, entries)
		snapshot[coll] = cp
	}
	f.dirty = false
	f.mu.RUnlock()

	dir := filepath.Dir(f.path)
	if dir != "." && dir != "" {
		os.MkdirAll(dir, 0755)
	}
	tmp := f.path + ".tmp"
	data, err := json.Marshal(snapshot)
	if err != nil {
		return
	}
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return
	}
	os.Rename(tmp, f.path)
}

func (f *File) Save(_ context.Context, collection, id string, doc interface{}) error {
	data, err := json.Marshal(doc)
	if err != nil {
		return fmt.Errorf("marshal doc: %w", err)
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.index[collection] == nil {
		f.index[collection] = make(map[string]int)
	}
	entry := fileEntry{ID: id, Data: data, CreatedAt: time.Now().Unix()}
	if pos, ok := f.index[collection][id]; ok {
		f.data[collection][pos] = entry
	} else {
		f.data[collection] = append(f.data[collection], entry)
		f.index[collection][id] = len(f.data[collection]) - 1
		f.counts[collection]++
	}
	f.dirty = true
	return nil
}

func (f *File) Get(_ context.Context, collection, id string) (interface{}, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	idx, ok := f.index[collection]
	if !ok {
		return nil, nil
	}
	pos, ok := idx[id]
	if !ok {
		return nil, nil
	}
	var doc interface{}
	if err := json.Unmarshal(f.data[collection][pos].Data, &doc); err != nil {
		return nil, err
	}
	return doc, nil
}

func (f *File) Paginate(_ context.Context, collection string, page, count int) ([]interface{}, int64, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	entries := f.data[collection]
	total := int64(len(entries))
	page, offset, _ := NormalizePage(page, count, int(total))
	if count <= 0 {
		count = 20
	}
	if offset >= len(entries) {
		return []interface{}{}, total, nil
	}

	result := make([]interface{}, 0, count)
	for i := len(entries) - 1 - offset; i >= 0 && len(result) < count; i-- {
		var doc interface{}
		if err := json.Unmarshal(entries[i].Data, &doc); err == nil {
			result = append(result, doc)
		}
	}
	return result, total, nil
}

func (f *File) Count(_ context.Context, collection string) int64 {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.counts[collection]
}

func (f *File) All(_ context.Context, collection string) ([]interface{}, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	entries := f.data[collection]
	result := make([]interface{}, 0, len(entries))
	for _, e := range entries {
		var doc interface{}
		if err := json.Unmarshal(e.Data, &doc); err == nil {
			result = append(result, doc)
		}
	}
	return result, nil
}

func (f *File) Close() error {
	close(f.stopCh)
	return nil
}
