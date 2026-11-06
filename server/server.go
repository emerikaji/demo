// Package server is an implementation of server logic to generate http servers at will.
package server

import (
	"context"
	"errors"
	"log"
	"net/http"
	"time"
)

// ─── Server Generator ────────────────────────────────────────────────────────

// Server wraps http.Handler.
type Server struct {
	mux         *http.ServeMux
	middlewares []func(w http.ResponseWriter, r *http.Request)
}

type LoggingResponseWriter struct {
	http.ResponseWriter
	status int
}

func (lrw *LoggingResponseWriter) WriteHeader(status int) {
	lrw.ResponseWriter.WriteHeader(status)
	lrw.status = status
}

// ServeHTTP wraps the httpServer's method to implement the http.Handler interface.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	lrw := &LoggingResponseWriter{ResponseWriter: w}

	for _, middleware := range s.middlewares {
		middleware(lrw, r)
	}
	s.mux.ServeHTTP(lrw, r)

	duration := time.Since(start)

	log.Printf("%s %s %d %v %s", r.Method, r.URL.Path, lrw.status, duration, r.RemoteAddr)
}

// New is a generator that creates an http.Server with the provided address and an http.ServeMux handler.
// Each provided route is registered to the multiplexer.
func New(addr string, routes map[string]http.HandlerFunc, middlewares ...func(w http.ResponseWriter, r *http.Request)) *http.Server {
	mux := http.NewServeMux()

	for pattern, handler := range routes {
		mux.HandleFunc(pattern, handler)
	}

	return &http.Server{
		Addr:              addr,
		Handler:           &Server{mux: mux, middlewares: middlewares},
		ReadHeaderTimeout: 3 * time.Second,
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       120 * time.Second,
	}
}

// Start opens the server for connections and prepares for graceful shutdown once the provided context expires.
func Start(ctx context.Context, s *http.Server) error {
	errChan := make(chan error, 1)

	go func() {
		if err := s.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
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

		return s.Shutdown(shutdownCtx)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
