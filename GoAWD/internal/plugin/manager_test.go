package plugin

import (
	"testing"

	"goawd/internal/types"
)

// stubPlugin registers hooks the way the bundled plugins do: every hook needs
// the caller to be able to raise alerts.
type stubPlugin struct {
	seen []interface{}
}

func (p *stubPlugin) Name() string { return "stub" }

func (p *stubPlugin) Register(m *Manager) {
	m.Register("Web", "processLog", func(caller interface{}, data interface{}) interface{} {
		p.seen = append(p.seen, caller)
		if web, ok := data.(*types.WebLogData); ok {
			web.Buffer = "modified:" + web.Buffer
		}
		return data
	})
	m.Register("Web", "processLog", func(caller interface{}, data interface{}) interface{} {
		p.seen = append(p.seen, caller)
		return data
	})
}

func TestInvokePassesCallerToEveryHook(t *testing.T) {
	m := NewManager()
	p := &stubPlugin{}
	m.RegisterPlugin(p)

	caller := &fakeReceiver{}
	web := &types.WebLogData{Buffer: "hello"}
	got := m.Invoke(caller, "Web", "processLog", web)

	// Regression guard: the caller used to be dropped, which silently disabled
	// alerts in every bundled plugin.
	if len(p.seen) != 2 {
		t.Fatalf("hooks invoked %d times, want 2", len(p.seen))
	}
	for i, c := range p.seen {
		if c != caller {
			t.Fatalf("hook %d received caller %v, want the invoker", i, c)
		}
	}

	if got.(*types.WebLogData).Buffer != "modified:hello" {
		t.Errorf("buffer = %q, want %q", got.(*types.WebLogData).Buffer, "modified:hello")
	}
}

func TestInvokeIsCaseInsensitiveAndReturnsDataWhenNoHooks(t *testing.T) {
	m := NewManager()
	data := &types.WebLogData{Buffer: "x"}
	if got := m.Invoke(&fakeReceiver{}, "Web", "processLog", data); got != data {
		t.Error("Invoke with no hooks must return the input unchanged")
	}

	m.Register("FileSystem", "processLog", func(_ interface{}, data interface{}) interface{} {
		return "replaced"
	})
	if got := m.Invoke(&fakeReceiver{}, "filesystem", "PROCESSLoG", data); got != "replaced" {
		t.Errorf("hook key should be case-insensitive, got %v", got)
	}
}

func TestReloadUsesLastLoadedDir(t *testing.T) {
	m := NewManager()
	// No .so files exist in the temp dir, but the directory must be remembered
	// so a later Reload() actually scans it instead of silently doing nothing.
	m.LoadPlugins(t.TempDir())
	m.mu.RLock()
	dir := m.pluginDir
	m.mu.RUnlock()
	if dir == "" {
		t.Fatal("LoadPlugins must record the plugin directory for Reload()")
	}
}

func TestNamesReturnsCopy(t *testing.T) {
	m := NewManager()
	m.RegisterPlugin(&stubPlugin{})
	names := m.Names()
	names[0] = "mutated"
	if m.Names()[0] != "stub" {
		t.Error("Names() must return a copy")
	}
}

type fakeReceiver struct{ alerts []string }

func (f *fakeReceiver) SetAlert(alertType, pluginName, message, refID string, refPage int) {
	f.alerts = append(f.alerts, pluginName)
}
