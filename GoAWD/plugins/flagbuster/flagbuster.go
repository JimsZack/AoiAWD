package flagbuster

import (
	"crypto/rand"
	"encoding/hex"
	"regexp"

	"goawd/internal/plugin"
	"goawd/internal/types"
)

type FlagBuster struct{}

func init() {
	plugin.Register(&FlagBuster{})
}

func (f *FlagBuster) Name() string {
	return "FlagBuster"
}

func (f *FlagBuster) Register(m *plugin.Manager) {
	m.Register("Web", "processLog", f.processWebBuffer)
	m.Register("PWN", "processLog", f.processPWNBuffer)
}

var (
	flagRegex1 = regexp.MustCompile(`flag\{[^}]+\}`)
	flagRegex2 = regexp.MustCompile(`\{"flag":"[^"]*"\}`)
)

func fakeFlag() string {
	b := make([]byte, 16)
	rand.Read(b)
	return "flag{" + hex.EncodeToString(b) + "}"
}

func (f *FlagBuster) processWebBuffer(caller interface{}, data interface{}) interface{} {
	web, ok := data.(*types.WebLogData)
	if !ok {
		return data
	}

	original := web.Buffer
	modified := flagRegex1.ReplaceAllStringFunc(original, func(_ string) string {
		return fakeFlag()
	})
	modified = flagRegex2.ReplaceAllStringFunc(modified, func(_ string) string {
		return `{"flag":"` + fakeFlag() + `"}`
	})

	if modified != original {
		web.Buffer = modified
		alert(caller, "Web", "FlagBuster", "发现Web应答包含flag，已替换为随机flag", web.ID, 1)
	}
	return data
}

func (f *FlagBuster) processPWNBuffer(caller interface{}, data interface{}) interface{} {
	proc, ok := data.(*types.PwnProcess)
	if !ok {
		return data
	}

	for i, sl := range proc.StreamLog {
		if sl.Type != "stdout" {
			continue
		}
		original := sl.Buffer
		modified := flagRegex1.ReplaceAllStringFunc(original, func(_ string) string {
			return fakeFlag()
		})
		if modified != original {
			proc.StreamLog[i].Buffer = modified
			alert(caller, "PWN", "FlagBuster", "发现PWN输出包含flag，已替换为随机flag", proc.ID, 1)
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
