package worker

import (
	"log"
	"time"

	"github.com/hibiken/asynq"

	certadapter "github.com/ALLAN-star-glitch/nuruvent-backend/internal/app/adapters/certificate"
	certservice "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/certificate/service"
	certwork "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/certificate/worker"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/notification/notification-domain"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/notification/service"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/config"
)

// StartEmbeddedWorker initializes and starts the Asynq workers.
//
// Two independent asynq.Server instances are run:
//
//  1. Notification worker — critical/default/low queues, Concurrency=6.
//     Handles OTP, welcome emails, invites, payments, registration,
//     certificate-ready delivery, etc.
//
//  2. Certificate render worker — certificates queue only, Concurrency=4.
//     Isolated so a burst of certificate renders can NEVER starve the
//     notification queue, and so the render cap is a hard limit (not a
//     weighted share). This protects Gotenberg from 429s under load.
//
// Returns a shutdown function that gracefully drains both servers.
func StartEmbeddedWorker(
	cfg *config.Config,
	certSvc certservice.Service,
) func() {
	log.Println("🚀 Initializing embedded workers...")

	log.Printf("📧 Email Configuration:")
	log.Printf("   API Key: %s", maskString(cfg.Email.APIKey))
	log.Printf("   From: %s", cfg.Email.From)
	if cfg.Email.APIKey == "" {
		log.Printf("⚠️ WARNING: EMAIL_API_KEY is empty! Check your configuration")
	}

	// ------------------------------------------------------------
	// Shared: notification worker instance
	// ------------------------------------------------------------
	emailConfig := service.EmailChannelConfig{
		EMAIL_API_KEY: cfg.Email.APIKey,
		EMAIL_FROM:    cfg.Email.From,
	}
	emailChannel := service.NewEmailChannel(emailConfig)
	notificationWorker := service.NewNotificationWorker(emailChannel)

	// ------------------------------------------------------------
	// Redis options (shared by both servers)
	// ------------------------------------------------------------
	redisURL := cfg.GetRedisURL()
	redisOpt, err := asynq.ParseRedisURI(redisURL)
	if err != nil {
		log.Fatalf("Failed to parse Redis URI for Asynq worker (%s): %v", redisURL, err)
	}

	// ============================================================
	// SERVER 1 — Notifications
	// ============================================================
	notifSrv := asynq.NewServer(
		redisOpt,
		asynq.Config{
			Concurrency: 6,
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

	notifMux := asynq.NewServeMux()

	notifMux.HandleFunc(notificationdomain.TaskVerificationOTP, notificationWorker.HandleVerificationOTP)
	notifMux.HandleFunc(notificationdomain.TaskWelcomeIndividual, notificationWorker.HandleWelcomeIndividual)
	notifMux.HandleFunc(notificationdomain.TaskWelcomeInstitution, notificationWorker.HandleWelcomeInstitution)
	notifMux.HandleFunc(notificationdomain.TaskPasswordResetConfirm, notificationWorker.HandlePasswordResetConfirm)
	notifMux.HandleFunc(notificationdomain.TaskLoginNotification, notificationWorker.HandleLoginNotification)
	notifMux.HandleFunc(notificationdomain.TaskWelcomeInstitutionKYC, notificationWorker.HandleWelcomeInstitutionKYC)
	notifMux.HandleFunc(notificationdomain.TaskNewPersonalAccountRegistration, notificationWorker.HandleNewPersonalAccountRegistration)
	notifMux.HandleFunc(notificationdomain.TaskNewInstitutionAccountRegistration, notificationWorker.HandleNewInstitutionAccountRegistration)
	notifMux.HandleFunc(notificationdomain.TaskCertificateReady, notificationWorker.HandleCertificateReady)
	notifMux.HandleFunc(notificationdomain.TaskTeamInviteExistingUser, notificationWorker.HandleTeamInviteExistingUser)
	notifMux.HandleFunc(notificationdomain.TaskTeamInviteRegistration, notificationWorker.HandleTeamInviteRegistration)
	notifMux.HandleFunc(notificationdomain.TaskTeamInviteAccepted, notificationWorker.HandleTeamInviteAccepted)
	notifMux.HandleFunc(notificationdomain.TaskTeamInviteDeclined, notificationWorker.HandleTeamInviteDeclined)
	notifMux.HandleFunc(notificationdomain.TaskPaymentInitiated, notificationWorker.HandlePaymentInitiated)
	notifMux.HandleFunc(notificationdomain.TaskPaymentSucceeded, notificationWorker.HandlePaymentSucceeded)
	notifMux.HandleFunc(notificationdomain.TaskPaymentFailed, notificationWorker.HandlePaymentFailed)
	notifMux.HandleFunc(notificationdomain.TaskPaymentExpired, notificationWorker.HandlePaymentExpired)
	notifMux.HandleFunc(notificationdomain.TaskRefundIssued, notificationWorker.HandleRefundIssued)
	notifMux.HandleFunc(notificationdomain.TaskRegistrationConfirmed, notificationWorker.HandleRegistrationConfirmed)

	log.Println("✅ Notification worker configured (Concurrency=6)")
	log.Println("📋 Notification tasks:")
	log.Printf("   - %s", notificationdomain.TaskVerificationOTP)
	log.Printf("   - %s", notificationdomain.TaskWelcomeIndividual)
	log.Printf("   - %s", notificationdomain.TaskWelcomeInstitution)
	log.Printf("   - %s", notificationdomain.TaskPasswordResetConfirm)
	log.Printf("   - %s", notificationdomain.TaskLoginNotification)
	log.Printf("   - %s", notificationdomain.TaskWelcomeInstitutionKYC)
	log.Printf("   - %s", notificationdomain.TaskNewPersonalAccountRegistration)
	log.Printf("   - %s", notificationdomain.TaskNewInstitutionAccountRegistration)
	log.Printf("   - %s", notificationdomain.TaskCertificateReady)
	log.Printf("   - %s", notificationdomain.TaskTeamInviteExistingUser)
	log.Printf("   - %s", notificationdomain.TaskTeamInviteRegistration)
	log.Printf("   - %s", notificationdomain.TaskTeamInviteAccepted)
	log.Printf("   - %s", notificationdomain.TaskTeamInviteDeclined)
	log.Printf("   - %s", notificationdomain.TaskPaymentInitiated)
	log.Printf("   - %s", notificationdomain.TaskPaymentSucceeded)
	log.Printf("   - %s", notificationdomain.TaskPaymentFailed)
	log.Printf("   - %s", notificationdomain.TaskPaymentExpired)
	log.Printf("   - %s", notificationdomain.TaskRefundIssued)
	log.Printf("   - %s", notificationdomain.TaskRegistrationConfirmed)

	// ============================================================
	// SERVER 2 — Certificate renders (isolated, hard cap 4)
	// ============================================================
	var certSrv *asynq.Server
	var certMux *asynq.ServeMux

	if certSvc != nil {
		certSrv = asynq.NewServer(
			redisOpt,
			asynq.Config{
				Concurrency: 4, // ← HARD cap: at most 4 concurrent Gotenberg renders
				Queues: map[string]int{
					"certificates": 1, // only listen to the certificates queue
				},
				RetryDelayFunc: func(n int, err error, task *asynq.Task) time.Duration {
					// Exponential-ish backoff for renders: 30s, 60s, 90s...
					return time.Duration(n) * 30 * time.Second
				},
			},
		)

		renderHandler := certwork.NewRenderCertificateHandler(certSvc)

		certMux = asynq.NewServeMux()
		certMux.Handle(certadapter.TaskTypeRenderCertificate, renderHandler)

		log.Println("✅ Certificate render worker configured (Concurrency=4)")
		log.Printf("   - %s (certificate PDF render)", certadapter.TaskTypeRenderCertificate)
	} else {
		log.Println("⚠️ Certificate service not provided — render task disabled")
	}

	// ============================================================
	// Start both workers in background goroutines
	// ============================================================
	go func() {
		log.Println("🚀 Notification worker active (critical=6, default=3, low=1)")
		if err := notifSrv.Run(notifMux); err != nil {
			log.Printf("❌ Notification worker stopped: %v", err)
		}
	}()

	if certSrv != nil {
		go func() {
			log.Println("🚀 Certificate render worker active (certificates=4)")
			if err := certSrv.Run(certMux); err != nil {
				log.Printf("❌ Certificate render worker stopped: %v", err)
			}
		}()
	}

	// ============================================================
	// Shutdown closure
	// ============================================================
	return func() {
		log.Println("🛑 Shutting down embedded workers...")
		notifSrv.Shutdown()
		if certSrv != nil {
			certSrv.Shutdown()
		}
		log.Println("✅ Embedded workers stopped")
	}
}

func maskString(s string) string {
	if s == "" {
		return "[EMPTY]"
	}
	if len(s) <= 8 {
		return "***"
	}
	return s[:4] + "..." + s[len(s)-4:]
}