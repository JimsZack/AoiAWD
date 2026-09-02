package zombiekiller

import (
	"testing"
	"time"

	"goawd/internal/types"
)

func TestZombieKillerName(t *testing.T) {
	zk := &ZombieKiller{
		deletes: make(map[string]time.Time),
		window:  5 * time.Second,
	}
	if zk.Name() != "ZombieKiller" {
		t.Errorf("Name() = %q, want %q", zk.Name(), "ZombieKiller")
	}
}

func TestProcessFileEventWrongType(t *testing.T) {
	zk := &ZombieKiller{
		deletes: make(map[string]time.Time),
		window:  5 * time.Second,
	}
	result := zk.processFileEvent(nil, "not a file event")
	if result != "not a file event" {
		t.Error("processFileEvent should return data unchanged for wrong type")
	}
}

func TestProcessFileEventDirectory(t *testing.T) {
	zk := &ZombieKiller{
		deletes: make(map[string]time.Time),
		window:  5 * time.Second,
	}
	fe := &types.FileEventData{
		ID:    "test-id",
		Path:  "/tmp/test",
		IsDir: true,
		Oper:  "DELETE",
	}
	result := zk.processFileEvent(nil, fe)
	if result != fe {
		t.Error("processFileEvent should return unchanged data for directory")
	}
}

func TestProcessFileEventDelete(t *testing.T) {
	zk := &ZombieKiller{
		deletes: make(map[string]time.Time),
		window:  5 * time.Second,
	}
	fe := &types.FileEventData{
		ID:   "test-id",
		Path: "/tmp/backdoor.php",
		Oper: "DELETE",
	}
	zk.processFileEvent(nil, fe)
	if _, exists := zk.deletes["/tmp/backdoor.php"]; !exists {
		t.Error("DELETE should record path in deletes map")
	}
}

func TestProcessFileEventCreateAfterDelete(t *testing.T) {
	zk := &ZombieKiller{
		deletes: make(map[string]time.Time),
		window:  5 * time.Second,
	}
	// First delete
	deleteFe := &types.FileEventData{
		ID:   "test-id-1",
		Path: "/tmp/backdoor.php",
		Oper: "DELETE",
	}
	zk.processFileEvent(nil, deleteFe)

	// Then create within window
	createFe := &types.FileEventData{
		ID:   "test-id-2",
		Path: "/tmp/backdoor.php",
		Oper: "CREATE",
	}
	zk.processFileEvent(nil, createFe)

	// Path should be removed from deletes
	if _, exists := zk.deletes["/tmp/backdoor.php"]; exists {
		t.Error("CREATE after DELETE should remove path from deletes")
	}
}

func TestProcessFileEventCreateAfterDeleteOutsideWindow(t *testing.T) {
	zk := &ZombieKiller{
		deletes: make(map[string]time.Time),
		window:  1 * time.Millisecond,
	}
	// First delete
	deleteFe := &types.FileEventData{
		ID:   "test-id-1",
		Path: "/tmp/backdoor.php",
		Oper: "DELETE",
	}
	zk.processFileEvent(nil, deleteFe)

	// Wait outside window
	time.Sleep(10 * time.Millisecond)

	// Then create outside window
	createFe := &types.FileEventData{
		ID:   "test-id-2",
		Path: "/tmp/backdoor.php",
		Oper: "CREATE",
	}
	zk.processFileEvent(nil, createFe)

	// Path should be removed from deletes
	if _, exists := zk.deletes["/tmp/backdoor.php"]; exists {
		t.Error("CREATE after DELETE outside window should remove path from deletes")
	}
}

func TestProcessFileEventModifyAfterDelete(t *testing.T) {
	zk := &ZombieKiller{
		deletes: make(map[string]time.Time),
		window:  5 * time.Second,
	}
	// First delete
	deleteFe := &types.FileEventData{
		ID:   "test-id-1",
		Path: "/tmp/backdoor.php",
		Oper: "DELETE",
	}
	zk.processFileEvent(nil, deleteFe)

	// Then modify within window
	modifyFe := &types.FileEventData{
		ID:   "test-id-2",
		Path: "/tmp/backdoor.php",
		Oper: "MODIFY",
	}
	zk.processFileEvent(nil, modifyFe)
}

func TestProcessFileEventMovedFrom(t *testing.T) {
	zk := &ZombieKiller{
		deletes: make(map[string]time.Time),
		window:  5 * time.Second,
	}
	fe := &types.FileEventData{
		ID:   "test-id",
		Path: "/tmp/backdoor.php",
		Oper: "MOVED_FROM",
	}
	zk.processFileEvent(nil, fe)
	if _, exists := zk.deletes["/tmp/backdoor.php"]; !exists {
		t.Error("MOVED_FROM should record path in deletes map")
	}
}

func TestProcessFileEventMovedTo(t *testing.T) {
	zk := &ZombieKiller{
		deletes: make(map[string]time.Time),
		window:  5 * time.Second,
	}
	// First delete
	deleteFe := &types.FileEventData{
		ID:   "test-id-1",
		Path: "/tmp/backdoor.php",
		Oper: "DELETE",
	}
	zk.processFileEvent(nil, deleteFe)

	// Then move to within window
	moveFe := &types.FileEventData{
		ID:   "test-id-2",
		Path: "/tmp/backdoor.php",
		Oper: "MOVED_TO",
	}
	zk.processFileEvent(nil, moveFe)
}

func TestProcessFileEventDeleteSelf(t *testing.T) {
	zk := &ZombieKiller{
		deletes: make(map[string]time.Time),
		window:  5 * time.Second,
	}
	fe := &types.FileEventData{
		ID:   "test-id",
		Path: "/tmp/backdoor.php",
		Oper: "DELETE_SELF",
	}
	zk.processFileEvent(nil, fe)
	if _, exists := zk.deletes["/tmp/backdoor.php"]; !exists {
		t.Error("DELETE_SELF should record path in deletes map")
	}
}

func TestProcessFileEventCleanup(t *testing.T) {
	zk := &ZombieKiller{
		deletes: make(map[string]time.Time),
		window:  1 * time.Millisecond,
	}

	// Add many entries
	for i := 0; i < 1001; i++ {
		zk.deletes["/tmp/file"+string(rune(i))] = time.Now().Add(-10 * time.Millisecond)
	}

	// Trigger cleanup
	fe := &types.FileEventData{
		ID:   "test-id",
		Path: "/tmp/trigger",
		Oper: "MODIFY",
	}
	zk.processFileEvent(nil, fe)

	// Old entries should be cleaned up
	if len(zk.deletes) > 1000 {
		t.Errorf("cleanup should reduce deletes size, got %d", len(zk.deletes))
	}
}

func TestProcessFileEventAccess(t *testing.T) {
	zk := &ZombieKiller{
		deletes: make(map[string]time.Time),
		window:  5 * time.Second,
	}
	fe := &types.FileEventData{
		ID:   "test-id",
		Path: "/tmp/test.php",
		Oper: "ACCESS",
	}
	result := zk.processFileEvent(nil, fe)
	if result != fe {
		t.Error("ACCESS should not modify data")
	}
}

func TestAlertWithCaller(t *testing.T) {
	type mockCaller struct {
		alertType string
		plugin    string
		message   string
	}
	// Should not panic
	alert(nil, "FileSystem", "ZombieKiller", "test", "id", 1)
	alert(&mockCaller{}, "FileSystem", "ZombieKiller", "test", "id", 1)
}
