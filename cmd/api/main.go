package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"backend-sport-team-report-go/internal/bootstrap"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	app, err := bootstrap.NewApp(ctx)
	if err != nil {
		log.Fatalf("bootstrap app: %v", err)
	}

	if err := app.Start(ctx); err != nil {
		log.Fatalf("start app: %v", err)
	}

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), app.ShutdownTimeout())
	defer cancel()

	if err := app.Stop(shutdownCtx); err != nil {
		log.Fatalf("stop app: %v", err)
	}
}
