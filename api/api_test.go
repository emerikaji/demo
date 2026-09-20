package api

import (
	"bytes"
	"context"
	"demo/server"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type apiTest struct {
	name           string
	route          string
	method         string
	body           []byte
	expectedStatus int
	expectedBody   string
}

var tests = []apiTest{
	{"demo should OK", "/demo", http.MethodGet, nil, http.StatusOK, ""},
}

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

func TestRoutes_InMemory(t *testing.T) {
	h := New()
	s := server.New(":0", h.Routes())

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			doInMemory(t, s, tt)
		})
	}
}

func TestRoutes_HTTP(t *testing.T) {
	h := New()
	s := server.New(":0", h.Routes())
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
}

// ─────────────────────────────────────────────────────────────────────────────
