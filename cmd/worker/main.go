package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"autocheck-microservices/internal/app"
	"autocheck-microservices/pkg/config"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	if err := app.RunWorker(ctx, config.Load()); err != nil {
		log.Fatal(err)
	}
}
