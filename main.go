package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"demo/server"

	"golang.org/x/sync/errgroup"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	apiRoutes := map[string]http.HandlerFunc{
		"GET /": func(w http.ResponseWriter, r *http.Request) {
			_, err := w.Write([]byte("Hello World"))

			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		},
	}

	apiServer := server.New(":8080", apiRoutes)

	g, gCtx := errgroup.WithContext(ctx)

	log.Println("Server starting...")

	g.Go(func() error {
		return apiServer.Start(gCtx)
	})

	log.Println("Server accepting connections")

	if err := g.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		log.Fatalf("Server error: %v", err)
	}

	log.Println("Server stopped gracefully")
}
