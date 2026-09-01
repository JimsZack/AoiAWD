package plugin

import (
	goplugin "plugin"
	"path/filepath"
	"strings"
	"sync"
)

type HookFunc func(data interface{}) interface{}

type Plugin interface {
	Name() string
	Register(m *Manager)
}

type Manager struct {
	mu      sync.RWMutex
	hooks   map[string][]HookFunc
	plugins []Plugin
	names   []string
}

func NewManager() *Manager {
	return &Manager{
		hooks: make(map[string][]HookFunc),
	}
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

func (m *Manager) Invoke(caller interface{}, routine, operation string, data interface{}) interface{} {
	m.mu.RLock()
	key := hookKey(routine, operation)
	hooks, ok := m.hooks[key]
	m.mu.RUnlock()
	if !ok || len(hooks) == 0 {
		return data
	}

	SetCaller(caller)
	defer SetCaller(nil)

	for _, hook := range hooks {
		data = hook(data)
	}
	return data
}

func (m *Manager) GetCaller() interface{} {
	return GetCaller()
}

var (
	currentCaller interface{}
	callerMu      sync.RWMutex
)

func SetCaller(c interface{}) {
	callerMu.Lock()
	defer callerMu.Unlock()
	currentCaller = c
}

func GetCaller() interface{} {
	callerMu.RLock()
	defer callerMu.RUnlock()
	return currentCaller
}

func (m *Manager) Names() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]string, len(m.names))
	copy(out, m.names)
	return out
}

func (m *Manager) LoadPlugins(dir string) []string {
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
