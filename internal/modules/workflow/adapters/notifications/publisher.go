package notifications

import (
	"context"
	"fmt"

	notifports "github.com/kore/kore/internal/modules/notifications/ports"
	"github.com/kore/kore/internal/modules/workflow/domain"
	"github.com/kore/kore/internal/modules/workflow/ports"
)

type TransitionPublisher struct {
	notifier notifports.NotificationPublisher
}

func NewTransitionPublisher(notifier notifports.NotificationPublisher) ports.TransitionPublisher {
	return &TransitionPublisher{notifier: notifier}
}

func (p *TransitionPublisher) Publish(ctx context.Context, evt domain.TransitionOccurred) error {
	if p.notifier == nil {
		return nil
	}
	return p.notifier.Notify(ctx, notifports.NotificationEvent{
		TenantID: evt.TenantID,
		Trigger:  fmt.Sprintf("workflow.%s.%s", evt.DefinitionCode, evt.Action),
		Subject:  fmt.Sprintf("Transition %s → %s", evt.FromState, evt.ToState),
		Vars: map[string]string{
			"entityId": evt.EntityID,
			"action":   string(evt.Action),
		},
	})
}
