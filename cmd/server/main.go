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
	"notification-service/internal/httpapi"
	"notification-service/internal/messaging"
	"notification-service/internal/notification"
	"notification-service/internal/realtime"
	"notification-service/internal/repository"
)

func main() {
	// --------------------------------------------------
	// Config
	// --------------------------------------------------

	cfg := config.Load()

	// --------------------------------------------------
	// Application Context
	// --------------------------------------------------

	appCtx, appCancel := context.WithCancel(
		context.Background(),
	)
	defer appCancel()

	// --------------------------------------------------
	// Startup Context
	// --------------------------------------------------

	startupCtx, startupCancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer startupCancel()

	// --------------------------------------------------
	// MongoDB
	// --------------------------------------------------

	mongoClient, err := database.ConnectMongo(
		startupCtx,
		cfg.MongoURI,
	)
	if err != nil {
		panic(err)
	}

	fmt.Println("Connected to MongoDB")

	db := mongoClient.Database(
		cfg.MongoDatabase,
	)

	// --------------------------------------------------
	// Repositories
	// --------------------------------------------------

	notificationRepo :=
		repository.NewNotificationRepository(db)

	outboxRepo :=
		repository.NewOutboxRepository(db)

	// --------------------------------------------------
	// MongoDB Indexes
	// --------------------------------------------------

	if err := notificationRepo.EnsureIndexes(
		startupCtx,
	); err != nil {
		panic(err)
	}

	if err := outboxRepo.EnsureIndexes(
		startupCtx,
	); err != nil {
		panic(err)
	}

	fmt.Println("MongoDB indexes ready")

	// --------------------------------------------------
	// NATS
	// --------------------------------------------------

	natsConn, err := messaging.Connect(
		cfg.NATSURL,
	)
	if err != nil {
		panic(err)
	}

	fmt.Println("Connected to NATS")

	// --------------------------------------------------
	// JetStream
	// --------------------------------------------------

	js, err := messaging.JetStream(
		natsConn,
	)
	if err != nil {
		panic(err)
	}

	if err := messaging.EnsureTicketStream(
		js,
	); err != nil {
		panic(err)
	}

	fmt.Println("JetStream ready")

	// --------------------------------------------------
	// Realtime Hub
	// --------------------------------------------------

	hub := realtime.NewHub()

	// --------------------------------------------------
	// Notification Service
	// --------------------------------------------------

	notificationService :=
		notification.NewService(
			notificationRepo,
		)

	// --------------------------------------------------
	// Realtime Event Publisher
	// --------------------------------------------------

	eventPublisher :=
		messaging.NewEventPublisher(
			natsConn,
		)

	// --------------------------------------------------
	// Outbox Worker
	// --------------------------------------------------

	outboxWorker :=
		notification.NewOutboxWorker(
			outboxRepo,
			eventPublisher,
			1*time.Second,
			100,
		)

	go outboxWorker.Run(appCtx)

	// --------------------------------------------------
	// Ticket Consumer
	// --------------------------------------------------

	ticketConsumer :=
		messaging.NewTicketConsumer(
			notificationService,
		)

	ticketSubscription, err :=
		ticketConsumer.Start(js)

	if err != nil {
		panic(err)
	}

	fmt.Println("Ticket consumer started")

	// --------------------------------------------------
	// Realtime Consumer
	// --------------------------------------------------

	realtimeConsumer :=
		messaging.NewRealtimeConsumer(
			hub,
		)

	realtimeSubscription, err :=
		realtimeConsumer.Start(
			natsConn,
		)

	if err != nil {
		panic(err)
	}

	fmt.Println("Realtime consumer started")

	// --------------------------------------------------
	// HTTP Handlers
	// --------------------------------------------------

	notificationHandler :=
		httpapi.NewNotificationHandler(
			notificationService,
		)

	realtimeHandler :=
		realtime.NewHandler(
			hub,
		)

	// --------------------------------------------------
	// Router
	// --------------------------------------------------

	mux := http.NewServeMux()

	mux.HandleFunc(
		"GET /health",
		func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			w.WriteHeader(
				http.StatusOK,
			)

			_, _ = w.Write(
				[]byte("OK"),
			)
		},
	)

	mux.HandleFunc(
		"GET /v1/notifications",
		notificationHandler.List,
	)

	mux.HandleFunc(
		"POST /v1/notifications/{id}/read",
		notificationHandler.MarkAsRead,
	)

	mux.HandleFunc(
		"GET /ws",
		realtimeHandler.ServeWS,
	)

	// --------------------------------------------------
	// HTTP Server
	// --------------------------------------------------

	server := &http.Server{
		Addr: ":" + cfg.HTTPPort,

		Handler: mux,

		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		fmt.Printf(
			"Notification Service listening on %s\n",
			server.Addr,
		)

		if err := server.ListenAndServe(); err != nil &&
			!errors.Is(
				err,
				http.ErrServerClosed,
			) {
			panic(err)
		}
	}()

	// --------------------------------------------------
	// Shutdown Signal
	// --------------------------------------------------

	shutdownSignal := make(
		chan os.Signal,
		1,
	)

	signal.Notify(
		shutdownSignal,
		syscall.SIGINT,
		syscall.SIGTERM,
	)

	<-shutdownSignal

	fmt.Println(
		"Shutting down Notification Service...",
	)

	// --------------------------------------------------
	// Stop Background Workers
	// --------------------------------------------------

	appCancel()

	// --------------------------------------------------
	// Shutdown Context
	// --------------------------------------------------

	shutdownCtx, shutdownCancel :=
		context.WithTimeout(
			context.Background(),
			5*time.Second,
		)

	defer shutdownCancel()

	// --------------------------------------------------
	// Stop HTTP Server
	// --------------------------------------------------

	if err := server.Shutdown(
		shutdownCtx,
	); err != nil {
		fmt.Printf(
			"HTTP server shutdown error: %v\n",
			err,
		)
	}

	// --------------------------------------------------
	// Stop Ticket Consumer
	// --------------------------------------------------

	if err := ticketSubscription.Unsubscribe(); err != nil {
		fmt.Printf(
			"Ticket subscription shutdown error: %v\n",
			err,
		)
	}

	// --------------------------------------------------
	// Stop Realtime Consumer
	// --------------------------------------------------

	if err := realtimeSubscription.Unsubscribe(); err != nil {
		fmt.Printf(
			"Realtime subscription shutdown error: %v\n",
			err,
		)
	}

	// --------------------------------------------------
	// Drain NATS
	// --------------------------------------------------

	if err := natsConn.Drain(); err != nil {
		fmt.Printf(
			"NATS drain error: %v\n",
			err,
		)
	}

	// --------------------------------------------------
	// Disconnect MongoDB
	// --------------------------------------------------

	if err := mongoClient.Disconnect(
		shutdownCtx,
	); err != nil {
		fmt.Printf(
			"MongoDB shutdown error: %v\n",
			err,
		)
	}

	fmt.Println(
		"Notification Service stopped",
	)
}
