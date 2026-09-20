package server

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name           string
		addr           string
		routes         map[string]http.HandlerFunc
		testRoute      string
		method         string
		expectedStatus int
	}{
		{
			name: "registered route responds with 200 OK",
			addr: ":8080",
			routes: map[string]http.HandlerFunc{
				"GET /demo": func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
				},
			},
			testRoute:      "/demo",
			method:         http.MethodGet,
			expectedStatus: http.StatusOK,
		},
		{
			name: "unregistered route returns 404 Not Found",
			addr: ":8080",
			routes: map[string]http.HandlerFunc{
				"GET /demo": func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
				},
			},
			testRoute:      "/missing",
			method:         http.MethodGet,
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := New(tt.addr, tt.routes)

			if s.httpServer.Addr != tt.addr {
				t.Errorf("got addr %q, want %q", s.httpServer.Addr, tt.addr)
			}

			req := httptest.NewRequest(tt.method, tt.testRoute, nil)
			rec := httptest.NewRecorder()

			s.httpServer.Handler.ServeHTTP(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("route %q returned status %d, want %d", tt.testRoute, rec.Code, tt.expectedStatus)
			}
		})
	}
}

func Test_Lifecycle(t *testing.T) {
	routes := map[string]http.HandlerFunc{
		"GET /demo": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		},
	}

	s := New(":8080", routes)

	// Setup context for lifecycle control
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errChan := make(chan error, 1)

	// Start the server asynchronously
	go func() {
		errChan <- s.Start(ctx)
	}()

	// Verify the server is live and accepting requests
	client := &http.Client{Timeout: 1 * time.Second}
	resp, err := client.Get("http://localhost:8080/demo")
	if err != nil {
		t.Log(<-errChan)
		t.Fatalf("failed to send request to started server: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("got status %d, want %d", resp.StatusCode, http.StatusOK)
	}

	// Trigger graceful shutdown
	cancel()

	// Assert server shuts down cleanly within a reasonable timeout
	select {
	case err := <-errChan:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			t.Fatalf("server exited with unexpected error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("server failed to shut down within timeout")
	}
}

// ─────────────────────────────────────────────────────────────────────────────
