package types

import (
	"testing"
)

func TestGenID(t *testing.T) {
	id1 := GenID()
	id2 := GenID()

	if id1 == "" {
		t.Error("GenID returned empty string")
	}
	if id1 == id2 {
		t.Error("GenID returned duplicate IDs")
	}
	if len(id1) != 32 {
		t.Errorf("GenID length = %d, want 32 (16 bytes hex)", len(id1))
	}
}

func TestInotifyEventName(t *testing.T) {
	tests := []struct {
		mask uint32
		want string
	}{
		{0x00000001, "ACCESS"},
		{0x00000002, "MODIFY"},
		{0x00000100, "CREATE"},
		{0x00000200, "DELETE"},
		{0x00000000, "UNKNOWN"},
	}
	for _, tt := range tests {
		got := InotifyEventName(tt.mask)
		if got != tt.want {
			t.Errorf("InotifyEventName(0x%x) = %s, want %s", tt.mask, got, tt.want)
		}
	}
}

func TestNow(t *testing.T) {
	n := Now()
	if n <= 0 {
		t.Error("Now should return positive unix timestamp")
	}
}
