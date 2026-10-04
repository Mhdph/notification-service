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
	"notification-service/internal/identity"
	"notification-service/internal/messaging"
	"notification-service/internal/notification"
	"notification-service/internal/realtime"
	"notification-service/internal/repository"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func main() {
	cfg := config.Load()

	if cfg.JWTSecret == "" {
		panic("JWT_SECRET is required")
	}

	appCtx, appCancel :=
		context.WithCancel(
			context.Background(),
		)
	defer appCancel()

	startupCtx, startupCancel :=
		context.WithTimeout(
			context.Background(),
			10*time.Second,
		)
	defer startupCancel()

	// --------------------------------------------------
	// MongoDB
	// --------------------------------------------------

	mongoClient, err :=
		database.ConnectMongo(
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
		repository.NewNotificationRepository(
			db,
		)

	outboxRepo :=
		repository.NewOutboxRepository(
			db,
		)

	// --------------------------------------------------
	// Indexes
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

	natsConn, err :=
		messaging.Connect(
			cfg.NATSURL,
		)

	if err != nil {
		panic(err)
	}

	fmt.Println("Connected to NATS")

	js, err :=
		messaging.JetStream(
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
	// Outbox
	// --------------------------------------------------

	eventPublisher :=
		messaging.NewEventPublisher(
			natsConn,
		)

	workerID :=
		bson.NewObjectID().Hex()

	outboxWorker :=
		notification.NewOutboxWorker(
			outboxRepo,
			eventPublisher,
			workerID,
			time.Second,
			30*time.Second,
			100,
		)

	fmt.Printf(
		"Outbox worker ID: %s\n",
		workerID,
	)

	go outboxWorker.Run(
		appCtx,
	)

	// --------------------------------------------------
	// Domain Event Consumer
	// --------------------------------------------------

	ticketConsumer :=
		messaging.NewTicketConsumer(
			notificationService,
		)

	ticketSubscription, err :=
		ticketConsumer.Start(
			js,
		)

	if err != nil {
		panic(err)
	}

	fmt.Println(
		"Ticket consumer started",
	)

	// --------------------------------------------------
	// Realtime Event Consumer
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

	fmt.Println(
		"Realtime consumer started",
	)

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

	identityMiddleware :=
		identity.NewMiddleware(
			cfg.JWTSecret,
		)

	// --------------------------------------------------
	// Public Router
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

	// --------------------------------------------------
	// Protected Router
	// --------------------------------------------------

	protectedMux :=
		http.NewServeMux()

	protectedMux.HandleFunc(
		"GET /v1/notifications",
		notificationHandler.List,
	)

	protectedMux.HandleFunc(
		"GET /v1/notifications/unread-count",
		notificationHandler.UnreadCount,
	)

	protectedMux.HandleFunc(
		"POST /v1/notifications/{id}/read",
		notificationHandler.MarkAsRead,
	)

	protectedMux.HandleFunc(
		"POST /v1/notifications/read-all",
		notificationHandler.MarkAllAsRead,
	)

	protectedMux.HandleFunc(
		"GET /ws",
		realtimeHandler.ServeWS,
	)

	protectedHandler :=
		identityMiddleware.Handle(
			protectedMux,
		)

	mux.Handle(
		"/v1/",
		protectedHandler,
	)

	mux.Handle(
		"/ws",
		protectedHandler,
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
	// Shutdown
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

	appCancel()

	shutdownCtx, shutdownCancel :=
		context.WithTimeout(
			context.Background(),
			5*time.Second,
		)
	defer shutdownCancel()

	if err := server.Shutdown(
		shutdownCtx,
	); err != nil {
		fmt.Printf(
			"HTTP server shutdown error: %v\n",
			err,
		)
	}

	if err := ticketSubscription.Unsubscribe(); err != nil {
		fmt.Printf(
			"Ticket subscription shutdown error: %v\n",
			err,
		)
	}

	if err := realtimeSubscription.Unsubscribe(); err != nil {
		fmt.Printf(
			"Realtime subscription shutdown error: %v\n",
			err,
		)
	}

	if err := natsConn.Drain(); err != nil {
		fmt.Printf(
			"NATS drain error: %v\n",
			err,
		)
	}

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
