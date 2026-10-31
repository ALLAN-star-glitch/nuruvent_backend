// internal/shared/worker/video_provisioner.go

package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/hibiken/asynq"

	eventsDomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/events/domain"
	eventsService "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/events/service"
)

// videoProvisionPayload is the worker-side payload. It matches the
// JSON the events service enqueues. Lives here (not in domain) because
// JSON is a transport concern.
type videoProvisionPayload struct {
	EventID string `json:"event_id,omitempty"`
}

// VideoProvisioner handles the video:provision_meetings task.
//
// Two modes:
//   - Single event (payload.EventID set): provision that event's
//     unprovisioned schedules. Enqueued after CreateEvent.
//   - Sweep (payload.EventID empty): find every event with
//     unprovisioned schedules in the 30-day window and provision
//     them. Run daily via the scheduler.
type VideoProvisioner struct {
	eventsService eventsService.Service
}

func NewVideoProvisioner(events eventsService.Service) *VideoProvisioner {
	return &VideoProvisioner{eventsService: events}
}

func (p *VideoProvisioner) HandleProvisionVideoMeetings(
	ctx context.Context,
	task *asynq.Task,
) error {
	var payload videoProvisionPayload
	if len(task.Payload()) > 0 {
		if err := json.Unmarshal(task.Payload(), &payload); err != nil {
			return fmt.Errorf("unmarshal payload: %w", err)
		}
	}

	start := time.Now()

	if payload.EventID != "" {
		log.Printf("[video-provisioner] provision single event=%s", payload.EventID)
		if err := p.provisionForEvent(ctx, payload.EventID); err != nil {
			return err
		}
	} else {
		log.Printf("[video-provisioner] sweep started")
		if err := p.sweep(ctx); err != nil {
			return err
		}
	}

	log.Printf("[video-provisioner] done in %v", time.Since(start))
	return nil
}

// provisionForEvent calls CreateEventMeeting on the events service.
// It's additive — schedules that already have a meeting are skipped,
// and only missing meetings are created.
//
// We load the event first so we can pass its creator as the "user"
// for the permission check. The worker has no authenticated user
// context, but the event's owner is always authorized to manage
// their own event's meetings.
func (p *VideoProvisioner) provisionForEvent(ctx context.Context, eventID string) error {
	event, err := p.eventsService.GetEventByID(ctx, eventID)
	if err != nil {
		return fmt.Errorf("load event: %w", err)
	}
	if event == nil {
		return fmt.Errorf("event not found: %s", eventID)
	}
	if event.CreatedBy == "" {
		return fmt.Errorf("event %s has no creator; cannot provision meetings", eventID)
	}

	_, err = p.eventsService.CreateEventMeeting(ctx, eventID, event.CreatedBy, "")
	return err
}

// sweep finds all events with unprovisioned virtual schedules within
// the next 30 days and provisions them, one event at a time.
//
// Runs daily via the Asynq scheduler (see cmd/api/main.go). Each
// event is provisioned sequentially to avoid hammering the video
// provider APIs. Failures are logged and skipped — the next sweep
// will retry them.
func (p *VideoProvisioner) sweep(ctx context.Context) error {
	eventIDs, err := p.eventsService.ListEventIDsWithUnprovisionedSchedules(ctx, 30)
	if err != nil {
		return fmt.Errorf("list unprovisioned events: %w", err)
	}

	log.Printf("[video-provisioner] sweep: %d events to provision", len(eventIDs))

	var succeeded, failed int
	for _, eventID := range eventIDs {
		if err := p.provisionForEvent(ctx, eventID); err != nil {
			log.Printf("[video-provisioner] sweep FAIL event=%s err=%v", eventID, err)
			failed++
			continue
		}
		succeeded++
	}

	log.Printf("[video-provisioner] sweep DONE ok=%d failed=%d", succeeded, failed)
	return nil
}

// Ensure the constant reference compiles even if we don't use it here.
var _ = eventsDomain.TaskProvisionVideoMeetings