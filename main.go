package main

import (
	"context"
	"demo/api"
	"demo/store"
	"errors"
	"log"
	"os"
	"os/signal"
	"syscall"

	"demo/server"

	"golang.org/x/sync/errgroup"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	memStore := store.NewMemoryStore()

	h := api.New(memStore, memStore, memStore)

	apiServer := server.New(":8080", h.Routes())

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
