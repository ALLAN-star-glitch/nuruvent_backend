// internal/modules/notification/providers.go

package notification

import (
	"github.com/google/wire"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/notification/notification-domain"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/notification/service"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/config"
)

var ProviderSet = wire.NewSet(
	// Configuration
	NewEmailConfig,

	// Channels (returns interface)
	service.NewEmailChannel,
	NewEmailChannels,

	// Task enqueuer (returns interface) ✅
	service.NewTaskEnqueuer,

	// Services (returns interface)
	service.NewNotificationServiceWithQueue,

	// Worker
	service.NewNotificationWorker,

	// ❌ No wire.Bind needed!
)

func NewEmailConfig(cfg *config.Config) service.EmailChannelConfig {
	return service.EmailChannelConfig{
		EMAIL_API_KEY: cfg.Email.EMAIL_API_KEY,
		EMAIL_FROM:    cfg.Email.EMAIL_FROM,
	}
}

func NewEmailChannels(emailChannel notificationdomain.Channel) []notificationdomain.Channel {
	return []notificationdomain.Channel{emailChannel}
}