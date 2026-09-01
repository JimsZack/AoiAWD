package api

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"goawd/internal/plugin"
	"goawd/internal/storage"
	"goawd/internal/types"
)

type ProcessProvider interface {
	CurrentProcessPIDs() []int
	CurrentProcessList() []*types.ProcessInfo
}

type V1 struct {
	store     storage.Storage
	pluginMgr *plugin.Manager
	procProv  ProcessProvider
	startTime time.Time
}

func NewV1(store storage.Storage, mgr *plugin.Manager, pp ProcessProvider, startTime time.Time) *V1 {
	return &V1{
		store:     store,
		pluginMgr: mgr,
		procProv:  pp,
		startTime: startTime,
	}
}

func (v *V1) RegisterRoutes(mux *http.ServeMux, prefix, token string) {
	wrap := func(h http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			t := r.URL.Query().Get("token")
			if t == "" {
				t = r.Header.Get("Token")
			}
			if token != "" && t != token {
				http.Error(w, `{"result":"0","message":"Forbidden"}`, http.StatusForbidden)
				return
			}
			h(w, r)
		}
	}

	mux.HandleFunc(prefix+"info", wrap(v.Info))
	mux.HandleFunc(prefix+"listweb", wrap(v.ListWeb))
	mux.HandleFunc(prefix+"webdetail", wrap(v.WebDetail))
	mux.HandleFunc(prefix+"downloadwebautoscript", wrap(v.DownloadWebAutoScript))
	mux.HandleFunc(prefix+"listpwn", wrap(v.ListPWN))
	mux.HandleFunc(prefix+"pwndetail", wrap(v.PWNDetail))
	mux.HandleFunc(prefix+"downloadpwn", wrap(v.DownloadPWN))
	mux.HandleFunc(prefix+"generatepwnbin", wrap(v.GeneratePWNBin))
	mux.HandleFunc(prefix+"listfilesystem", wrap(v.ListFilesystem))
	mux.HandleFunc(prefix+"downloadfile", wrap(v.DownloadFile))
	mux.HandleFunc(prefix+"listprocess", wrap(v.ListProcess))
	mux.HandleFunc(prefix+"listcurrentprocess", wrap(v.ListCurrentProcess))
	mux.HandleFunc(prefix+"currentprocess", wrap(v.CurrentProcess))
	mux.HandleFunc(prefix+"listalert", wrap(v.ListAlert))
	mux.HandleFunc(prefix+"listplugin", wrap(v.ListPlugin))
	mux.HandleFunc(prefix+"reloadplugin", wrap(v.ReloadPlugin))
}

func writeJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func okResp(data interface{}) types.APIResponse {
	return types.APIResponse{Result: "1", Message: "ok", Data: data}
}

func errResp(msg string) types.APIResponse {
	return types.APIResponse{Result: "0", Message: msg}
}

func toMap(v interface{}) map[string]interface{} {
	raw, err := json.Marshal(v)
	if err != nil {
		return nil
	}
	var m map[string]interface{}
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil
	}
	return m
}

func formatTime(t int64) string {
	return time.Unix(t, 0).Format("2006-01-02 15:04:05")
}

func parsePage(r *http.Request) (int, int) {
	q := r.URL.Query()
	page, _ := strconv.Atoi(q.Get("page"))
	count, _ := strconv.Atoi(q.Get("count"))
	if page == 0 {
		page = 1
	}
	if count == 0 {
		count = 20
	}
	return page, count
}

func (v *V1) Info(w http.ResponseWriter, r *http.Request) {
	count := v.store.Count(r.Context(), types.CollAlert)
	uptime := time.Since(v.startTime)
	writeJSON(w, types.InfoResponse{
		TimestampLastUpdate:  time.Now().Format("2006-01-02 15:04:05"),
		CountAlert:           int(count),
		TimestampRunningTime: formatDuration(uptime),
	})
}

func formatDuration(d time.Duration) string {
	d = d.Round(time.Second)
	h := int(d / time.Hour)
	d -= time.Duration(h) * time.Hour
	m := int(d / time.Minute)
	d -= time.Duration(m) * time.Minute
	s := int(d / time.Second)
	return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
}

func (v *V1) ListWeb(w http.ResponseWriter, r *http.Request) {
	page, count := parsePage(r)
	docs, total, _ := v.store.Paginate(r.Context(), types.CollWeb, page, count)
	lastPage := lastPageOf(total, count)

	items := make([]map[string]interface{}, 0, len(docs))
	for _, d := range docs {
		m := toMap(d)
		if m == nil {
			continue
		}
		items = append(items, map[string]interface{}{
			"id":     m["id"],
			"time":   formatTimeField(m["time"]),
			"method": m["method"],
			"uri":    m["uri"],
			"remote": m["remote"],
		})
	}
	writeJSON(w, types.PaginatedResponse{Page: page, LastPage: lastPage, Data: items})
}

