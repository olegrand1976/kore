package push

import (
	"context"
	"log/slog"

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

var _ ports.PushSender = (*StubSender)(nil)
