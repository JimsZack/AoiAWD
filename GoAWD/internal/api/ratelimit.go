package api

import (
	"net/http"
	"sync"
	"time"
)

type RateLimiter struct {
	mu       sync.Mutex
	clients  map[string]*clientInfo
	limit    int
	window   time.Duration
}

type clientInfo struct {
	count    int
	resetAt  time.Time
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		clients: make(map[string]*clientInfo),
		limit:   limit,
		window:  window,
	}
	
	// Cleanup goroutine
	go func() {
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			rl.cleanup()
		}
	}()
	
	return rl
}

func (rl *RateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	now := time.Now()
	for ip, info := range rl.clients {
		if now.After(info.resetAt) {
			delete(rl.clients, ip)
		}
	}
}

func (rl *RateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	
	now := time.Now()
	info, exists := rl.clients[ip]
	
	if !exists || now.After(info.resetAt) {
		rl.clients[ip] = &clientInfo{
			count:   1,
			resetAt: now.Add(rl.window),
		}
		return true
	}
	
	if info.count >= rl.limit {
		return false
	}
	
	info.count++
	return true
}

func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr
		if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
			ip = fwd
		}
		
		if !rl.Allow(ip) {
			http.Error(w, `{"result":"0","message":"Rate limit exceeded"}`, http.StatusTooManyRequests)
			return
		}
		
		next.ServeHTTP(w, r)
	})
}