package core

import (
	"bufio"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/url"
	"sync"
	"sync/atomic"
	"time"

	"goawd/internal/plugin"
	"goawd/internal/storage"
	"goawd/internal/types"
)

// maxPWNStreamLog caps how many chunks a single PWN session keeps, so a
// long-running binary cannot exhaust the server's memory.
const maxPWNStreamLog = 4096

type AlertSetter interface {
	SetAlert(alertType, pluginName, message, refID string, refPage int)
}

type pwnProcEntry struct {
	mu   sync.Mutex
	proc *types.PwnProcess
}

type Receiver struct {
	addr      string
	storage   storage.Storage
	pluginMgr *plugin.Manager
	hub       *Hub

	pwnSockets sync.Map
	pwnProcess sync.Map

	currentProcs sync.Map

	connCounter uint64
	listener    net.Listener
}

func NewReceiver(addr string, store storage.Storage, hub *Hub, mgr *plugin.Manager) *Receiver {
	return &Receiver{
		addr:      addr,
		storage:   store,
		pluginMgr: mgr,
		hub:       hub,
	}
}

func (r *Receiver) CurrentProcessPIDs() []int {
	var pids []int
	r.currentProcs.Range(func(k, _ interface{}) bool {
		pids = append(pids, k.(int))
		return true
	})
	return pids
}

func (r *Receiver) CurrentProcessList() []*types.ProcessInfo {
	var list []*types.ProcessInfo
	r.currentProcs.Range(func(_, v interface{}) bool {
		list = append(list, v.(*types.ProcessInfo))
		return true
	})
	return list
}

func (r *Receiver) Start(ctx context.Context) error {
	var err error
	r.listener, err = net.Listen("tcp", r.addr)
	if err != nil {
		return fmt.Errorf("tcp listen %s: %w", r.addr, err)
	}
	log.Printf("TCP Receiver listening on %s", r.addr)

	go func() {
		<-ctx.Done()
		r.listener.Close()
	}()

	for {
		conn, err := r.listener.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				log.Printf("accept error: %v", err)
				continue
			}
		}
		connID := atomic.AddUint64(&r.connCounter, 1)
		go r.handleConn(ctx, conn, connID)
	}
}

func (r *Receiver) handleConn(ctx context.Context, conn net.Conn, connID uint64) {
	defer func() {
		if rec := recover(); rec != nil {
			log.Printf("panic in handleConn: %v", rec)
		}
		conn.Close()
		r.handlePWNClose(connID)
	}()

	scanner := bufio.NewScanner(conn)
	scanner.Buffer(make([]byte, 0, 4*1024*1024), 4*1024*1024)

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return
		default:
		}
		data := scanner.Bytes()
		if len(data) == 0 {
			continue
		}
		if _, ok := r.pwnSockets.Load(connID); ok {
			r.processPWNStream(data, connID)
			continue
		}
		r.processMessage(data, conn, connID)
	}
}

func (r *Receiver) processMessage(data []byte, conn net.Conn, connID uint64) {
	var msg types.ProbeMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		log.Printf("json decode error: %v (data=%q)", err, string(data[:min(len(data), 128)]))
		return
	}

	switch msg.Type {
	case types.MsgTypeWeb:
		r.handleWeb(msg.Data, conn)
	case types.MsgTypeNewFile:
		r.handleNewFile(msg.Data)
	case types.MsgTypeNewProcess:
		r.handleNewProcess(msg.Data)
	case types.MsgTypeCurrentProcess:
		r.handleProcessList(msg.Data)
	case types.MsgTypePWN:
		r.handleNewPWN(msg.Data, connID)
	case types.MsgTypePing:
		r.handlePing(conn)
	default:
		log.Printf("unknown message type: %s", msg.Type)
	}
}

func (r *Receiver) handlePing(conn net.Conn) {
	resp := types.ProbeMessage{Type: types.MsgTypePong, Data: []interface{}{}}
	data, err := json.Marshal(resp)
	if err != nil {
		return
	}
	if _, err := conn.Write(append(data, '\n')); err != nil {
		log.Printf("write pong: %v", err)
	}
}

func (r *Receiver) handleWeb(data interface{}, conn net.Conn) {
	raw, err := json.Marshal(data)
	if err != nil {
		return
	}
	var web types.WebLogData
	if err := json.Unmarshal(raw, &web); err != nil {
		log.Printf("web data decode error: %v", err)
		return
	}

	web.ID = types.GenID()
	web.Time = types.Now()
	r.decodeWebData(&web)

	processed := r.pluginMgr.Invoke(r, "Web", "processLog", &web)
	if pwd, ok := processed.(*types.WebLogData); ok {
		web = *pwd
	}

	resp := base64.StdEncoding.EncodeToString([]byte(web.Buffer))
	conn.Write([]byte(resp + "\n"))

	r.storage.Save(context.Background(), types.CollWeb, web.ID, web)
	r.hub.Notify(types.WSTypeWeb)
}

func (r *Receiver) decodeWebData(w *types.WebLogData) {
	w.Header = decodeMap(w.Header)
	w.GET = decodeMap(w.GET)
	w.POST = decodeMap(w.POST)
	w.Cookie = decodeMap(w.Cookie)
}

func decodeMap(m map[string]string) map[string]string {
	if m == nil {
		return map[string]string{}
	}
	out := make(map[string]string, len(m))
	for k, v := range m {
		if decoded, err := url.QueryUnescape(v); err == nil {
			out[k] = decoded
		} else {
			out[k] = v
		}
	}
	return out
}

