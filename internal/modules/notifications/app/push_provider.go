package app

import (
	"log/slog"

	notifpush "github.com/kore/kore/internal/modules/notifications/adapters/push"
	"github.com/kore/kore/internal/modules/notifications/ports"
	"github.com/kore/kore/internal/platform/config"
)

func NewPushSender(cfg config.Config) ports.PushSender {
	stub := notifpush.NewStubSender(slog.Default())
	if !cfg.PushEnabled {
		return stub
	}
	return notifpush.NewFCMSender(cfg.FCMProjectID, stub)
}