func formatTimeField(v interface{}) string {
	switch t := v.(type) {
	case float64:
		return formatTime(int64(t))
	case int64:
		return formatTime(t)
	case int:
		return formatTime(int64(t))
	case string:
		return t
	default:
		return ""
	}
}

func lastPageOf(total int64, count int) int {
	if count <= 0 {
		count = 20
	}
	lp := int((total + int64(count) - 1) / int64(count))
	if lp == 0 {
		lp = 1
	}
	return lp
}

func (v *V1) WebDetail(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		writeJSON(w, errResp("id is required"))
		return
	}
	doc, err := v.store.Get(r.Context(), types.CollWeb, id)
	if err != nil || doc == nil {
		writeJSON(w, errResp("not found"))
		return
	}
	writeJSON(w, doc)
}

func (v *V1) DownloadWebAutoScript(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	doc, err := v.store.Get(r.Context(), types.CollWeb, id)
	if err != nil || doc == nil {
		writeJSON(w, errResp("not found"))
		return
	}
	m := toMap(doc)
	if m == nil {
		writeJSON(w, errResp("decode error"))
		return
	}

	method, _ := m["method"].(string)
	uri, _ := m["uri"].(string)
	remote, _ := m["remote"].(string)

	script := fmt.Sprintf("#!/bin/bash\n# Auto replay script for web log %s\n# Time: %s\n", id, formatTimeField(m["time"]))
	script += fmt.Sprintf("curl -X %s 'http://%s%s'\n", method, remote, uri)

	w.Header().Set("Content-Type", "text/x-shellscript")
	w.Header().Set("Content-Disposition", "attachment; filename=replay_"+id+".sh")
	w.Write([]byte(script))
}

func (v *V1) ListPWN(w http.ResponseWriter, r *http.Request) {
	page, count := parsePage(r)
	docs, total, _ := v.store.Paginate(r.Context(), types.CollPWN, page, count)
	lastPage := lastPageOf(total, count)

	items := make([]map[string]interface{}, 0, len(docs))
	for _, d := range docs {
		m := toMap(d)
		if m == nil {
			continue
		}
		items = append(items, map[string]interface{}{
			"id":    m["id"],
			"time":  formatTimeField(m["time"]),
			"bin":   m["bin"],
			"stdin": m["stdin"],
			"stdout": m["stdout"],
		})
	}
	writeJSON(w, types.PaginatedResponse{Page: page, LastPage: lastPage, Data: items})
}

func (v *V1) PWNDetail(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		writeJSON(w, errResp("id is required"))
		return
	}
	doc, err := v.store.Get(r.Context(), types.CollPWN, id)
	if err != nil || doc == nil {
		writeJSON(w, errResp("not found"))
		return
	}
	writeJSON(w, doc)
}

func (v *V1) DownloadPWN(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	dlType := r.URL.Query().Get("type")
	part := r.URL.Query().Get("part")

	doc, err := v.store.Get(r.Context(), types.CollPWN, id)
	if err != nil || doc == nil {
		writeJSON(w, errResp("not found"))
		return
	}
	m := toMap(doc)
	if m == nil {
		writeJSON(w, errResp("decode error"))
		return
	}

	switch dlType {
	case "maps":
		maps, _ := m["maps"].(string)
		data, err := base64.StdEncoding.DecodeString(maps)
		if err != nil {
			data = []byte(maps)
		}
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Write(data)
	case "stream":
		streamlog, _ := m["streamlog"].([]interface{})
		if part == "" || part == "all" {
			var buf []byte
			for _, s := range streamlog {
				if sm, ok := s.(map[string]interface{}); ok {
					if b, ok := sm["buffer"].(string); ok {
						if decoded, err := base64.StdEncoding.DecodeString(b); err == nil {
							buf = append(buf, decoded...)
						}
					}
				}
			}
			w.Header().Set("Content-Type", "application/octet-stream")
			w.Write(buf)
		} else {
			idx, _ := strconv.Atoi(part)
			if idx >= 0 && idx < len(streamlog) {
				if sm, ok := streamlog[idx].(map[string]interface{}); ok {
					if b, ok := sm["buffer"].(string); ok {
						if decoded, err := base64.StdEncoding.DecodeString(b); err == nil {
							w.Header().Set("Content-Type", "application/octet-stream")
							w.Write(decoded)
							return
						}
					}
				}
			}
			writeJSON(w, errResp("part not found"))
		}
	default:
		writeJSON(w, errResp("invalid type"))
	}
}

