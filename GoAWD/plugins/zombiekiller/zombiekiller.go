package zombiekiller

import (
	"sync"
	"time"

	"goawd/internal/plugin"
	"goawd/internal/types"
)

type ZombieKiller struct {
	mu      sync.Mutex
	deletes map[string]time.Time
	window  time.Duration
}

func init() {
	plugin.Register(&ZombieKiller{
		deletes: make(map[string]time.Time),
		window:  5 * time.Second,
	})
}

func (z *ZombieKiller) Name() string {
	return "ZombieKiller"
}

func (z *ZombieKiller) Register(m *plugin.Manager) {
	m.Register("FileSystem", "processLog", z.processFileEvent)
}

func (z *ZombieKiller) processFileEvent(caller interface{}, data interface{}) interface{} {
	fe, ok := data.(*types.FileEventData)
	if !ok {
		return data
	}
	if fe.IsDir {
		return data
	}

	z.mu.Lock()
	defer z.mu.Unlock()

	now := time.Now()

	switch fe.Oper {
	case "DELETE", "DELETE_SELF", "MOVED_FROM":
		z.deletes[fe.Path] = now
	case "CREATE", "MODIFY", "MOVED_TO":
		if deleteTime, exists := z.deletes[fe.Path]; exists {
			if now.Sub(deleteTime) <= z.window {
				delete(z.deletes, fe.Path)
				alert(caller, "FileSystem", "ZombieKiller",
					"检测到不死马行为: "+fe.Path+" 在删除后被快速重建",
					fe.ID, 1)
			} else {
				delete(z.deletes, fe.Path)
			}
		}
	}

	if len(z.deletes) > 1000 {
		for path, t := range z.deletes {
			if now.Sub(t) > z.window {
				delete(z.deletes, path)
			}
		}
	}
	return data
}

// alert reports through the caller (the GoAWD receiver) if it supports it.
func alert(caller interface{}, alertType, pluginName, message, refID string, refPage int) {
	if setter, ok := caller.(interface {
		SetAlert(string, string, string, string, int)
	}); ok {
		setter.SetAlert(alertType, pluginName, message, refID, refPage)
	}
}
