package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"goawd/pkg/inotify"
	"goawd/pkg/proc"
	"goawd/pkg/tcpclient"
)

var (
	version   = "dev"
	buildTime = "unknown"
)

func main() {
	serverAddr := flag.String("s", "127.0.0.1", "GoAWD server address")
	serverPort := flag.Int("p", 8023, "GoAWD server port")
	watchDirs := flag.String("w", "/tmp", "Watch directories (separated by ;)")
	interval := flag.Duration("i", 100*time.Millisecond, "Process scan interval")
	daemon := flag.Bool("d", false, "Run as daemon")
	showVersion := flag.Bool("v", false, "Show version")
	flag.Parse()

	if *showVersion {
		log.Printf("goawd-roundworm %s (built %s)", version, buildTime)
		return
	}

	if *daemon {
		if os.Getppid() != 1 {
			args := os.Args[1:]
			cmd := exec.Command(os.Args[0], args...)
			cmd.Stdin = nil
			cmd.Stdout = nil
			cmd.Stderr = nil
			cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
			if err := cmd.Start(); err != nil {
				log.Fatalf("daemonize: %v", err)
			}
			log.Printf("RoundWorm daemonized, pid: %d", cmd.Process.Pid)
			os.Exit(0)
		}
	}

	log.Printf("RoundWorm %s starting, server: %s:%d", version, *serverAddr, *serverPort)

	sender, err := tcpclient.New(*serverAddr, *serverPort)
	if err != nil {
		log.Fatalf("connect to server: %v", err)
	}
	defer sender.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigCh
		log.Printf("received signal %v, shutting down...", sig)
		cancel()
	}()

	dirs := strings.Split(*watchDirs, ";")
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := inotify.Watch(ctx, dirs, sender); err != nil && ctx.Err() == nil {
			log.Printf("inotify watch error: %v", err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		scanner := proc.NewScanner()
		if err := scanner.Scan(ctx, *interval, sender); err != nil && ctx.Err() == nil {
			log.Printf("proc scan error: %v", err)
		}
	}()

	<-ctx.Done()
	wg.Wait()
	log.Println("RoundWorm stopped")
}
