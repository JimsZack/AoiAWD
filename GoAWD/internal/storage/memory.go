package storage

import (
	"context"
	"sync"
)

type memEntry struct {
	id  string
	doc interface{}
}

type Memory struct {
	mu    sync.RWMutex
	data  map[string][]memEntry
	index map[string]map[string]int

	// maxEntries caps how many documents a collection keeps; the oldest are
	// evicted once the cap is exceeded.
	maxEntries int
}

// maxMemEntries is the default per-collection cap for the memory backend.
const maxMemEntries = 50000

func NewMemory() *Memory {
	return &Memory{
		data:       make(map[string][]memEntry),
		index:      make(map[string]map[string]int),
		maxEntries: maxMemEntries,
	}
}

func (m *Memory) Save(_ context.Context, collection, id string, doc interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.index[collection] == nil {
		m.index[collection] = make(map[string]int)
	}
	if pos, ok := m.index[collection][id]; ok {
		m.data[collection][pos].doc = doc
		return nil
	}
	m.data[collection] = append(m.data[collection], memEntry{id: id, doc: doc})
	m.index[collection][id] = len(m.data[collection]) - 1
	if m.evictLocked(collection) {
		m.reindexLocked(collection)
	}
	return nil
}

// evictLocked drops the oldest entries once a collection grows past its cap.
// It reports whether anything was dropped; the caller must reindex then.
//
// Eviction waits until a full batch has accumulated so the O(n) reindex is
// amortised across evictBatch inserts instead of running on every Save.
func (m *Memory) evictLocked(collection string) bool {
	entries := m.data[collection]
	drop := len(entries) - m.maxEntries
	if drop < evictBatch {
		return false
	}
	// Copy into a fresh slice so the evicted prefix can be garbage collected.
	kept := make([]memEntry, len(entries)-drop)
	copy(kept, entries[drop:])
	m.data[collection] = kept
	return true
}

func (m *Memory) reindexLocked(collection string) {
	entries := m.data[collection]
	idx := make(map[string]int, len(entries))
	for i, e := range entries {
		idx[e.id] = i
	}
	m.index[collection] = idx
}

func (m *Memory) Get(_ context.Context, collection, id string) (interface{}, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	idx, ok := m.index[collection]
	if !ok {
		return nil, nil
	}
	pos, ok := idx[id]
	if !ok {
		return nil, nil
	}
	return m.data[collection][pos].doc, nil
}

func (m *Memory) Paginate(_ context.Context, collection string, page, count int) ([]interface{}, int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	entries := m.data[collection]
	total := int64(len(entries))
	page, offset, _ := NormalizePage(page, count, int(total))

	if count <= 0 {
		count = 20
	}
	if offset >= len(entries) {
		return []interface{}{}, total, nil
	}

	reversed := make([]interface{}, 0, count)
	for i := len(entries) - 1 - offset; i >= 0 && len(reversed) < count; i-- {
		reversed = append(reversed, entries[i].doc)
	}
	return reversed, total, nil
}

func (m *Memory) Count(_ context.Context, collection string) int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return int64(len(m.data[collection]))
}

func (m *Memory) All(_ context.Context, collection string) ([]interface{}, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	entries := m.data[collection]
	result := make([]interface{}, len(entries))
	for i, e := range entries {
		result[i] = e.doc
	}
	return result, nil
}

func (m *Memory) Close() error { return nil }
