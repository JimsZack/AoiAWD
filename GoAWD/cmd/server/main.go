package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"goawd/internal/config"
	"goawd/internal/core"
	"goawd/internal/plugin"

	_ "goawd/plugins/flagbuster"
	_ "goawd/plugins/kingwatcher"
	_ "goawd/plugins/zombiekiller"
)

var (
	version   = "dev"
	buildTime = "unknown"
)

func main() {
	cfg := config.Default()

	flag.StringVar(&cfg.HTTPAddr, "http", cfg.HTTPAddr, "HTTP server bind address")
	flag.StringVar(&cfg.TCPAddr, "tcp", cfg.TCPAddr, "TCP receiver bind address")
	flag.StringVar(&cfg.MongoDBURI, "mongo", cfg.MongoDBURI, "MongoDB URI (legacy, ignored)")
	flag.StringVar(&cfg.Database, "db", cfg.Database, "Database name")
	flag.StringVar(&cfg.Token, "token", "", "Access token (random if empty)")
	flag.StringVar(&cfg.PluginDir, "plugins", cfg.PluginDir, "Plugin directory")
	flag.StringVar(&cfg.PublicDir, "public", cfg.PublicDir, "Frontend static files directory")
	flag.StringVar(&cfg.Storage, "storage", cfg.Storage, "Storage backend: memory|file")
	flag.StringVar(&cfg.FilePath, "file-path", cfg.FilePath, "JSON storage file path (for file backend)")
	showVersion := flag.Bool("v", false, "Show version")
	flag.Parse()

	if *showVersion {
		log.Printf("GoAWD %s (built %s)", version, buildTime)
		return
	}

	log.Printf("GoAWD %s (built %s) starting...", version, buildTime)

	tokenProvided := cfg.Token != ""

	server, err := core.NewServer(cfg)
	if err != nil {
		log.Fatalf("failed to create server: %v", err)
	}

	if !tokenProvided {
		// Validate() generated a random token: print it or the panel is unusable.
		log.Printf("No token supplied, generated access token: %s", cfg.Token)
	}

	for _, p := range plugin.Registered() {
		server.RegisterPlugin(p)
	}
	registered := plugin.Registered()
	names := make([]string, len(registered))
	for i, p := range registered {
		names[i] = p.Name()
	}
	log.Printf("Loaded built-in plugins: %v", names)

	server.LoadPlugins(cfg.PluginDir)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigCh
		log.Printf("received signal %v, shutting down...", sig)
		cancel()
	}()

	if err := server.Run(ctx); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
