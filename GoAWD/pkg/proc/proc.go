package proc

import (
	"bufio"
	"context"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"goawd/internal/types"
	"goawd/pkg/tcpclient"
)

type Scanner struct {
	mu        sync.Mutex
	knownPIDs map[int]bool
}

func NewScanner() *Scanner {
	return &Scanner{knownPIDs: make(map[int]bool)}
}

func (s *Scanner) Scan(ctx context.Context, interval time.Duration, sender *tcpclient.Sender) error {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	s.scanOnce(sender)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			s.scanOnce(sender)
		}
	}
}

func (s *Scanner) scanOnce(sender *tcpclient.Sender) {
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return
	}

	currentPIDs := make(map[int]bool)
	var newProcs []*types.ProcessData

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		pid, err := strconv.Atoi(entry.Name())
		if err != nil {
			continue
		}
		currentPIDs[pid] = true

		s.mu.Lock()
		known := s.knownPIDs[pid]
		s.mu.Unlock()
		if known {
			continue
		}

		info := readProcessInfo(pid)
		if info == nil {
			continue
		}
		info.Time = time.Now().Unix()
		newProcs = append(newProcs, info)
	}

	s.mu.Lock()
	removed := []int{}
	for pid := range s.knownPIDs {
		if !currentPIDs[pid] {
			removed = append(removed, pid)
		}
	}
	for _, pid := range removed {
		delete(s.knownPIDs, pid)
	}
	for pid := range currentPIDs {
		s.knownPIDs[pid] = true
	}
	s.mu.Unlock()

	for _, pd := range newProcs {
		msg := types.ProbeMessage{Type: types.MsgTypeNewProcess, Data: pd}
		if err := sender.Send(msg); err != nil {
			log.Printf("send process event: %v", err)
		}
	}

	pidList := make([]int, 0, len(currentPIDs))
	for pid := range currentPIDs {
		pidList = append(pidList, pid)
	}
	msg := types.ProbeMessage{Type: types.MsgTypeCurrentProcess, Data: pidList}
	if err := sender.Send(msg); err != nil {
		log.Printf("send pid list: %v", err)
	}
}

func readProcessInfo(pid int) *types.ProcessData {
	base := filepath.Join("/proc", strconv.Itoa(pid))

	statData, err := os.ReadFile(filepath.Join(base, "stat"))
	if err != nil {
		return nil
	}

	fields := strings.Fields(string(statData))
	if len(fields) < 4 {
		return nil
	}

	ppid, _ := strconv.Atoi(fields[3])

	statusData, err := os.ReadFile(filepath.Join(base, "status"))
	uid := 0
	username := ""
	if err == nil {
		uid, username = parseStatus(statusData)
	}

	cmdlineData, _ := os.ReadFile(filepath.Join(base, "cmdline"))
	cmd, param := parseCmdline(cmdlineData)

	if cmd == "" {
		cmd = strings.Trim(fields[1], "()")
	}

	return &types.ProcessData{
		PID:      pid,
		PPID:     ppid,
		UID:      uid,
		Username: username,
		Cmd:      cmd,
		Param:    param,
	}
}

func parseStatus(data []byte) (int, string) {
	uid := 0
	username := ""
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "Uid:") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				uid, _ = strconv.Atoi(parts[1])
			}
		}
	}
	username = lookupUsername(uid)
	return uid, username
}

func lookupUsername(uid int) string {
	f, err := os.Open("/etc/passwd")
	if err != nil {
		return strconv.Itoa(uid)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		parts := strings.Split(scanner.Text(), ":")
		if len(parts) >= 3 {
			if u, err := strconv.Atoi(parts[2]); err == nil && u == uid {
				return parts[0]
			}
		}
	}
	return strconv.Itoa(uid)
}

func parseCmdline(data []byte) (string, string) {
	if len(data) == 0 {
		return "", ""
	}
	parts := strings.Split(string(data), "\x00")
	var nonEmpty []string
	for _, p := range parts {
		if p != "" {
			nonEmpty = append(nonEmpty, p)
		}
	}
	if len(nonEmpty) == 0 {
		return "", ""
	}
	cmd := nonEmpty[0]
	param := ""
	if len(nonEmpty) > 1 {
		param = strings.Join(nonEmpty[1:], " ")
	}
	return cmd, param
}
