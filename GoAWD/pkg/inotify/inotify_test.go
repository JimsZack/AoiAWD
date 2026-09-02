package inotify

import (
	"syscall"
	"testing"
	"unsafe"
)

func TestOpName(t *testing.T) {
	tests := []struct {
		name     string
		mask     uint32
		expected string
	}{
		{"IN_CREATE", syscall.IN_CREATE, "CREATE"},
		{"IN_MODIFY", syscall.IN_MODIFY, "MODIFY"},
		{"IN_DELETE", syscall.IN_DELETE, "DELETE"},
		{"IN_DELETE_SELF", syscall.IN_DELETE_SELF, "DELETE_SELF"},
		{"IN_MOVED_FROM", syscall.IN_MOVED_FROM, "MOVED_FROM"},
		{"IN_MOVED_TO", syscall.IN_MOVED_TO, "MOVED_TO"},
		{"IN_MOVE_SELF", syscall.IN_MOVE_SELF, "MOVE_SELF"},
		{"IN_ATTRIB", syscall.IN_ATTRIB, "ATTRIB"},
		{"IN_CLOSE_WRITE", syscall.IN_CLOSE_WRITE, "CLOSE_WRITE"},
		{"unknown mask", 0x00000000, "UNKNOWN"},
		{"combined flags", syscall.IN_CREATE | syscall.IN_MODIFY, "CREATE"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := opName(tt.mask)
			if got != tt.expected {
				t.Errorf("opName(%d) = %q, want %q", tt.mask, got, tt.expected)
			}
		})
	}
}

func TestWatcherWd2path(t *testing.T) {
	w := &Watcher{wd2path: make(map[int32]string)}

	w.mu.Lock()
	w.wd2path[1] = "/tmp/test"
	w.wd2path[2] = "/var/log"
	w.mu.Unlock()

	w.mu.Lock()
	if w.wd2path[1] != "/tmp/test" {
		t.Errorf("wd2path[1] = %q, want %q", w.wd2path[1], "/tmp/test")
	}
	if w.wd2path[2] != "/var/log" {
		t.Errorf("wd2path[2] = %q, want %q", w.wd2path[2], "/var/log")
	}
	w.mu.Unlock()
}

func TestWatcherWd2pathDelete(t *testing.T) {
	w := &Watcher{wd2path: make(map[int32]string)}

	w.mu.Lock()
	w.wd2path[1] = "/tmp/test"
	delete(w.wd2path, 1)
	if _, ok := w.wd2path[1]; ok {
		t.Error("expected wd2path[1] to be deleted")
	}
	w.mu.Unlock()
}

func TestInotifyEventStructSize(t *testing.T) {
	// inotify_event should be 16 bytes
	var ev inotifyEvent
	size := unsafe.Sizeof(ev)
	if size != 16 {
		t.Errorf("inotifyEvent size = %d, want 16", size)
	}
}
