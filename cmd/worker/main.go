package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/config"
	authemail "github.com/ALLAN_star_glitch/nuruvent-backend/internal/modules/auth/auth_email"
	"github.com/ALLAN_star_glitch/nuruvent-backend/pkg/email"
	"github.com/ALLAN_star_glitch/nuruvent-backend/pkg/redis"
	"github.com/hibiken/asynq"
)

func main() {
	// Load configuration
	cfg := config.Load()
	log.Printf("Starting Nuruvent email worker in %s mode", cfg.Environment)

	// ================================================
	// Initialize Redis
	// ================================================
	if err := redis.Init(cfg.Redis.URL); err != nil {
		log.Fatalf("Failed to initialize Redis: %v", err)
	}
	log.Println("Redis initialized successfully")

	// ================================================
	// Initialize Email Service
	// ================================================
	emailService := email.NewEmailService(cfg.Email.APIKey, cfg.Email.From)
	emailWorker := authemail.NewEmailWorker(emailService)

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
			// Retry delay for failed tasks
			RetryDelayFunc: func(n int, err error, task *asynq.Task) time.Duration {
				return time.Duration(n) * 30 * time.Second
			},
		},
	)

	// ================================================
	// Register Task Handlers
	// ================================================
	mux := asynq.NewServeMux()

	// Verification OTP (generic - handles registration, email change, phone change, password reset, 2FA, business verification)
	mux.HandleFunc(authemail.TypeVerificationOTP, emailWorker.HandleVerificationOTP)

	// Welcome emails
	mux.HandleFunc(authemail.TypeWelcome, emailWorker.HandleWelcome)
	mux.HandleFunc(authemail.TypeBusinessWelcome, emailWorker.HandleBusinessWelcome)

	// Security emails
	mux.HandleFunc(authemail.TypeTwoFactorOTP, emailWorker.HandleTwoFactorOTP)
	mux.HandleFunc(authemail.TypePasswordResetOTP, emailWorker.HandlePasswordResetOTP)
	mux.HandleFunc(authemail.TypePasswordResetConfirm, emailWorker.HandlePasswordResetConfirm)
	mux.HandleFunc(authemail.TypeLoginNotification, emailWorker.HandleLoginNotification)
	mux.HandleFunc(authemail.TypeEmailIndividualProfessionalWelcome, emailWorker.HandleIndividualCoachWelcome)


	// ================================================
	// Start Worker
	// ================================================
	go func() {
		log.Println("Asynq email worker started. Listening for tasks...")
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

	log.Println("Shutting down email worker gracefully...")
	srv.Shutdown()
	log.Println("Email worker stopped")
}