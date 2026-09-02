package plugin

import (
	"path/filepath"
	goplugin "plugin"
	"strings"
	"sync"
)

// HookFunc receives the caller that triggered the invocation (used to raise
// alerts) together with the payload, and returns the (possibly modified) data.
type HookFunc func(caller interface{}, data interface{}) interface{}

type Plugin interface {
	Name() string
	Register(m *Manager)
}

type Manager struct {
	mu        sync.RWMutex
	hooks     map[string][]HookFunc
	plugins   []Plugin
	names     []string
	pluginDir string
}

func NewManager() *Manager {
	return &Manager{
		hooks: make(map[string][]HookFunc),
	}
}

func (m *Manager) SetPluginDir(dir string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.pluginDir = dir
}

var (
	globalPlugins []Plugin
	globalMu      sync.Mutex
)

func Register(p Plugin) {
	globalMu.Lock()
	defer globalMu.Unlock()
	globalPlugins = append(globalPlugins, p)
}

func Registered() []Plugin {
	globalMu.Lock()
	defer globalMu.Unlock()
	out := make([]Plugin, len(globalPlugins))
	copy(out, globalPlugins)
	return out
}

func hookKey(routine, operation string) string {
	return strings.ToLower(routine + "/" + operation)
}

func (m *Manager) Register(routine, operation string, hook HookFunc) {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := hookKey(routine, operation)
	m.hooks[key] = append(m.hooks[key], hook)
}

func (m *Manager) RegisterPlugin(p Plugin) {
	m.mu.Lock()
	m.plugins = append(m.plugins, p)
	m.names = append(m.names, p.Name())
	m.mu.Unlock()
	p.Register(m)
}

// Invoke runs the registered hooks for routine/operation in order, passing the
// caller through to each hook so that plugins can report alerts back to it.
func (m *Manager) Invoke(caller interface{}, routine, operation string, data interface{}) interface{} {
	m.mu.RLock()
	key := hookKey(routine, operation)
	hooks := make([]HookFunc, len(m.hooks[key]))
	copy(hooks, m.hooks[key])
	m.mu.RUnlock()

	for _, hook := range hooks {
		data = hook(caller, data)
	}
	return data
}

func (m *Manager) Names() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]string, len(m.names))
	copy(out, m.names)
	return out
}

func (m *Manager) Reload() []string {
	m.mu.Lock()
	// Clear existing plugins and hooks
	m.hooks = make(map[string][]HookFunc)
	m.plugins = nil
	m.names = nil
	dir := m.pluginDir
	m.mu.Unlock()

	if dir == "" {
		return nil
	}
	return m.LoadPlugins(dir)
}

func (m *Manager) LoadPlugins(dir string) []string {
	m.SetPluginDir(dir)

	pattern := filepath.Join(dir, "*.so")
	matches, err := filepath.Glob(pattern)
	if err != nil || len(matches) == 0 {
		return nil
	}

	var loaded []string
	for _, path := range matches {
		p, err := goplugin.Open(path)
		if err != nil {
			continue
		}
		sym, err := p.Lookup("PluginInstance")
		if err != nil {
			continue
		}
		if inst, ok := sym.(Plugin); ok {
			m.RegisterPlugin(inst)
			loaded = append(loaded, inst.Name())
		}
	}
	return loaded
}
