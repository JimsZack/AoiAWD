package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"goawd/internal/types"
)

var (
	version   = "dev"
	buildTime = "unknown"
)

func main() {
	binary := flag.String("binary", "", "PWN binary to run")
	serverHost := flag.String("host", "127.0.0.1", "GoAWD server host")
	serverPort := flag.Int("port", 8023, "GoAWD server port")
	showVersion := flag.Bool("v", false, "Show version")
	flag.Parse()

	if *showVersion {
		log.Printf("goawd-guardian %s (built %s)", version, buildTime)
		return
	}

	if *binary == "" {
		log.Fatal("binary is required: -binary <path>")
	}
	if _, err := os.Stat(*binary); err != nil {
		log.Fatalf("binary not found: %v", err)
	}

	log.Printf("Guardian %s starting, binary: %s, server: %s:%d", version, *binary, *serverHost, *serverPort)

	cmd := exec.Command(*binary)
	cmd.Stderr = os.Stderr

	stdinPipe, err := cmd.StdinPipe()
	if err != nil {
		log.Fatalf("stdin pipe: %v", err)
	}
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatalf("stdout pipe: %v", err)
	}

	if err := cmd.Start(); err != nil {
		log.Fatalf("start binary: %v", err)
	}

	pid := cmd.Process.Pid
	log.Printf("PWN process started, pid: %d", pid)

	time.Sleep(200 * time.Millisecond)
	maps := readMemoryMap(pid)

	serverAddr := net.JoinHostPort(*serverHost, strconv.Itoa(*serverPort))
	stdoutConn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		log.Fatalf("connect stdout: %v", err)
	}
	defer stdoutConn.Close()

	stdinConn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		log.Fatalf("connect stdin: %v", err)
	}
	defer stdinConn.Close()

	exeName := filepath.Base(*binary)
	stdoutInit := types.ProbeMessage{
		Type: types.MsgTypePWN,
		Data: types.PWNInitData{File: exeName, Type: "stdout", PID: pid, Maps: maps},
	}
	if err := json.NewEncoder(stdoutConn).Encode(stdoutInit); err != nil {
		log.Fatalf("send stdout init: %v", err)
	}

	stdinInit := types.ProbeMessage{
		Type: types.MsgTypePWN,
		Data: types.PWNInitData{File: exeName, Type: "stdin", PID: pid, Maps: maps},
	}
	if err := json.NewEncoder(stdinConn).Encode(stdinInit); err != nil {
		log.Fatalf("send stdin init: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		cancel()
		if cmd.Process != nil {
			cmd.Process.Signal(syscall.SIGTERM)
		}
	}()

	done := make(chan struct{}, 3)

	go func() {
		defer func() { done <- struct{}{} }()
		buf := make([]byte, 4096)
		for {
			n, err := stdoutPipe.Read(buf)
			if n > 0 {
				os.Stdout.Write(buf[:n])
				// Send raw bytes without adding newline (binary safe)
				stdoutConn.Write(buf[:n])
			}
			if err != nil {
				return
			}
			select {
			case <-ctx.Done():
				return
			default:
			}
		}
	}()

	go func() {
		defer func() { done <- struct{}{} }()
		// Read raw bytes from stdin, not line by line (interactive mode)
		buf := make([]byte, 4096)
		for {
			n, err := os.Stdin.Read(buf)
			if n > 0 {
				stdinPipe.Write(buf[:n])
				stdinConn.Write(buf[:n])
			}
			if err != nil {
				return
			}
			select {
			case <-ctx.Done():
				return
			default:
			}
		}
	}()

	go func() {
		cmd.Wait()
		done <- struct{}{}
	}()

	<-done
	log.Println("Guardian shutting down")
}

func readMemoryMap(pid int) string {
	path := fmt.Sprintf("/proc/%d/maps", pid)
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return base64.StdEncoding.EncodeToString(data)
}
