package types

import (
	"crypto/rand"
	"encoding/hex"
	"time"
)

func GenID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return hex.EncodeToString([]byte(time.Now().String()))
	}
	return hex.EncodeToString(b)
}

const (
	MsgTypeWeb            = "web"
	MsgTypeNewProcess     = "new_process"
	MsgTypeCurrentProcess = "pid_list"
	MsgTypeNewFile        = "file"
	MsgTypePWN            = "pwn"
	MsgTypePing           = "ping"
	MsgTypePong           = "pong"
)

type ProbeMessage struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

type FileUpload struct {
	Name     string `json:"name" bson:"name"`
	Type     string `json:"type" bson:"type"`
	Size     int64  `json:"size" bson:"size"`
	TempName string `json:"tmp_name" bson:"tmp_name"`
}

type WebLogData struct {
	ID     string `json:"id,omitempty" bson:"-"`
	Time   int64  `json:"time" bson:"time"`
	Script string `json:"script" bson:"script"`
	Method string `json:"method" bson:"method"`
	URI    string `json:"uri" bson:"uri"`
	Remote string `json:"remote" bson:"remote"`
	Header map[string]string `json:"header" bson:"header"`
	GET    map[string]string `json:"get" bson:"get"`
	POST   map[string]string `json:"post" bson:"post"`
	Cookie map[string]string `json:"cookie" bson:"cookie"`
	File   []FileUpload      `json:"file" bson:"file"`
	Buffer string            `json:"buffer" bson:"buffer"`
}

type PWNInitData struct {
	File string `json:"file" bson:"file"`
	Type string `json:"type" bson:"type"`
	PID  int    `json:"pid" bson:"pid"`
	Maps string `json:"maps" bson:"maps"`
}

type PwnSocket struct {
	ConnID uint64
	PID    int
	Type   string
}

type StreamStat struct {
	Group int `json:"group" bson:"group"`
	Byte  int `json:"byte" bson:"byte"`
}

type StreamLog struct {
	Type   string `json:"type" bson:"type"`
	Buffer string `json:"buffer" bson:"buffer"`
}

type PwnProcess struct {
	ID        string      `json:"id,omitempty" bson:"-"`
	Time      int64       `json:"time" bson:"time"`
	Bin       string      `json:"bin" bson:"bin"`
	Maps      string      `json:"maps" bson:"maps"`
	Stdin     StreamStat  `json:"stdin" bson:"stdin"`
	Stdout    StreamStat  `json:"stdout" bson:"stdout"`
	StreamLog []StreamLog `json:"streamlog" bson:"streamlog"`
}

type FileEventData struct {
	ID      string `json:"id,omitempty" bson:"-"`
	Time    int64  `json:"time" bson:"time"`
	Path    string `json:"path" bson:"path"`
	IsDir   bool   `json:"isdir" bson:"isdir"`
	Oper    string `json:"oper" bson:"oper"`
	Size    int64  `json:"size" bson:"size"`
	Content string `json:"content" bson:"content"`
}

var InotifyEventMap = map[uint32]string{
	0x00000001: "ACCESS",
	0x00000002: "MODIFY",
	0x00000004: "ATTRIB",
	0x00000008: "CLOSE_WRITE",
	0x00000010: "CLOSE_NOWRITE",
	0x00000018: "CLOSE",
	0x00000020: "OPEN",
	0x00000040: "MOVED_FROM",
	0x00000080: "MOVED_TO",
	0x000000C0: "MOVE",
	0x00000100: "CREATE",
	0x00000200: "DELETE",
	0x00000400: "DELETE_SELF",
	0x00000800: "MOVE_SELF",
}

func InotifyEventName(mask uint32) string {
	var parts []string
	for bit, name := range InotifyEventMap {
		if mask&bit == bit && bit != 0x00000018 && bit != 0x000000C0 {
			parts = append(parts, name)
		}
	}
	if len(parts) == 0 {
		return "UNKNOWN"
	}
	return parts[0]
}

type ProcessData struct {
	ID       string `json:"id,omitempty" bson:"-"`
	Time     int64  `json:"time" bson:"time"`
	PID      int    `json:"pid" bson:"pid"`
	PPID     int    `json:"ppid" bson:"ppid"`
	UID      int    `json:"uid" bson:"uid"`
	Username string `json:"user" bson:"user"`
	Cmd      string `json:"bin" bson:"bin"`
	Param    string `json:"arg" bson:"arg"`
}

type ProcessInfo struct {
	PID      int
	PPID     int
	UID      int
	Username string
	Cmd      string
	Param    string
	Time     time.Time
}

type AlertReference struct {
	Page int    `json:"page" bson:"page"`
	ID   string `json:"id" bson:"id"`
}

type AlertLogData struct {
	ID        string        `json:"id,omitempty" bson:"-"`
	Time      int64         `json:"time" bson:"time"`
	Type      string        `json:"type" bson:"type"`
	Plugin    string        `json:"plugin" bson:"plugin"`
	Message   string        `json:"message" bson:"message"`
	Reference AlertReference `json:"reference" bson:"reference"`
}

type APIResponse struct {
	Result  string      `json:"result,omitempty"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

type PaginatedResponse struct {
	Page     int         `json:"page"`
	LastPage int         `json:"last_page"`
	Data     interface{} `json:"data"`
}

type InfoResponse struct {
	TimestampLastUpdate  string `json:"timestamp_lastupdate"`
	CountAlert           int    `json:"count_alert"`
	TimestampRunningTime string `json:"timestamp_runningtime"`
}

type WSMessage struct {
	Operation string `json:"operation"`
	Type      string `json:"type"`
}

const (
	WSOpReload = "reload"
)

const (
	WSTypeWeb     = "web"
	WSTypePWN     = "pwn"
	WSTypeFile    = "file"
	WSTypeProcess = "process"
	WSTypeAlert   = "alert"
)

const (
	CollWeb        = "web"
	CollPWN        = "pwn"
	CollFilesystem = "filesystem"
	CollProcess    = "process"
	CollAlert      = "alert"
)

func Now() int64 {
	return time.Now().Unix()
}
