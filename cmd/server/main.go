package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"notification-service/internal/config"
	"notification-service/internal/database"
)

func main() {
	cfg := config.Load()

	startupCtx, startupCancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer startupCancel()

	mongoClient, err := database.ConnectMongo(
		startupCtx,
		cfg.MongoURI,
	)
	if err != nil {
		panic(err)
	}

	fmt.Println("Connected to MongoDB")

	db := mongoClient.Database(cfg.MongoDatabase)

	_ = db

	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	server := &http.Server{
		Addr:    ":" + cfg.HTTPPort,
		Handler: mux,
	}

	go func() {
		fmt.Printf("Notification Service listening on %s\n", server.Addr)

		if err := server.ListenAndServe(); err != nil &&
			!errors.Is(err, http.ErrServerClosed) {
			panic(err)
		}
	}()

	shutdownSignal := make(chan os.Signal, 1)

	signal.Notify(
		shutdownSignal,
		syscall.SIGINT,
		syscall.SIGTERM,
	)

	<-shutdownSignal

	fmt.Println("Shutting down Notification Service...")

	shutdownCtx, shutdownCancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		panic(err)
	}

	if err := mongoClient.Disconnect(shutdownCtx); err != nil {
		panic(err)
	}

	fmt.Println("Notification Service stopped")
}
