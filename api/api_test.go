package api

import (
	"bytes"
	"context"
	"demo/domain"
	"demo/server"
	"demo/store"
	"demo/util"
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
	setup          func(t *testing.T, s *store.MemoryStore, j *MockJWT)
}

// ─── Test Helpers & Runners ──────────────────────────────────────────────────

type MockJWT struct {
	returnToken  string
	returnErr    error
	returnClaims util.CustomClaims
}

func (m MockJWT) Generate(user *domain.User) (string, error) {
	return m.returnToken, m.returnErr
}

func (m MockJWT) Parse(tokenString string) (*util.CustomClaims, error) {
	return &m.returnClaims, m.returnErr
}

func setupTestServer(t *testing.T, tt apiTest) *server.Server {
	t.Helper()

	memStore := store.NewMemoryStore()
	tokens := &MockJWT{}
	if tt.setup != nil {
		tt.setup(t, memStore, tokens)
	}
	h := New(memStore, memStore, memStore, tokens)
	s := server.New(":8080", h.Routes())
	return s
}

func testRoutesInMemory(t *testing.T, tts []apiTest) {
	t.Helper()

	for _, tt := range tts {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			s := setupTestServer(t, tt)

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

	for _, tt := range tts {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			s := setupTestServer(t, tt)
			ts := httptest.NewTestServer(t, s)
			ts.Start()
			client := ts.Client()
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
