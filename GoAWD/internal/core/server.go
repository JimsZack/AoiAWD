package core

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"goawd/internal/api"
	"goawd/internal/config"
	"goawd/internal/plugin"
	"goawd/internal/storage"
)

type Server struct {
	config      *config.Config
	store       storage.Storage
	hub         *Hub
	pluginMgr   *plugin.Manager
	receiver    *Receiver
	httpServer  *http.Server
	rateLimiter *api.RateLimiter
	startTime   time.Time
}

func NewServer(cfg *config.Config) (*Server, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	store, err := storage.New(cfg.Storage, cfg.FilePath)
	if err != nil {
		return nil, fmt.Errorf("init storage: %w", err)
	}

	hub := NewHub()
	mgr := plugin.NewManager()
	recv := NewReceiver(cfg.TCPAddr, store, hub, mgr)

	s := &Server{
		config:    cfg,
		store:     store,
		hub:       hub,
		pluginMgr: mgr,
		receiver:  recv,
		startTime: time.Now(),
	}
	s.setupRouter()
	return s, nil
}

func (s *Server) setupRouter() {
	mux := http.NewServeMux()

	mux.HandleFunc("/websocket", s.hub.HandleWebSocket)

	v1 := api.NewV1(s.store, s.pluginMgr, s.receiver, s.startTime)
	v1.RegisterRoutes(mux, "/api/v1/", s.config.Token)

	mux.HandleFunc("/", s.serveStatic)
	mux.HandleFunc("/index.html", s.serveStatic)

	// Rate limiter: 600 requests per minute per IP, API endpoints only
	rateLimiter := api.NewRateLimiter(600, time.Minute)
	s.rateLimiter = rateLimiter

	s.httpServer = &http.Server{
		Addr:    s.config.HTTPAddr,
		Handler: rateLimiter.Middleware(s.withCORS(mux)),
	}
}

func (s *Server) withCORS(next http.Handler) http.Handler {
	allowedOrigins := make(map[string]bool)
	for _, o := range strings.Split(s.config.AllowedOrigins, ",") {
		if o = strings.TrimSpace(o); o != "" {
			allowedOrigins[o] = true
		}
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if allowedOrigins[origin] {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
		}
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Token")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) serveStatic(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	if path == "/" {
		path = "/index.html"
	}
	full := strings.TrimPrefix(path, "/")
	if full == "" {
		full = "index.html"
	}

	// Prevent path traversal by cleaning the path and verifying it stays within PublicDir
	cleanPath := filepath.Clean(full)
	fullPath := filepath.Join(s.config.PublicDir, cleanPath)

	// Ensure the resolved path is still within PublicDir
	if !strings.HasPrefix(fullPath, filepath.Clean(s.config.PublicDir)+string(filepath.Separator)) &&
		fullPath != filepath.Clean(s.config.PublicDir) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	http.ServeFile(w, r, fullPath)
}

func (s *Server) RegisterPlugin(p plugin.Plugin) {
	s.pluginMgr.RegisterPlugin(p)
}

func (s *Server) LoadPlugins(dir string) []string {
	return s.pluginMgr.LoadPlugins(dir)
}

func (s *Server) Run(ctx context.Context) error {
	errCh := make(chan error, 3)

	go func() {
		log.Printf("HTTP Server listening on %s", s.config.HTTPAddr)
		if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- fmt.Errorf("http server: %w", err)
		}
	}()

	go func() {
		if err := s.receiver.Start(ctx); err != nil && !errors.Is(err, context.Canceled) {
			errCh <- fmt.Errorf("tcp receiver: %w", err)
		}
	}()

	go s.hub.Run(ctx)

	log.Printf("GoAWD server started")
	log.Printf("Endpoints:")
	log.Printf("  HTTP API:  http://%s/api/v1/info", s.config.HTTPAddr)
	log.Printf("  WebSocket: ws://%s/websocket", s.config.HTTPAddr)
	log.Printf("  TCP Probe: %s", s.config.TCPAddr)

	select {
	case <-ctx.Done():
	case err := <-errCh:
		return err
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	s.httpServer.Shutdown(shutdownCtx)
	if s.rateLimiter != nil {
		s.rateLimiter.Stop()
	}
	s.store.Close()
	log.Println("GoAWD server stopped")
	return nil
}
