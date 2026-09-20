// Package server is an implementation of server logic to generate http servers at will.
package server

import (
	"context"
	"errors"
	"net/http"
	"time"
)

// Server wraps http.Server.
type Server struct {
	httpServer *http.Server
}

// ServeHTTP wraps the httpServer's method to implement the http.Handler interface.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.httpServer.Handler.ServeHTTP(w, r)
}

// New is a Server generator that creates an http.Server with the provided address and an http.ServeMux handler.
// Each provided route is registered to the multiplexer.
func New(addr string, routes map[string]http.HandlerFunc) *Server {
	mux := http.NewServeMux()

	for pattern, handler := range routes {
		mux.HandleFunc(pattern, handler)
	}

	return &Server{
		httpServer: &http.Server{
			Addr:              addr,
			Handler:           mux,
			ReadHeaderTimeout: 3 * time.Second,
			ReadTimeout:       5 * time.Second,
			WriteTimeout:      10 * time.Second,
			IdleTimeout:       120 * time.Second,
		},
	}
}

// Start opens the server for connections and prepares for graceful shutdown once the provided context expires.
func (s *Server) Start(ctx context.Context) error {
	errChan := make(chan error, 1)

	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errChan <- err
		}

		close(errChan)
	}()

	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		return s.httpServer.Shutdown(shutdownCtx)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
