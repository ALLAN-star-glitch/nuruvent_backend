package notification

import (
	"github.com/google/wire"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/notification/notification-domain"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/notification/service"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/config"
)

// ProviderSet provides all notification module dependencies
var ProviderSet = wire.NewSet(
	// Email Channel Config
	ProvideEmailChannelConfig,  // ← Capitalized

	// Email Channel
	service.NewEmailChannel,

	// Provide channel as slice
	ProvideChannelSlice,  // ← Capitalized

	// Task Enqueuer
	service.NewTaskEnqueuer,

	// Notification Service with Queue support
	service.NewNotificationServiceWithQueue,

	// Worker
	service.NewNotificationWorker,
)

// ProvideEmailChannelConfig creates email channel config from app config
func ProvideEmailChannelConfig(cfg *config.Config) service.EmailChannelConfig {  // ← Capitalized
	return service.EmailChannelConfig{
		APIKey: cfg.Email.APIKey,
		From:   cfg.Email.From,
	}
}

// ProvideChannelSlice returns a slice containing the email channel
func ProvideChannelSlice(emailChannel notificationdomain.Channel) []notificationdomain.Channel {  // ← Capitalized
	return []notificationdomain.Channel{emailChannel}
}