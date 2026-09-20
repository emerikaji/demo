package api

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"demo/server"
	"demo/store"
)

type apiTest struct {
	name           string
	route          string
	method         string
	body           []byte
	expectedStatus int
	expectedBody   string
}

// ─── Test Cases Table ────────────────────────────────────────────────────────

var tests = []apiTest{
	{
		name:           "health check should return status available",
		route:          "/v1/health",
		method:         http.MethodGet,
		body:           nil,
		expectedStatus: http.StatusOK,
		expectedBody:   `{"status":"available"}`,
	},
}

// ─── Test Helpers & Runners ──────────────────────────────────────────────────

func doInMemory(t *testing.T, s *server.Server, tt apiTest) {
	t.Helper()

	var bodyReader io.Reader
	if tt.body != nil {
		bodyReader = bytes.NewReader(tt.body)
	}

	req := httptest.NewRequest(tt.method, tt.route, bodyReader)
	rec := httptest.NewRecorder()

	req.Header.Set("Content-Type", "application/json")

	s.ServeHTTP(rec, req)

	if rec.Code != tt.expectedStatus {
		t.Errorf("got status %d, want %d", rec.Code, tt.expectedStatus)
	}

	gotBody := string(bytes.TrimSpace(rec.Body.Bytes()))
	if gotBody != tt.expectedBody {
		t.Errorf("got body %q, want %q", gotBody, tt.expectedBody)
	}
}

func doHTTP(t *testing.T, client *http.Client, url string, tt apiTest) {
	t.Helper()

	var bodyReader io.Reader
	if tt.body != nil {
		bodyReader = bytes.NewReader(tt.body)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, tt.method, url, bodyReader)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("request failed (%s %s): %v", tt.method, url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != tt.expectedStatus {
		t.Errorf("got status %d, want %d", resp.StatusCode, tt.expectedStatus)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}

	gotBody := string(bytes.TrimSpace(respBody))
	if gotBody != tt.expectedBody {
		t.Errorf("got body %q, want %q", gotBody, tt.expectedBody)
	}
}

func setupTestServer() *server.Server {
	memStore := store.NewMemoryStore()
	h := New(memStore, memStore, memStore)
	s := server.New(":8080", h.Routes())
	return s
}

func TestRoutes_InMemory(t *testing.T) {
	s := setupTestServer()

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			doInMemory(t, s, tt)
		})
	}
}

func TestRoutes_HTTP(t *testing.T) {
	s := setupTestServer()
	ts := httptest.NewTestServer(t, s)
	ts.Start()

	client := ts.Client()

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			doHTTP(t, client, ts.URL+tt.route, tt)
		})
	}

	t.Cleanup(ts.Close)
}

// ─────────────────────────────────────────────────────────────────────────────
