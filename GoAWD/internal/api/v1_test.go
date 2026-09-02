package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"goawd/internal/plugin"
	"goawd/internal/storage"
	"goawd/internal/types"
)

type stubProcessProvider struct{}

func (stubProcessProvider) CurrentProcessPIDs() []int                { return nil }
func (stubProcessProvider) CurrentProcessList() []*types.ProcessInfo { return nil }

func newTestV1(t *testing.T) (*V1, storage.Storage) {
	t.Helper()
	store := storage.NewMemory()
	return NewV1(store, plugin.NewManager(), stubProcessProvider{}, time.Now()), store
}

func TestShellQuote(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"/index.php", "'/index.php'"},
		{"", "''"},
		{"/a'b", `'/a'\''b'`},
		{"$(id)", `'$(id)'`},
		{"; rm -rf /", `'; rm -rf /'`},
	}
	for _, tt := range tests {
		if got := shellQuote(tt.in); got != tt.want {
			t.Errorf("shellQuote(%q) = %s, want %s", tt.in, got, tt.want)
		}
	}
}

// Captured web traffic is fully attacker controlled, so a generated replay
// script must never let it break out of the quoted argument.
func TestDownloadWebAutoScriptQuotesAttackerControlledFields(t *testing.T) {
	v, store := newTestV1(t)
	ctx := context.Background()

	const uri = "/$(touch /tmp/pwned)"
	const remote = "1.2.3.4\n" + uri
	if err := store.Save(ctx, types.CollWeb, "abc", types.WebLogData{
		ID:     "abc",
		Time:   time.Now().Unix(),
		Method: "GET",
		URI:    uri,
		Remote: remote,
	}); err != nil {
		t.Fatalf("Save: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/downloadwebautoscript?id=abc", nil)
	rec := httptest.NewRecorder()
	v.DownloadWebAutoScript(rec, req)

	want := "curl -sS -X " + shellQuote("GET") + " " + shellQuote("http://"+remote+uri)
	if body := rec.Body.String(); !strings.Contains(body, want) {
		t.Errorf("script does not pass the URL as a single quoted argument\nwant substring: %s\ngot:\n%s", want, body)
	}
}

func TestDownloadWebAutoScriptSanitisesFilename(t *testing.T) {
	v, store := newTestV1(t)
	ctx := context.Background()

	nastyID := "a'b\r\nSet-Cookie: evil=1"
	if err := store.Save(ctx, types.CollWeb, nastyID, types.WebLogData{ID: nastyID}); err != nil {
		t.Fatalf("Save: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/downloadwebautoscript?id="+url.QueryEscape(nastyID), nil)
	rec := httptest.NewRecorder()
	v.DownloadWebAutoScript(rec, req)

	cd := rec.Header().Get("Content-Disposition")
	if cd == "" {
		t.Fatal("missing Content-Disposition header")
	}
	if strings.ContainsAny(cd, "\r\n") {
		t.Errorf("Content-Disposition contains CRLF (header injection): %q", cd)
	}
	if strings.Contains(cd, "'") {
		t.Errorf("Content-Disposition contains quotes: %q", cd)
	}
}

func TestGeneratePWNBinEmbedsBinary(t *testing.T) {
	v, _ := newTestV1(t)

	body := strings.NewReader(`{"binary":"AAECAw==","host":"127.0.0.1; rm -rf /","port":8023}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/generatepwnbin", body)
	rec := httptest.NewRecorder()
	v.GeneratePWNBin(rec, req)

	script := rec.Body.String()
	if !strings.Contains(script, "AAECAw==") {
		t.Errorf("generated script does not embed the binary:\n%s", script)
	}
	if !strings.Contains(script, shellQuote("127.0.0.1; rm -rf /")) {
		t.Errorf("host is not shell quoted:\n%s", script)
	}
}
