// internal/shared/worker/worker.go

package worker

import (
	"log"
	"time"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/notification/notification-domain"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/notification/service"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/config"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/redis"
	"github.com/hibiken/asynq"
)

// StartEmbeddedWorker initializes and starts the Asynq worker server asynchronously.
// It returns a shutdown function to gracefully drain and stop tasks when the API exits.
func StartEmbeddedWorker(cfg *config.Config) func() {
	log.Println("🚀 Initializing embedded notification worker...")

	log.Printf("📧 Email Configuration:")
	log.Printf("   API Key: %s", maskString(cfg.Email.APIKey))
	log.Printf("   From: %s", cfg.Email.From)

	if cfg.Email.APIKey == "" {
		log.Printf("⚠️ WARNING: EMAIL_API_KEY is empty! Check your configuration")
	}

	// 1. Create Email Channel & Worker
	emailConfig := service.EmailChannelConfig{
		EMAIL_API_KEY: cfg.Email.APIKey,
		EMAIL_FROM:    cfg.Email.From,
	}

	emailChannel := service.NewEmailChannel(emailConfig)
	notificationWorker := service.NewNotificationWorker(emailChannel)

	// ✅ Parse Redis URL correctly for Asynq
	redisURL := cfg.GetRedisURL()
	redisOpts, err := redis.ParseURL(redisURL)
	if err != nil {
		log.Printf("⚠️ Failed to parse Redis URL for Asynq, using as host:port: %v", err)
		redisOpts = &asynq.RedisClientOpt{
			Addr: redisURL,
		}
	} else {
		// Convert redis.Options to asynq.RedisClientOpt
		redisOpts = &asynq.RedisClientOpt{
			Addr:     redisOpts.Addr,
			Password: redisOpts.Password,
			DB:       redisOpts.DB,
		}
	}

	// 2. Configure Asynq Server with parsed Redis options
	srv := asynq.NewServer(
		*redisOpts,
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

	// 3. Register Task Handlers
	mux := asynq.NewServeMux()
	mux.HandleFunc(notificationdomain.TaskVerificationOTP, notificationWorker.HandleVerificationOTP)
	mux.HandleFunc(notificationdomain.TaskWelcomeIndividual, notificationWorker.HandleWelcomeIndividual)
	mux.HandleFunc(notificationdomain.TaskWelcomeInstitution, notificationWorker.HandleWelcomeInstitution)
	mux.HandleFunc(notificationdomain.TaskPasswordResetConfirm, notificationWorker.HandlePasswordResetConfirm)
	mux.HandleFunc(notificationdomain.TaskLoginNotification, notificationWorker.HandleLoginNotification)

	log.Println("✅ All task handlers registered")
	log.Println("📋 Registered tasks:")
	log.Printf("   - %s (unified for all OTP purposes)", notificationdomain.TaskVerificationOTP)
	log.Printf("   - %s", notificationdomain.TaskWelcomeIndividual)
	log.Printf("   - %s", notificationdomain.TaskWelcomeInstitution)
	log.Printf("   - %s", notificationdomain.TaskPasswordResetConfirm)
	log.Printf("   - %s", notificationdomain.TaskLoginNotification)

	// 4. Start Worker in a Background Goroutine
	go func() {
		log.Println("🚀 Asynq notification worker active. Listening for tasks...")
		log.Printf("📊 Queue priorities: critical=6, default=3, low=1")
		if err := srv.Run(mux); err != nil {
			log.Printf("❌ Asynq worker execution stopped: %v", err)
		}
	}()

	// 5. Return closure for clean shutdown
	return func() {
		log.Println("🛑 Shutting down embedded Asynq worker...")
		srv.Shutdown()
		log.Println("✅ Embedded worker stopped")
	}
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