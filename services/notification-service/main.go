package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"

	"notification-service/internal/application/services"
	"notification-service/internal/config"
	postgres "notification-service/internal/infrastructure"
	"notification-service/internal/infrastructure/email"
	"notification-service/internal/infrastructure/rabbitmq"
	"notification-service/internal/interface/routes"
	"notification-service/pkg/logger"
	"notification-service/pkg/metrics"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	appLogger := logger.New(cfg.LogLevel)
	metrics.Init()

	db, err := postgres.NewConntection(cfg.Database)
	if err != nil {
		appLogger.Fatalf("failed to connect to db: %v", err)
	}
	defer db.Close()

	if err := postgres.RunMigrations(db); err != nil {
		appLogger.Fatal("failed to run migrations: " + err.Error())
	}

	notificationRepo := postgres.NewNotificationRepository(db)
	notificationService := services.NewNotificationService(notificationRepo, appLogger)

	emailSender := email.NewSender(cfg.Email.ResendAPIKey, cfg.Email.FromAddress, appLogger)
	if !emailSender.Enabled() {
		appLogger.Warn("RESEND_API_KEY not set: emails will be logged, not sent")
	}
	rfqEmailService := services.NewRFQEmailService(notificationService, emailSender, cfg.Email.FrontendURL, appLogger)

	rabbitMQClient := rabbitmq.NewClient(cfg.RabbitMQ, appLogger)

	if err := rabbitMQClient.Connect(); err != nil {
		appLogger.Fatal("failed to connect to rabbit " + err.Error())
	}
	defer rabbitMQClient.Close()

	messageHanlder := func(routingKey string, body []byte) error {
		switch routingKey {
		case "rfq.created":
			return rfqEmailService.ProcessRFQCreatedEvent(context.Background(), body)
		case "quote.submitted":
			return rfqEmailService.ProcessQuoteSubmittedEvent(context.Background(), body)
		default:
			appLogger.Warn("received unsupported routing key: " + routingKey)
			return nil
		}
	}

	if err := rabbitMQClient.StartConsuming(messageHanlder); err != nil {
		appLogger.Fatal("failed to start consuming messages " + err.Error())
	}

	// On connection/channel close, exit fatally so the container restart policy
	// respawns the process and re-runs Connect + StartConsuming from scratch.
	go func() {
		if closeErr := <-rabbitMQClient.NotifyClose(); closeErr != nil {
			appLogger.Fatal("rabbit connection closed, exiting for restart: " + closeErr.Error())
		}
	}()

	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(metrics.GinMiddleware("notification-service"))
	router.GET("/metrics", gin.WrapH(metrics.Handler()))

	routes.SetupNotificationRoutes(router, notificationService, rabbitMQClient, appLogger)

	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if err := notificationService.CleanupOldNotifications(context.Background(), cfg.Notification.CleanupDays); err != nil {
					appLogger.Error("failed to cleanup old notifs: " + err.Error())
				}
			}
		}
	}()

	go func() {
		appLogger.Info("notif server starting on port " + cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			appLogger.Fatal("failed to start server " + err.Error())
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	appLogger.Info("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		appLogger.Fatal("server forced to shutdown: " + err.Error())
	}

	appLogger.Info("server exited")
}
