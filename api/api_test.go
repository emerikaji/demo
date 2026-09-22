package api

import (
	"bytes"
	"context"
	"demo/server"
	"demo/store"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// ─── Overall Test Structure ──────────────────────────────────────────────────

type apiTest struct {
	name           string
	route          string
	method         string
	body           []byte
	expectedStatus int
	expectedBody   string
}

// ─── Test Helpers & Runners ──────────────────────────────────────────────────

func setupTestServer(t *testing.T) *server.Server {
	t.Helper()

	memStore := store.NewMemoryStore()
	h := New(memStore, memStore, memStore)
	s := server.New(":8080", h.Routes())
	return s
}

func testRoutesInMemory(t *testing.T, tts []apiTest) {
	t.Helper()

	s := setupTestServer(t)

	for _, tt := range tts {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

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
		})
	}
}

func testRoutesHTTP(t *testing.T, tts []apiTest) {
	t.Helper()

	s := setupTestServer(t)

	ts := httptest.NewTestServer(t, s)
	ts.Start()
	client := ts.Client()

	for _, tt := range tts {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			url := ts.URL + tt.route

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
		})
	}
}

// ─────────────────────────────────────────────────────────────────────────────
