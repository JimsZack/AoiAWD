package kingwatcher

import (
	"os"
	"strings"

	"goawd/internal/plugin"
	"goawd/internal/types"
)

type KingWatcher struct {
	targetFile string
}

func init() {
	target := os.Getenv("GOAWD_KING_FILE")
	if target == "" {
		target = "/flag"
	}
	plugin.Register(&KingWatcher{targetFile: target})
}

func (k *KingWatcher) Name() string {
	return "KingWatcher"
}

func (k *KingWatcher) Register(m *plugin.Manager) {
	m.Register("FileSystem", "processLog", k.processFileEvent)
}

func (k *KingWatcher) processFileEvent(caller interface{}, data interface{}) interface{} {
	fe, ok := data.(*types.FileEventData)
	if !ok {
		return data
	}

	if !strings.Contains(fe.Path, k.targetFile) {
		return data
	}

	switch fe.Oper {
	case "MODIFY", "CREATE", "DELETE", "DELETE_SELF", "MOVED_FROM", "MOVED_TO":
		alert(caller, "FileSystem", "KingWatcher",
			"检测到赛点文件被修改: "+fe.Path+" (操作: "+fe.Oper+")",
			fe.ID, 1)
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
