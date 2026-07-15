package app

import (
	"log/slog"
	"os"
	"strings"

	notifpush "github.com/kore/kore/internal/modules/notifications/adapters/push"
	"github.com/kore/kore/internal/modules/notifications/ports"
	"github.com/kore/kore/internal/platform/config"
)

func NewPushSender(cfg config.Config) ports.PushSender {
	stub := notifpush.NewStubSender(slog.Default())
	if !cfg.PushEnabled {
		return stub
	}
	credsPath := cfg.FCMCredentialsPath
	if credsPath == "" {
		credsPath = strings.TrimSpace(os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"))
	}
	return notifpush.NewFCMSender(notifpush.FCMConfig{
		ProjectID:       cfg.FCMProjectID,
		CredentialsPath: credsPath,
	}, stub)
}
