// cmd/worker/main.go

package main

import (
	"context"
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
	// Debug: Log email configuration
	// ================================================
	log.Printf("📧 Email Configuration:")
	log.Printf("   API Key: %s", maskString(cfg.Email.EMAIL_API_KEY))
	log.Printf("   From: %s", cfg.Email.EMAIL_FROM)

	if cfg.Email.EMAIL_API_KEY == "" {
		log.Printf("⚠️ WARNING: EMAIL_API_KEY is empty! Check your .env file")
	}
	if cfg.Email.EMAIL_FROM == "" {
		log.Printf("⚠️ WARNING: EMAIL_FROM is empty! Check your .env file")
	}

	// ================================================
	// Initialize Redis Client
	// ================================================
	redisClient, err := redis.NewClient(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize Redis: %v", err)
	}
	log.Println("✅ Redis initialized successfully")

	// ================================================
	// Create Email Channel (Implements Outbound Port)
	// ================================================
	emailConfig := service.EmailChannelConfig{
		EMAIL_API_KEY: cfg.Email.EMAIL_API_KEY,
		EMAIL_FROM:    cfg.Email.EMAIL_FROM,
	}

	// ✅ EmailChannel implements notificationdomain.Channel (outbound port)
	emailChannel := service.NewEmailChannel(emailConfig)
	if emailChannel == nil {
		log.Fatalf("❌ Failed to create email channel")
	}
	log.Println("✅ Email channel created (implements notificationdomain.Channel)")

	// ================================================
	// Create Worker (Implements Inbound Port)
	// ================================================
	// ✅ Worker implements notificationdomain.TaskProcessor (inbound port)
	// ✅ Depends on notificationdomain.Channel (outbound port)
	notificationWorker := service.NewNotificationWorker(emailChannel)
	if notificationWorker == nil {
		log.Fatalf("❌ Failed to create notification worker")
	}
	log.Println("✅ Notification worker created (implements TaskProcessor interface)")

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
	log.Println("✅ Asynq server created")

	// ================================================
	// Register Task Handlers (Unified - No Deprecated Tasks)
	// ================================================
	mux := asynq.NewServeMux()

	// Register all handlers
	// ✅ Unified OTP handler - handles ALL OTP purposes
	mux.HandleFunc(notificationdomain.TaskVerificationOTP, notificationWorker.HandleVerificationOTP)
	
	// Welcome handlers
	mux.HandleFunc(notificationdomain.TaskWelcomeIndividual, notificationWorker.HandleWelcomeIndividual)
	mux.HandleFunc(notificationdomain.TaskWelcomeInstitution, notificationWorker.HandleWelcomeInstitution)
	
	// Password reset confirm
	mux.HandleFunc(notificationdomain.TaskPasswordResetConfirm, notificationWorker.HandlePasswordResetConfirm)
	
	// Login notification
	mux.HandleFunc(notificationdomain.TaskLoginNotification, notificationWorker.HandleLoginNotification)

	log.Println("✅ All task handlers registered")
	log.Println("📋 Registered tasks:")
	log.Printf("   - %s (unified for all OTP purposes)", notificationdomain.TaskVerificationOTP)
	log.Printf("   - %s", notificationdomain.TaskWelcomeIndividual)
	log.Printf("   - %s", notificationdomain.TaskWelcomeInstitution)
	log.Printf("   - %s", notificationdomain.TaskPasswordResetConfirm)
	log.Printf("   - %s", notificationdomain.TaskLoginNotification)

	// ================================================
	// Start Worker
	// ================================================
	go func() {
		log.Println("🚀 Asynq notification worker started. Listening for tasks...")
		log.Printf("📊 Queue priorities: critical=6, default=3, low=1")
		if err := srv.Run(mux); err != nil {
			log.Fatalf("❌ Worker failed: %v", err)
		}
	}()

	// ================================================
	// Graceful Shutdown
	// ================================================
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("🛑 Shutting down notification worker gracefully...")

	// Graceful shutdown
	_, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	srv.Shutdown()

	// ✅ Close Redis connection
	if err := redisClient.Close(); err != nil {
		log.Printf("Error closing Redis: %v", err)
	}

	log.Println("✅ Notification worker stopped")
}

// maskString masks a string for logging (shows first 4 and last 4 characters)
func maskString(s string) string {
	if s == "" {
		return "[EMPTY]"
	}
	if len(s) <= 8 {
		return "***"
	}
	return s[:4] + "..." + s[len(s)-4:]
}