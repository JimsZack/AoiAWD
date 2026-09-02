package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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

func TestScriptComment(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"hello", "hello"},
		{"hello\nworld", "helloworld"},
		{"hello\r\nworld", "helloworld"},
		{"it's", "its"},
	}
	for _, tt := range tests {
		if got := scriptComment(tt.in); got != tt.want {
			t.Errorf("scriptComment(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestLastPageOf(t *testing.T) {
	tests := []struct {
		total int64
		count int
		want  int
	}{
		{0, 20, 1},
		{1, 20, 1},
		{20, 20, 1},
		{21, 20, 2},
		{40, 20, 2},
		{41, 20, 3},
		{100, 10, 10},
		{0, 0, 1},
	}
	for _, tt := range tests {
		if got := lastPageOf(tt.total, tt.count); got != tt.want {
			t.Errorf("lastPageOf(%d, %d) = %d, want %d", tt.total, tt.count, got, tt.want)
		}
	}
}

func TestFormatTimeField(t *testing.T) {
	tests := []struct {
		name  string
		input interface{}
		want  string
	}{
		{"float64", float64(1609459200), "2021-01-01 00:00:00"},
		{"int64", int64(1609459200), "2021-01-01 00:00:00"},
		{"int", int(1609459200), "2021-01-01 00:00:00"},
		{"string", "2021-01-01", "2021-01-01"},
		{"nil", nil, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatTimeField(tt.input); got != tt.want {
				t.Errorf("formatTimeField(%v) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		d    time.Duration
		want string
	}{
		{0, "00:00:00"},
		{time.Second, "00:00:01"},
		{time.Minute, "00:01:00"},
		{time.Hour, "01:00:00"},
		{time.Hour + 2*time.Minute + 3*time.Second, "01:02:03"},
	}
	for _, tt := range tests {
		if got := formatDuration(tt.d); got != tt.want {
			t.Errorf("formatDuration(%v) = %q, want %q", tt.d, got, tt.want)
		}
	}
}

func TestParsePage(t *testing.T) {
	tests := []struct {
		name      string
		url       string
		wantPage  int
		wantCount int
	}{
		{"defaults", "/api/v1/listweb", 1, 20},
		{"page1", "/api/v1/listweb?page=1&count=10", 1, 10},
		{"page5", "/api/v1/listweb?page=5&count=5", 5, 5},
		{"zero page", "/api/v1/listweb?page=0", 1, 20},
		{"zero count", "/api/v1/listweb?count=0", 1, 20},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, tt.url, nil)
			page, count := parsePage(r)
			if page != tt.wantPage || count != tt.wantCount {
				t.Errorf("parsePage() = (%d, %d), want (%d, %d)", page, count, tt.wantPage, tt.wantCount)
			}
		})
	}
}

func TestPing(t *testing.T) {
	v, _ := newTestV1(t)
	req := httptest.NewRequest(http.MethodGet, "/api/ping", nil)
	rec := httptest.NewRecorder()
	v.Ping(rec, req)
	if rec.Body.String() != "pong" {
		t.Errorf("Ping() = %q, want %q", rec.Body.String(), "pong")
	}
}

func TestInfo(t *testing.T) {
	v, _ := newTestV1(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/info", nil)
	rec := httptest.NewRecorder()
	v.Info(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("Info() status = %d, want %d", rec.Code, http.StatusOK)
	}
	var resp types.InfoResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Info() response is not valid JSON: %v", err)
	}
}

func TestListWebEmpty(t *testing.T) {
	v, _ := newTestV1(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/listweb?page=1&count=20", nil)
	rec := httptest.NewRecorder()
	v.ListWeb(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("ListWeb() status = %d, want %d", rec.Code, http.StatusOK)
	}
	var resp types.PaginatedResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("ListWeb() response is not valid JSON: %v", err)
	}
	if resp.Page != 1 {
		t.Errorf("ListWeb() page = %d, want 1", resp.Page)
	}
}

func TestListWebWithData(t *testing.T) {
	v, store := newTestV1(t)
	ctx := context.Background()
	for i := 0; i < 5; i++ {
		store.Save(ctx, types.CollWeb, "web"+string(rune('0'+i)), types.WebLogData{
			ID:     "web" + string(rune('0'+i)),
			Time:   time.Now().Unix(),
			Method: "GET",
			URI:    "/test",
			Remote: "127.0.0.1",
		})
	}
	req := httptest.NewRequest(http.MethodGet, "/api/v1/listweb?page=1&count=10", nil)
	rec := httptest.NewRecorder()
	v.ListWeb(rec, req)
	var resp types.PaginatedResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("ListWeb() response is not valid JSON: %v", err)
	}
	items, ok := resp.Data.([]interface{})
	if !ok || len(items) != 5 {
		t.Errorf("ListWeb() data length = %v, want 5", len(items))
	}
}

func TestWebDetailMissing(t *testing.T) {
	v, _ := newTestV1(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/webdetail?id=nonexistent", nil)
	rec := httptest.NewRecorder()
	v.WebDetail(rec, req)
	var resp types.APIResponse
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp.Result != "0" {
		t.Errorf("WebDetail() result = %q, want 0", resp.Result)
	}
}

func TestWebDetailMissingID(t *testing.T) {
	v, _ := newTestV1(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/webdetail", nil)
	rec := httptest.NewRecorder()
	v.WebDetail(rec, req)
	var resp types.APIResponse
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp.Result != "0" {
		t.Errorf("WebDetail() result = %q, want 0", resp.Result)
	}
}

func TestWebDetailFound(t *testing.T) {
	v, store := newTestV1(t)
	ctx := context.Background()
	store.Save(ctx, types.CollWeb, "abc", types.WebLogData{ID: "abc", Method: "GET"})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/webdetail?id=abc", nil)
	rec := httptest.NewRecorder()
	v.WebDetail(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("WebDetail() status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestDownloadWebAutoScriptMissing(t *testing.T) {
	v, _ := newTestV1(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/downloadwebautoscript?id=nonexistent", nil)
	rec := httptest.NewRecorder()
	v.DownloadWebAutoScript(rec, req)
	var resp types.APIResponse
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp.Result != "0" {
		t.Errorf("DownloadWebAutoScript() result = %q, want 0", resp.Result)
	}
}

func TestListPWNEmpty(t *testing.T) {
	v, _ := newTestV1(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/listpwn?page=1&count=20", nil)
	rec := httptest.NewRecorder()
	v.ListPWN(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("ListPWN() status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestPWNDetailMissingID(t *testing.T) {
	v, _ := newTestV1(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/pwndetail", nil)
	rec := httptest.NewRecorder()
	v.PWNDetail(rec, req)
	var resp types.APIResponse
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp.Result != "0" {
		t.Errorf("PWNDetail() result = %q, want 0", resp.Result)
	}
}

func TestPWNDetailNotFound(t *testing.T) {
	v, _ := newTestV1(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/pwndetail?id=nonexistent", nil)
	rec := httptest.NewRecorder()
	v.PWNDetail(rec, req)
	var resp types.APIResponse
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp.Result != "0" {
		t.Errorf("PWNDetail() result = %q, want 0", resp.Result)
	}
}

func TestDownloadPWNMissing(t *testing.T) {
	v, _ := newTestV1(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/downloadpwn?id=nonexistent", nil)
	rec := httptest.NewRecorder()
	v.DownloadPWN(rec, req)
	var resp types.APIResponse
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp.Result != "0" {
		t.Errorf("DownloadPWN() result = %q, want 0", resp.Result)
	}
}

func TestGeneratePWNBinInvalidJSON(t *testing.T) {
	v, _ := newTestV1(t)
	body := strings.NewReader("not json")
	req := httptest.NewRequest(http.MethodPost, "/api/v1/generatepwnbin", body)
	rec := httptest.NewRecorder()
	v.GeneratePWNBin(rec, req)
	var resp types.APIResponse
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp.Result != "0" {
		t.Errorf("GeneratePWNBin() result = %q, want 0", resp.Result)
	}
}

func TestGeneratePWNBinMissingFields(t *testing.T) {
	v, _ := newTestV1(t)
	body := strings.NewReader(`{"binary":"AAECAw=="}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/generatepwnbin", body)
	rec := httptest.NewRecorder()
	v.GeneratePWNBin(rec, req)
	var resp types.APIResponse
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp.Result != "0" {
		t.Errorf("GeneratePWNBin() result = %q, want 0", resp.Result)
	}
}

func TestListFilesystemEmpty(t *testing.T) {
	v, _ := newTestV1(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/listfilesystem?page=1&count=20", nil)
	rec := httptest.NewRecorder()
	v.ListFilesystem(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("ListFilesystem() status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestDownloadFileNotFound(t *testing.T) {
	v, _ := newTestV1(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/downloadfile?id=nonexistent", nil)
	rec := httptest.NewRecorder()
	v.DownloadFile(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("DownloadFile() status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestListProcessEmpty(t *testing.T) {
	v, _ := newTestV1(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/listprocess?page=1&count=20", nil)
	rec := httptest.NewRecorder()
	v.ListProcess(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("ListProcess() status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestListCurrentProcess(t *testing.T) {
	v, _ := newTestV1(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/listcurrentprocess", nil)
	rec := httptest.NewRecorder()
	v.ListCurrentProcess(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("ListCurrentProcess() status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestCurrentProcess(t *testing.T) {
	v, _ := newTestV1(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/currentprocess", nil)
	rec := httptest.NewRecorder()
	v.CurrentProcess(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("CurrentProcess() status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestListAlertEmpty(t *testing.T) {
	v, _ := newTestV1(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/listalert?page=1&count=20", nil)
	rec := httptest.NewRecorder()
	v.ListAlert(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("ListAlert() status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestListPluginEmpty(t *testing.T) {
	v, _ := newTestV1(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/listplugin", nil)
	rec := httptest.NewRecorder()
	v.ListPlugin(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("ListPlugin() status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestReloadPlugin(t *testing.T) {
	v, _ := newTestV1(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/reloadplugin", nil)
	rec := httptest.NewRecorder()
	v.ReloadPlugin(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("ReloadPlugin() status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestDownloadPWNAutoScriptMissingID(t *testing.T) {
	v, _ := newTestV1(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/downloadpwnautoscript", nil)
	rec := httptest.NewRecorder()
	v.DownloadPWNAutoScript(rec, req)
	var resp types.APIResponse
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp.Result != "0" {
		t.Errorf("DownloadPWNAutoScript() result = %q, want 0", resp.Result)
	}
}

func TestDownloadPWNAutoScriptNotFound(t *testing.T) {
	v, _ := newTestV1(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/downloadpwnautoscript?id=nonexistent", nil)
	rec := httptest.NewRecorder()
	v.DownloadPWNAutoScript(rec, req)
	var resp types.APIResponse
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp.Result != "0" {
		t.Errorf("DownloadPWNAutoScript() result = %q, want 0", resp.Result)
	}
}

func TestTokenAuth(t *testing.T) {
	v, _ := newTestV1(t)
	mux := http.NewServeMux()
	v.RegisterRoutes(mux, "/api/v1/", "mytoken")

	// Without token
	req := httptest.NewRequest(http.MethodGet, "/api/v1/info", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Errorf("no token: status = %d, want %d", rec.Code, http.StatusForbidden)
	}

	// With token in header
	req = httptest.NewRequest(http.MethodGet, "/api/v1/info", nil)
	req.Header.Set("Token", "mytoken")
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("valid token: status = %d, want %d", rec.Code, http.StatusOK)
	}

	// With token in query
	req = httptest.NewRequest(http.MethodGet, "/api/v1/info?token=mytoken", nil)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("valid query token: status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestToMap(t *testing.T) {
	type TestStruct struct {
		Name string `json:"name"`
	}
	m := toMap(&TestStruct{Name: "test"})
	if m == nil {
		t.Fatal("toMap() returned nil")
	}
	if m["name"] != "test" {
		t.Errorf("toMap() name = %v, want test", m["name"])
	}
}

func TestToMapNil(t *testing.T) {
	m := toMap(nil)
	if m != nil {
		t.Errorf("toMap(nil) = %v, want nil", m)
	}
}
