package kingwatcher

import (
	"testing"

	"goawd/internal/types"
)

func TestKingWatcherName(t *testing.T) {
	kw := &KingWatcher{targetFile: "/flag"}
	if kw.Name() != "KingWatcher" {
		t.Errorf("Name() = %q, want %q", kw.Name(), "KingWatcher")
	}
}

func TestProcessFileEventTargetFileModified(t *testing.T) {
	kw := &KingWatcher{targetFile: "/flag"}
	fe := &types.FileEventData{
		ID:   "test-id",
		Path: "/flag",
		Oper: "MODIFY",
	}

	result := kw.processFileEvent(nil, fe)
	if result != fe {
		t.Error("processFileEvent should return the same data")
	}
}

func TestProcessFileEventTargetFileCreated(t *testing.T) {
	kw := &KingWatcher{targetFile: "/flag"}
	fe := &types.FileEventData{
		ID:   "test-id",
		Path: "/flag",
		Oper: "CREATE",
	}

	kw.processFileEvent(nil, fe)
}

func TestProcessFileEventTargetFileDeleted(t *testing.T) {
	kw := &KingWatcher{targetFile: "/flag"}
	fe := &types.FileEventData{
		ID:   "test-id",
		Path: "/flag",
		Oper: "DELETE",
	}

	kw.processFileEvent(nil, fe)
}

func TestProcessFileEventUnrelatedFile(t *testing.T) {
	kw := &KingWatcher{targetFile: "/flag"}
	fe := &types.FileEventData{
		ID:   "test-id",
		Path: "/tmp/unrelated.txt",
		Oper: "MODIFY",
	}

	result := kw.processFileEvent(nil, fe)
	if result != fe {
		t.Error("processFileEvent should return unchanged data for unrelated file")
	}
}

func TestProcessFileEventWrongType(t *testing.T) {
	kw := &KingWatcher{targetFile: "/flag"}
	result := kw.processFileEvent(nil, "not a file event")
	if result != "not a file event" {
		t.Error("processFileEvent should return data unchanged for wrong type")
	}
}

func TestProcessFileEventAccess(t *testing.T) {
	kw := &KingWatcher{targetFile: "/flag"}
	fe := &types.FileEventData{
		ID:   "test-id",
		Path: "/flag",
		Oper: "ACCESS",
	}

	result := kw.processFileEvent(nil, fe)
	if result != fe {
		t.Error("ACCESS operation should not trigger alert")
	}
}

func TestProcessFileEventMovedFrom(t *testing.T) {
	kw := &KingWatcher{targetFile: "/flag"}
	fe := &types.FileEventData{
		ID:   "test-id",
		Path: "/flag",
		Oper: "MOVED_FROM",
	}

	kw.processFileEvent(nil, fe)
}

func TestProcessFileEventMovedTo(t *testing.T) {
	kw := &KingWatcher{targetFile: "/flag"}
	fe := &types.FileEventData{
		ID:   "test-id",
		Path: "/flag",
		Oper: "MOVED_TO",
	}

	kw.processFileEvent(nil, fe)
}

func TestProcessFileEventDeleteSelf(t *testing.T) {
	kw := &KingWatcher{targetFile: "/flag"}
	fe := &types.FileEventData{
		ID:   "test-id",
		Path: "/flag",
		Oper: "DELETE_SELF",
	}

	kw.processFileEvent(nil, fe)
}

func TestAlertWithCaller(t *testing.T) {
	type mockCaller struct {
		alertType string
		plugin    string
		message   string
	}
	// Should not panic
	alert(nil, "FileSystem", "KingWatcher", "test", "id", 1)
	alert(&mockCaller{}, "FileSystem", "KingWatcher", "test", "id", 1)
}

func TestDefaultTargetFile(t *testing.T) {
	// Test that init() sets default target
	kw := &KingWatcher{}
	if kw.targetFile != "" {
		t.Errorf("default targetFile = %q, want empty", kw.targetFile)
	}
}
