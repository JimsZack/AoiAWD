package storage

import (
	"context"
	"testing"
)

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
