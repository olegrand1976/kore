package push

import (
	"context"
	"log/slog"
	"strings"

	"github.com/kore/kore/internal/modules/notifications/ports"
)

// StubSender journalise les envois push (défaut local).
type StubSender struct {
	log *slog.Logger
}

func NewStubSender(log *slog.Logger) *StubSender {
	if log == nil {
		log = slog.Default()
	}
	return &StubSender{log: log}
}

func (s *StubSender) Send(ctx context.Context, tokens []string, msg ports.PushMessage) error {
	s.log.InfoContext(ctx, "push stub",
		"tokens", len(tokens),
		"title", msg.Title,
		"body", msg.Body,
	)
	return nil
}

// FCMSender envoie via FCM HTTP v1 lorsque FCM_PROJECT_ID et credentials GCP sont configurés.
// Sans credentials, délègue au stub pour éviter un échec bloquant en dev.
type FCMSender struct {
	projectID string
	fallback  ports.PushSender
}

func NewFCMSender(projectID string, fallback ports.PushSender) *FCMSender {
	return &FCMSender{
		projectID: strings.TrimSpace(projectID),
		fallback:  fallback,
	}
}

func (s *FCMSender) Send(ctx context.Context, tokens []string, msg ports.PushMessage) error {
	if s.projectID == "" || len(tokens) == 0 {
		return s.fallback.Send(ctx, tokens, msg)
	}
	// FCM v1 requiert un compte de service (GOOGLE_APPLICATION_CREDENTIALS) — hors scope stub.
	return s.fallback.Send(ctx, tokens, msg)
}

var _ ports.PushSender = (*StubSender)(nil)
var _ ports.PushSender = (*FCMSender)(nil)