func (r *Receiver) handleNewFile(data interface{}) {
	raw, err := json.Marshal(data)
	if err != nil {
		return
	}
	var fe types.FileEventData
	if err := json.Unmarshal(raw, &fe); err != nil {
		log.Printf("file data decode error: %v", err)
		return
	}

	fe.ID = types.GenID()
	if fe.Time == 0 {
		fe.Time = types.Now()
	}
	if fe.Oper == "" {
		fe.Oper = "UNKNOWN"
	}

	r.pluginMgr.Invoke(r, "FileSystem", "processLog", &fe)

	r.storage.Save(context.Background(), types.CollFilesystem, fe.ID, fe)
	r.hub.Notify(types.WSTypeFile)
}

func (r *Receiver) handleNewProcess(data interface{}) {
	raw, err := json.Marshal(data)
	if err != nil {
		return
	}
	var pd types.ProcessData
	if err := json.Unmarshal(raw, &pd); err != nil {
		log.Printf("process data decode error: %v", err)
		return
	}

	pd.ID = types.GenID()
	if pd.Time == 0 {
		pd.Time = types.Now()
	}

	r.pluginMgr.Invoke(r, "Process", "processLog", &pd)

	r.storage.Save(context.Background(), types.CollProcess, pd.ID, pd)
	r.currentProcs.Store(pd.PID, &types.ProcessInfo{
		PID:      pd.PID,
		PPID:     pd.PPID,
		UID:      pd.UID,
		Username: pd.Username,
		Cmd:      pd.Cmd,
		Param:    pd.Param,
		Time:     time.Unix(pd.Time, 0),
	})
	r.hub.Notify(types.WSTypeProcess)
}

func (r *Receiver) handleProcessList(data interface{}) {
	raw, err := json.Marshal(data)
	if err != nil {
		return
	}
	var pids []int
	if err := json.Unmarshal(raw, &pids); err != nil {
		log.Printf("pid list decode error: %v", err)
		return
	}

	seen := make(map[int]bool, len(pids))
	for _, pid := range pids {
		seen[pid] = true
	}

	r.currentProcs.Range(func(k, _ interface{}) bool {
		pid := k.(int)
		if !seen[pid] {
			r.currentProcs.Delete(pid)
		}
		return true
	})
}

func (r *Receiver) handleNewPWN(data interface{}, connID uint64) {
	raw, err := json.Marshal(data)
	if err != nil {
		return
	}
	var init types.PWNInitData
	if err := json.Unmarshal(raw, &init); err != nil {
		log.Printf("pwn init decode error: %v", err)
		return
	}

	r.pwnSockets.Store(connID, &types.PwnSocket{
		ConnID: connID,
		PID:    init.PID,
		Type:   init.Type,
	})

	if _, exists := r.pwnProcess.Load(init.PID); !exists {
		r.pwnProcess.Store(init.PID, &pwnProcEntry{
			proc: &types.PwnProcess{
				ID:        types.GenID(),
				Time:      types.Now(),
				Bin:       init.File,
				Maps:      init.Maps,
				StreamLog: []types.StreamLog{},
			},
		})
	}
}

func (r *Receiver) processPWNStream(data []byte, connID uint64) {
	val, ok := r.pwnSockets.Load(connID)
	if !ok {
		return
	}
	socket := val.(*types.PwnSocket)

	pval, ok := r.pwnProcess.Load(socket.PID)
	if !ok {
		return
	}
	entry := pval.(*pwnProcEntry)

	chunk := string(data)
	entry.mu.Lock()
	if len(entry.proc.StreamLog) < maxPWNStreamLog {
		entry.proc.StreamLog = append(entry.proc.StreamLog, types.StreamLog{
			Type:   socket.Type,
			Buffer: chunk,
		})
	} else {
		entry.proc.Truncated = true
	}
	if socket.Type == "stdin" {
		entry.proc.Stdin.Group++
		entry.proc.Stdin.Byte += len(chunk)
	} else {
		entry.proc.Stdout.Group++
		entry.proc.Stdout.Byte += len(chunk)
	}
	entry.mu.Unlock()
}

func (r *Receiver) handlePWNClose(connID uint64) {
	val, ok := r.pwnSockets.Load(connID)
	if !ok {
		return
	}
	socket := val.(*types.PwnSocket)
	r.pwnSockets.Delete(connID)

	pval, ok := r.pwnProcess.Load(socket.PID)
	if !ok {
		return
	}
	entry := pval.(*pwnProcEntry)

	remaining := 0
	r.pwnSockets.Range(func(_, v interface{}) bool {
		if v.(*types.PwnSocket).PID == socket.PID {
			remaining++
		}
		return true
	})
	if remaining > 0 {
		return
	}

	r.pwnProcess.Delete(socket.PID)

	entry.mu.Lock()
	proc := entry.proc
	for i := range proc.StreamLog {
		proc.StreamLog[i].Buffer = base64.StdEncoding.EncodeToString([]byte(proc.StreamLog[i].Buffer))
	}
	entry.mu.Unlock()

	r.pluginMgr.Invoke(r, "PWN", "processLog", proc)

	r.storage.Save(context.Background(), types.CollPWN, proc.ID, proc)
	r.hub.Notify(types.WSTypePWN)
}

func (r *Receiver) SetAlert(alertType, pluginName, message, refID string, refPage int) {
	alert := types.AlertLogData{
		ID:      types.GenID(),
		Time:    types.Now(),
		Type:    alertType,
		Plugin:  pluginName,
		Message: message,
		Reference: types.AlertReference{
			Page: refPage,
			ID:   refID,
		},
	}
	r.storage.Save(context.Background(), types.CollAlert, alert.ID, alert)
	r.hub.Notify(types.WSTypeAlert)
}
