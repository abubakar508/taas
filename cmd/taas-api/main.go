package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/abubakar508/taas/internal/bootstrap"
)

func main() {
	ctx := context.Background()

	app, err := bootstrap.New(ctx)
	if err != nil {
		panic(err)
	}

	go func() {
		addr := fmt.Sprintf(":%s", app.Config.Port)
		app.Logger.Printf("starting server on %s", addr)

		if err := app.HTTP.Listen(addr); err != nil {
			app.Logger.Printf("server stopped: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	app.Logger.Println("shutdown signal received")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), app.Config.ShutdownTimeout)
	defer cancel()

	done := make(chan error, 1)

	go func() {
		if err := app.HTTP.Shutdown(); err != nil {
			done <- err
			return
		}

		app.DB.Close()
		done <- nil
	}()

	select {
	case err := <-done:
		if err != nil {
			app.Logger.Printf("shutdown error: %v", err)
			os.Exit(1)
		}
		app.Logger.Println("shutdown complete")
	case <-shutdownCtx.Done():
		app.Logger.Println("shutdown timeout reached")
		os.Exit(1)
	}

	time.Sleep(100 * time.Millisecond)
}
