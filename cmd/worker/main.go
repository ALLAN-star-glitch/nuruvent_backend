package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hibiken/asynq"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/notification/notification-domain"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/notification/service"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/config"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/redis"
)

func main() {
	// Load configuration
	cfg := config.Load()
	log.Printf("Starting Nuruvent notification worker in %s mode", cfg.Environment)

	// ================================================
	// Initialize Redis
	// ================================================
	if err := redis.Init(cfg.Redis.URL); err != nil {
		log.Fatalf("Failed to initialize Redis: %v", err)
	}
	log.Println("Redis initialized successfully")

	// ================================================
	// Initialize Notification Service
	// ================================================
	// ✅ Create email channel config
	emailConfig := service.EmailChannelConfig{
		APIKey: cfg.Email.APIKey,
		From:   cfg.Email.From,
	}

	// ✅ Create email channel using config
	emailChannel := service.NewEmailChannel(emailConfig)

	// Create notification service with channels
	notificationService := service.NewNotificationService(emailChannel)

	// Create notification worker
	notificationWorker := service.NewNotificationWorker(notificationService)

	// ================================================
	// Create Asynq Server
	// ================================================
	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: cfg.Redis.URL},
		asynq.Config{
			Concurrency: 10,
			Queues: map[string]int{
				"critical": 6,
				"default":  3,
				"low":      1,
			},
			RetryDelayFunc: func(n int, err error, task *asynq.Task) time.Duration {
				return time.Duration(n) * 30 * time.Second
			},
		},
	)

	// ================================================
	// Register Task Handlers
	// ================================================
	mux := asynq.NewServeMux()

	// Verification OTP
	mux.HandleFunc(notificationdomain.TaskVerificationOTP, notificationWorker.HandleVerificationOTP)

	// Welcome emails - Individual & Institution
	mux.HandleFunc(notificationdomain.TaskWelcomeIndividual, notificationWorker.HandleWelcomeIndividual)
	mux.HandleFunc(notificationdomain.TaskWelcomeInstitution, notificationWorker.HandleWelcomeInstitution)

	// Security emails
	mux.HandleFunc(notificationdomain.TaskTwoFactorOTP, notificationWorker.HandleTwoFactorOTP)
	mux.HandleFunc(notificationdomain.TaskPasswordResetOTP, notificationWorker.HandlePasswordResetOTP)
	mux.HandleFunc(notificationdomain.TaskPasswordResetConfirm, notificationWorker.HandlePasswordResetConfirm)
	mux.HandleFunc(notificationdomain.TaskLoginNotification, notificationWorker.HandleLoginNotification)

	// ================================================
	// Start Worker
	// ================================================
	go func() {
		log.Println("Asynq notification worker started. Listening for tasks...")
		log.Printf("Queue priorities: critical=6, default=3, low=1")
		if err := srv.Run(mux); err != nil {
			log.Fatalf("Worker failed: %v", err)
		}
	}()

	// ================================================
	// Graceful Shutdown
	// ================================================
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down notification worker gracefully...")
	srv.Shutdown()
	log.Println("Notification worker stopped")
}