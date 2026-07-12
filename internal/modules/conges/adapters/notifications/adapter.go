package notifications

import (
	"context"

	notifports "github.com/kore/kore/internal/modules/notifications/ports"
	"github.com/kore/kore/internal/modules/conges/ports"
)

type PublisherAdapter struct {
	notifier notifports.NotificationPublisher
}

func NewPublisherAdapter(notifier notifports.NotificationPublisher) ports.NotificationPublisher {
	return &PublisherAdapter{notifier: notifier}
}

func (a *PublisherAdapter) Notify(ctx context.Context, evt ports.NotificationEvent) error {
	return a.notifier.Notify(ctx, notifports.NotificationEvent{
		TenantID: evt.TenantID,
		Trigger:  "conges.notification",
		Subject:  evt.Subject,
		Vars: map[string]string{
			"body":   evt.Body,
			"userId": evt.UserID.String(),
		},
	})
}

var _ ports.NotificationPublisher = (*PublisherAdapter)(nil)
