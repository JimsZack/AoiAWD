package inotify

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"syscall"
	"time"
	"unsafe"

	"goawd/internal/types"
	"goawd/pkg/tcpclient"
)

type watchEntry struct {
	path string
}

type Watcher struct {
	mu      sync.Mutex
	wd2path map[int32]string
}

func Watch(ctx context.Context, dirs []string, sender *tcpclient.Sender) error {
	fd, err := syscall.InotifyInit()
	if err != nil {
		return fmt.Errorf("inotify_init: %w", err)
	}
	defer syscall.Close(fd)

	w := &Watcher{wd2path: make(map[int32]string)}

	mask := uint32(syscall.IN_CREATE | syscall.IN_MODIFY | syscall.IN_DELETE |
		syscall.IN_DELETE_SELF | syscall.IN_MOVE_SELF | syscall.IN_MOVED_FROM |
		syscall.IN_MOVED_TO | syscall.IN_ATTRIB | syscall.IN_CLOSE_WRITE)

	for _, dir := range dirs {
		if err := w.addRecursive(fd, dir, mask); err != nil {
			log.Printf("inotify watch %s: %v", dir, err)
		} else {
			log.Printf("inotify watching: %s", dir)
		}
	}

	buf := make([]byte, 4096)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		n, err := syscall.Read(fd, buf)
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			log.Printf("inotify read: %v", err)
			time.Sleep(100 * time.Millisecond)
			continue
		}
		w.parseEvents(buf[:n], sender)
	}
}

func (w *Watcher) addRecursive(fd int, root string, mask uint32) error {
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			wd, err := syscall.InotifyAddWatch(fd, path, mask)
			if err != nil {
				return err
			}
			w.mu.Lock()
			w.wd2path[int32(wd)] = path
			w.mu.Unlock()
		}
		return nil
	})
}

type inotifyEvent struct {
	Wd     int32
	Mask   uint32
	Cookie uint32
	Len    uint32
}

func (w *Watcher) parseEvents(buf []byte, sender *tcpclient.Sender) {
	offset := 0
	for offset+16 <= len(buf) {
		ev := (*inotifyEvent)(unsafe.Pointer(&buf[offset]))
		nameLen := int(ev.Len)
		name := ""
		if nameLen > 0 && offset+16+nameLen <= len(buf) {
			nameBytes := buf[offset+16 : offset+16+nameLen]
			for i, b := range nameBytes {
				if b == 0 {
					nameBytes = nameBytes[:i]
					break
				}
			}
			name = string(nameBytes)
		}

		w.mu.Lock()
		dir, ok := w.wd2path[ev.Wd]
		w.mu.Unlock()

		var fullPath string
		if ok {
			if name != "" {
				fullPath = filepath.Join(dir, name)
			} else {
				fullPath = dir
			}
		} else {
			fullPath = name
		}

		if fullPath != "" {
			handleEvent(fullPath, ev.Mask, sender)
		}

		offset += 16 + nameLen
	}
}

func handleEvent(path string, mask uint32, sender *tcpclient.Sender) {
	info, err := os.Stat(path)
	isDir := false
	size := int64(0)
	if err == nil {
		isDir = info.IsDir()
		size = info.Size()
	}

	oper := opName(mask)
	content := ""
	if !isDir && (mask&syscall.IN_CREATE != 0 || mask&syscall.IN_MODIFY != 0) {
		content = readFileHead(path, 50)
	}

	fe := types.FileEventData{
		Time:    time.Now().Unix(),
		Path:    path,
		IsDir:   isDir,
		Oper:    oper,
		Size:    size,
		Content: content,
	}

	msg := types.ProbeMessage{Type: types.MsgTypeNewFile, Data: fe}
	if err := sender.Send(msg); err != nil {
		log.Printf("send file event: %v", err)
	}
}

func opName(mask uint32) string {
	switch {
	case mask&syscall.IN_CREATE != 0:
		return "CREATE"
	case mask&syscall.IN_MODIFY != 0:
		return "MODIFY"
	case mask&syscall.IN_DELETE != 0:
		return "DELETE"
	case mask&syscall.IN_DELETE_SELF != 0:
		return "DELETE_SELF"
	case mask&syscall.IN_MOVED_FROM != 0:
		return "MOVED_FROM"
	case mask&syscall.IN_MOVED_TO != 0:
		return "MOVED_TO"
	case mask&syscall.IN_MOVE_SELF != 0:
		return "MOVE_SELF"
	case mask&syscall.IN_ATTRIB != 0:
		return "ATTRIB"
	case mask&syscall.IN_CLOSE_WRITE != 0:
		return "CLOSE_WRITE"
	default:
		return "UNKNOWN"
	}
}

func readFileHead(path string, n int) string {
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer f.Close()

	buf := make([]byte, n)
	read, _ := f.Read(buf)
	return base64.StdEncoding.EncodeToString(buf[:read])
}
