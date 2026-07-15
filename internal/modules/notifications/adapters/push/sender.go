package push

import (
	"errors"
	"log/slog"
	"strings"

	"github.com/kore/kore/internal/modules/notifications/ports"
)

// InvalidTokens extrait les tokens FCM invalides d'une erreur d'envoi.
func InvalidTokens(err error) []string {
	var inv *invalidTokensError
	if errors.As(err, &inv) {
		return inv.InvalidTokens()
	}
	return nil
}

// NewFCMSender crée un expéditeur FCM HTTP v1 ou retourne le fallback si config incomplète.
func NewFCMSender(cfg FCMConfig, fallback ports.PushSender) ports.PushSender {
	if strings.TrimSpace(cfg.ProjectID) == "" {
		return fallback
	}
	client, err := NewFCMClient(cfg)
	if err != nil {
		slog.Default().Warn("fcm sender disabled", "err", err)
		return fallback
	}
	return client
}

var _ ports.PushSender = (*FCMClient)(nil)
