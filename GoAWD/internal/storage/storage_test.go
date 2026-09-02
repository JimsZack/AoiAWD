package storage

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

func TestFileCloseFlushesPendingWrites(t *testing.T) {
	path := filepath.Join(t.TempDir(), "goawd.json")
	f, err := NewFile(path)
	if err != nil {
		t.Fatalf("NewFile: %v", err)
	}

	ctx := context.Background()
	if err := f.Save(ctx, "web", "id1", map[string]interface{}{"uri": "/index.php"}); err != nil {
		t.Fatalf("Save: %v", err)
	}

	// Nothing is flushed until the 5s tick or Close, so Close must flush sync.
	if err := f.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if !json.Valid(raw) {
		t.Fatalf("flushed file is not valid JSON: %s", raw)
	}
	var snapshot map[string][]fileEntry
	if err := json.Unmarshal(raw, &snapshot); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if len(snapshot["web"]) != 1 {
		t.Fatalf("flushed %d web entries, want 1", len(snapshot["web"]))
	}
	if string(snapshot["web"][0].Data) == "" {
		t.Error("entry payload is empty")
	}
}

func TestFileCloseIsIdempotent(t *testing.T) {
	f, err := NewFile(filepath.Join(t.TempDir(), "goawd.json"))
	if err != nil {
		t.Fatalf("NewFile: %v", err)
	}
	f.Close()
	// Regression guard: closing twice used to panic on a closed channel.
	f.Close()
}

func TestMemoryEvictsOldestEntries(t *testing.T) {
	m := NewMemory()
	m.maxEntries = 100
	ctx := context.Background()

	// Eviction triggers once the collection is evictBatch entries over the cap,
	// then trims back down to exactly maxEntries.
	total := m.maxEntries + evictBatch
	for i := 0; i < total; i++ {
		m.Save(ctx, "web", "id-"+strconv.Itoa(i), map[string]interface{}{"i": i})
	}

	if got := m.Count(ctx, "web"); got != int64(m.maxEntries) {
		t.Fatalf("count = %d, want %d", got, m.maxEntries)
	}

	// The newest documents must survive, the oldest must be gone.
	if _, err := m.Get(ctx, "web", "id-"+strconv.Itoa(total-1)); err != nil {
		t.Fatalf("Get newest: %v", err)
	}
	got, _ := m.Get(ctx, "web", "id-0")
	if got != nil {
		t.Error("oldest entry should have been evicted")
	}
}

func TestFileEvictsOldestEntries(t *testing.T) {
	f, err := NewFile(filepath.Join(t.TempDir(), "goawd.json"))
	if err != nil {
		t.Fatalf("NewFile: %v", err)
	}
	defer f.Close()

	f.maxEntries = 100
	ctx := context.Background()

	// Eviction triggers once the collection is evictBatch entries over the cap,
	// then trims back down to exactly maxEntries.
	total := f.maxEntries + evictBatch
	for i := 0; i < total; i++ {
		f.Save(ctx, "web", "id-"+strconv.Itoa(i), map[string]interface{}{"i": i})
	}

	if got := f.Count(ctx, "web"); got != int64(f.maxEntries) {
		t.Fatalf("count = %d, want %d", got, f.maxEntries)
	}
	if _, err := f.Get(ctx, "web", "id-"+strconv.Itoa(total-1)); err != nil {
		t.Fatalf("Get newest: %v", err)
	}
	got, _ := f.Get(ctx, "web", "id-0")
	if got != nil {
		t.Error("oldest entry should have been evicted")
	}
}

func TestMemorySaveAndGet(t *testing.T) {
	m := NewMemory()
	ctx := context.Background()

	doc := map[string]interface{}{"name": "test", "value": 42}
	if err := m.Save(ctx, "web", "id1", doc); err != nil {
		t.Fatalf("Save: %v", err)
	}

	got, err := m.Get(ctx, "web", "id1")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	m2, ok := got.(map[string]interface{})
	if !ok {
		t.Fatalf("unexpected type: %T", got)
	}
	if m2["name"] != "test" {
		t.Errorf("name = %v, want test", m2["name"])
	}
}

func TestMemoryPaginate(t *testing.T) {
	m := NewMemory()
	ctx := context.Background()

	for i := 0; i < 25; i++ {
		doc := map[string]interface{}{"i": i}
		m.Save(ctx, "web", "id-"+string(rune('a'+i)), doc)
	}

	docs, total, err := m.Paginate(ctx, "web", 1, 10)
	if err != nil {
		t.Fatalf("Paginate: %v", err)
	}
	if total != 25 {
		t.Errorf("total = %d, want 25", total)
	}
	if len(docs) != 10 {
		t.Errorf("len(docs) = %d, want 10", len(docs))
	}

	docs2, _, _ := m.Paginate(ctx, "web", 3, 10)
	if len(docs2) != 5 {
		t.Errorf("page 3 len = %d, want 5", len(docs2))
	}
}

func TestMemoryCount(t *testing.T) {
	m := NewMemory()
	ctx := context.Background()

	if m.Count(ctx, "web") != 0 {
		t.Error("initial count should be 0")
	}

	m.Save(ctx, "web", "a", map[string]interface{}{})
	m.Save(ctx, "web", "b", map[string]interface{}{})
	m.Save(ctx, "pwn", "c", map[string]interface{}{})

	if m.Count(ctx, "web") != 2 {
		t.Error("web count should be 2")
	}
	if m.Count(ctx, "pwn") != 1 {
		t.Error("pwn count should be 1")
	}
}

func TestMemoryOverwrite(t *testing.T) {
	m := NewMemory()
	ctx := context.Background()

	m.Save(ctx, "web", "id1", map[string]interface{}{"v": 1})
	m.Save(ctx, "web", "id1", map[string]interface{}{"v": 2})

	if m.Count(ctx, "web") != 1 {
		t.Error("overwrite should not increase count")
	}
	got, _ := m.Get(ctx, "web", "id1")
	if got.(map[string]interface{})["v"].(int) != 2 {
		t.Error("should return updated value")
	}
}

func TestNormalizePage(t *testing.T) {
	tests := []struct {
		page, count, total int
		wantPage           int
		wantLast           int
	}{
		{1, 10, 25, 1, 3},
		{0, 10, 25, 3, 3},
		{5, 10, 25, 3, 3},
		{1, 20, 0, 1, 1},
	}
	for _, tt := range tests {
		page, _, last := NormalizePage(tt.page, tt.count, tt.total)
		if page != tt.wantPage {
			t.Errorf("page = %d, want %d", page, tt.wantPage)
		}
		if last != tt.wantLast {
			t.Errorf("last = %d, want %d", last, tt.wantLast)
		}
	}
}
