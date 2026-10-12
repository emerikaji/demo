package main

import (
	"context"
	"demo/api"
	"demo/store"
	"demo/util"
	"errors"
	"log"
	"os"
	"os/signal"
	"syscall"

	"demo/server"

	"golang.org/x/sync/errgroup"
)

// ─── Entrypoint ──────────────────────────────────────────────────────────────

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// ─── API Server ──────────────────────────────────────────────────────

	memStore := store.NewMemoryStore()

	h := api.New(memStore, memStore, memStore, util.NewJWT(os.Getenv("JWT_SECRET")))

	apiServer := server.New(":8080", h.Routes(), h.Middlewares()...)

	g, gCtx := errgroup.WithContext(ctx)

	log.Println("API Server starting...")

	g.Go(func() error {
		return server.Start(gCtx, apiServer)
	})

	log.Println("API Server accepting connections.")

	if err := g.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		log.Fatalf("API Server error: %v", err)
	}

	log.Println("All servers stopped gracefully.")
}

// ─────────────────────────────────────────────────────────────────────────────
