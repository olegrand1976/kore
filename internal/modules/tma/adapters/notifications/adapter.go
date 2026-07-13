package notifications

import (
	"context"

	"github.com/google/uuid"
	notifports "github.com/kore/kore/internal/modules/notifications/ports"
	"github.com/kore/kore/internal/modules/tma/ports"
)

type PublisherAdapter struct {
	notifier notifports.NotificationPublisher
}

func NewPublisherAdapter(notifier notifports.NotificationPublisher) ports.NotificationPublisher {
	return &PublisherAdapter{notifier: notifier}
}

func (a *PublisherAdapter) Notify(ctx context.Context, evt ports.NotificationEvent) error {
	trigger := evt.Trigger
	if trigger == "" {
		trigger = "tma.notification"
	}
	vars := map[string]string{}
	for k, v := range evt.Vars {
		vars[k] = v
	}
	if evt.Body != "" {
		vars["body"] = evt.Body
	}
	if evt.UserID != uuid.Nil {
		vars["userId"] = evt.UserID.String()
	}
	return a.notifier.Notify(ctx, notifports.NotificationEvent{
		TenantID: evt.TenantID,
		Trigger:  trigger,
		Subject:  evt.Subject,
		Vars:     vars,
	})
}

var _ ports.NotificationPublisher = (*PublisherAdapter)(nil)
