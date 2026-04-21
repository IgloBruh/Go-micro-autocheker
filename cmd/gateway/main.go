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
	cfg := config.Load()
	if err := cfg.ValidateGateway(); err != nil {
		log.Fatal(err)
	}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	if err := app.RunGateway(ctx, cfg); err != nil {
		log.Fatal(err)
	}
}
