// internal/shared/worker/worker.go

package worker

import (
	"log"
	"time"

	"github.com/hibiken/asynq"

	certadapter "github.com/ALLAN-star-glitch/nuruvent-backend/internal/app/adapters/certificate"
	certservice "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/certificate/service"
	certwork "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/certificate/worker"
	eventsDomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/events/domain"
	eventsservice "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/events/service"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/notification/notification-domain"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/notification/service"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/config"
)

// StartEmbeddedWorker initializes and starts the Asynq workers.
//
// Three independent asynq.Server instances are run:
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
//  3. Video provisioning worker — video queue only, Concurrency=2.
//     Isolated so Zoom/Meet API calls don't compete with email sends.
//     Retries every 5 minutes (provisioning is idempotent).
//
// Additionally, a daily Asynq scheduler runs at 02:00 EAT to sweep
// events with unprovisioned video schedules that are now within the
// 30-day provisioning window.
//
// Returns a shutdown function that gracefully drains all servers and
// stops the scheduler.
func StartEmbeddedWorker(
	cfg *config.Config,
	certSvc certservice.Service,
	eventsSvc eventsservice.Service,
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
	// Redis options (shared by all servers)
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
				Concurrency: 4,
				Queues: map[string]int{
					"certificates": 1,
				},
				RetryDelayFunc: func(n int, err error, task *asynq.Task) time.Duration {
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
	// SERVER 3 — Video Provisioning (isolated, Concurrency=2)
	// ============================================================
	var videoSrv *asynq.Server
	var videoMux *asynq.ServeMux

	if eventsSvc != nil {
		videoSrv = asynq.NewServer(
			redisOpt,
			asynq.Config{
				Concurrency: 2, // Zoom/Meet API rate limits; keep low
				Queues: map[string]int{
					"video": 1,
				},
				RetryDelayFunc: func(n int, err error, task *asynq.Task) time.Duration {
					// Provisioning is idempotent — retry slowly.
					return 5 * time.Minute
				},
			},
		)

		videoWorker := NewVideoProvisioner(eventsSvc)

		videoMux = asynq.NewServeMux()
		videoMux.HandleFunc(eventsDomain.TaskProvisionVideoMeetings, videoWorker.HandleProvisionVideoMeetings)

		log.Println("✅ Video provisioning worker configured (Concurrency=2)")
		log.Printf("   - %s", eventsDomain.TaskProvisionVideoMeetings)
	} else {
		log.Println("⚠️ Events service not provided — video provisioning disabled")
	}

	// ============================================================
	// SCHEDULER — Daily video sweep at 02:00 EAT
	// ============================================================
	//
	// Sweeps events with virtual schedules that are within 30 days of
	// starting but still have no meeting provisioned. Catches:
	//   - Events created 60+ days out
	//   - Events where immediate provisioning failed and was retried
	//   - Events where the schedule was edited to move a session into
	//     the current 30-day window
	//
	// Enabled only when the events service is available.
	var scheduler *asynq.Scheduler
	if eventsSvc != nil {
		scheduler = asynq.NewScheduler(
			redisOpt,
			&asynq.SchedulerOpts{
				Location: time.FixedZone("EAT", 3*60*60), // Africa/Nairobi
			},
		)

		sweepTask := asynq.NewTask(
			eventsDomain.TaskProvisionVideoMeetings,
			[]byte(`{}`), // empty payload = sweep mode
		)
		sweepEntryID, err := scheduler.Register(
			"0 2 * * *", // daily at 02:00 EAT
			sweepTask,
			asynq.Queue("video"),
		)
		if err != nil {
			log.Printf("⚠️  Failed to register daily video sweep: %v", err)
			scheduler = nil
		} else {
			log.Printf("✅ Daily video sweep scheduled (cron=0 2 * * *, id=%s)", sweepEntryID)
		}
	}

	// ============================================================
	// Start all workers in background goroutines
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

	if videoSrv != nil {
		go func() {
			log.Println("🚀 Video provisioning worker active (video=2)")
			if err := videoSrv.Run(videoMux); err != nil {
				log.Printf("❌ Video provisioning worker stopped: %v", err)
			}
		}()
	}

	if scheduler != nil {
		go func() {
			log.Println("🚀 Daily video sweep scheduler active")
			if err := scheduler.Run(); err != nil {
				log.Printf("❌ Scheduler stopped: %v", err)
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
		if videoSrv != nil {
			videoSrv.Shutdown()
		}
		if scheduler != nil {
			scheduler.Shutdown()
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