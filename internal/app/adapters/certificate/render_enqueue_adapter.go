package certificate

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/hibiken/asynq"

	certdomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/certificate/domain"
)

const (
	TaskTypeRenderCertificate = "certificate:render"
	TaskQueueCertificate      = "certificates"
)

type RenderQueueAdapter struct {
	client *asynq.Client
}

func NewRenderQueueAdapter(client *asynq.Client) *RenderQueueAdapter {
	return &RenderQueueAdapter{client: client}
}

func (a *RenderQueueAdapter) EnqueueRender(
	ctx context.Context,
	job certdomain.CertificateRenderJob,
) error {
	payload, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("marshal render job: %w", err)
	}

	task := asynq.NewTask(TaskTypeRenderCertificate, payload)

	_, err = a.client.EnqueueContext(
		ctx,
		task,
		asynq.Queue(TaskQueueCertificate),
		asynq.TaskID(job.CertificateID),
		asynq.MaxRetry(5),
		asynq.Timeout(2*time.Minute),
	)
	if err != nil {
		// TaskID conflict = job already pending/running. Not an error.
		// asynq returns asynq.ErrTaskIDConflict for this case.
		if errors.Is(err, asynq.ErrTaskIDConflict) {
			return nil
		}
		return fmt.Errorf("enqueue render: %w", err)
	}
	return nil
}

var _ certdomain.RenderJobQueue = (*RenderQueueAdapter)(nil)