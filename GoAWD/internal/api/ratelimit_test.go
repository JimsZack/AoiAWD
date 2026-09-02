package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRateLimiterMiddlewareOnlyLimitsAPI(t *testing.T) {
	rl := NewRateLimiter(2, time.Minute)
	defer rl.Stop()

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	handler := rl.Middleware(next)

	// Static assets must never consume the API budget: the panel polls them.
	for i := 0; i < 10; i++ {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/index.html", nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("static request %d was limited: %d", i, rec.Code)
		}
	}

	for i := 0; i < 2; i++ {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/info", nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("api request %d should pass: %d", i, rec.Code)
		}
	}

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/listweb", nil))
	if rec.Code != http.StatusTooManyRequests {
		t.Errorf("third api request = %d, want %d", rec.Code, http.StatusTooManyRequests)
	}
}

func TestRateLimiterStopIsIdempotent(t *testing.T) {
	rl := NewRateLimiter(1, time.Minute)
	rl.Stop()
	rl.Stop()
}