func (v *V1) GeneratePWNBin(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Binary string `json:"binary"`
		Host   string `json:"host"`
		Port   int    `json:"port"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, errResp("invalid request body"))
		return
	}
	if req.Binary == "" || req.Host == "" || req.Port == 0 {
		writeJSON(w, errResp("binary, host, port are required"))
		return
	}

	elfData, err := base64.StdEncoding.DecodeString(req.Binary)
	if err != nil {
		writeJSON(w, errResp("invalid base64 binary"))
		return
	}

	wrapper := fmt.Sprintf(`#!/bin/bash
# Guardian wrapper generated by GoAWD
# Original ELF size: %d bytes
# Server: %s:%d
exec goawd-guardian -binary "%s" -host "%s" -port %d
`, len(elfData), req.Host, req.Port, "/dev/stdin", req.Host, req.Port)

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", "attachment; filename=guardianed.sh")
	w.Write([]byte(wrapper))
}

func (v *V1) ListFilesystem(w http.ResponseWriter, r *http.Request) {
	page, count := parsePage(r)
	docs, total, _ := v.store.Paginate(r.Context(), types.CollFilesystem, page, count)
	lastPage := lastPageOf(total, count)

	items := make([]map[string]interface{}, 0, len(docs))
	for _, d := range docs {
		m := toMap(d)
		if m == nil {
			continue
		}
		items = append(items, map[string]interface{}{
			"id":      m["id"],
			"time":    formatTimeField(m["time"]),
			"path":    m["path"],
			"oper":    m["oper"],
			"isdir":   m["isdir"],
			"content": m["content"],
		})
	}
	writeJSON(w, types.PaginatedResponse{Page: page, LastPage: lastPage, Data: items})
}

func (v *V1) DownloadFile(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	doc, err := v.store.Get(r.Context(), types.CollFilesystem, id)
	if err != nil || doc == nil {
		writeJSON(w, errResp("not found"))
		return
	}
	m := toMap(doc)
	content, _ := m["content"].(string)
	data, err := base64.StdEncoding.DecodeString(content)
	if err != nil {
		data = []byte(content)
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(data)
}

func (v *V1) ListProcess(w http.ResponseWriter, r *http.Request) {
	page, count := parsePage(r)
	docs, total, _ := v.store.Paginate(r.Context(), types.CollProcess, page, count)
	lastPage := lastPageOf(total, count)

	items := make([]map[string]interface{}, 0, len(docs))
	for _, d := range docs {
		m := toMap(d)
		if m == nil {
			continue
		}
		items = append(items, m)
	}
	writeJSON(w, types.PaginatedResponse{Page: page, LastPage: lastPage, Data: items})
}

func (v *V1) ListCurrentProcess(w http.ResponseWriter, r *http.Request) {
	procs := v.procProv.CurrentProcessList()
	items := make([]map[string]interface{}, 0, len(procs))
	for _, p := range procs {
		items = append(items, map[string]interface{}{
			"id":    p.PID,
			"time":  p.Time.Format("2006-01-02 15:04:05"),
			"pid":   p.PID,
			"ppid":  p.PPID,
			"uid":   p.UID,
			"user":  p.Username,
			"bin":   p.Cmd,
			"arg":   p.Param,
		})
	}
	writeJSON(w, types.PaginatedResponse{Page: 1, LastPage: 1, Data: items})
}

func (v *V1) CurrentProcess(w http.ResponseWriter, r *http.Request) {
	pids := v.procProv.CurrentProcessPIDs()
	if pids == nil {
		pids = []int{}
	}
	writeJSON(w, pids)
}

func (v *V1) ListAlert(w http.ResponseWriter, r *http.Request) {
	page, count := parsePage(r)
	docs, total, _ := v.store.Paginate(r.Context(), types.CollAlert, page, count)
	lastPage := lastPageOf(total, count)

	items := make([]map[string]interface{}, 0, len(docs))
	for _, d := range docs {
		m := toMap(d)
		if m == nil {
			continue
		}
		items = append(items, m)
	}
	writeJSON(w, types.PaginatedResponse{Page: page, LastPage: lastPage, Data: items})
}

func (v *V1) ListPlugin(w http.ResponseWriter, r *http.Request) {
	names := v.pluginMgr.Names()
	if names == nil {
		names = []string{}
	}
	for i, n := range names {
		if !strings.HasSuffix(n, ".go") {
			names[i] = n + ".go"
		}
	}
	writeJSON(w, okResp(names))
}

func (v *V1) ReloadPlugin(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, okResp(nil))
}
