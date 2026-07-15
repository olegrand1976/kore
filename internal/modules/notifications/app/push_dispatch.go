package app

import (
	"context"
	"log/slog"
	"strings"

	"github.com/google/uuid"
	notifpush "github.com/kore/kore/internal/modules/notifications/adapters/push"
	"github.com/kore/kore/internal/modules/notifications/domain"
	"github.com/kore/kore/internal/modules/notifications/ports"
	"github.com/kore/kore/pkg/kernel"
)

const pushBodyMaxLen = 200

func pushBodyForNotification(subject, body string) string {
	body = strings.TrimSpace(body)
	if cut := strings.Index(body, "\n--\n"); cut >= 0 {
		body = strings.TrimSpace(body[:cut])
	}
	if body == "" {
		return strings.TrimSpace(subject)
	}
	if len(body) > pushBodyMaxLen {
		return body[:pushBodyMaxLen-3] + "..."
	}
	return body
}

func (s *Service) dispatchPush(ctx context.Context, tenant kernel.TenantID, policy domain.RecipientPolicy, subject, body string) {
	if s.push == nil || !s.pushEnabled || s.devices == nil {
		return
	}
	userIDs, err := s.resolveTargetUserIDs(ctx, tenant, policy)
	if err != nil {
		slog.Default().WarnContext(ctx, "push: resolve users failed", "err", err)
		return
	}
	if len(userIDs) == 0 {
		return
	}
	subject = strings.TrimSpace(subject)
	body = strings.TrimSpace(body)
	if subject == "" && body == "" {
		return
	}
	if subject == "" {
		subject = "Kore"
	}
	if body == "" {
		body = subject
	}

	var tokens []string
	seen := map[string]struct{}{}
	for _, userID := range userIDs {
		devs, err := s.devices.ListDeviceTokens(ctx, tenant, userID)
		if err != nil {
			slog.Default().WarnContext(ctx, "push: list device tokens failed", "userId", userID, "err", err)
			continue
		}
		for _, d := range devs {
			if _, ok := seen[d.Token]; ok {
				continue
			}
			seen[d.Token] = struct{}{}
			tokens = append(tokens, d.Token)
		}
	}
	if len(tokens) == 0 {
		return
	}
	err = s.push.Send(ctx, tokens, ports.PushMessage{Title: subject, Body: body})
	if err == nil {
		return
	}
	for _, token := range notifpush.InvalidTokens(err) {
		if delErr := s.devices.DeleteDeviceTokenByValue(ctx, tenant, token); delErr != nil {
			slog.Default().WarnContext(ctx, "push: purge invalid token failed", "token", token, "err", delErr)
		}
	}
	if len(notifpush.InvalidTokens(err)) == 0 {
		slog.Default().WarnContext(ctx, "push: send failed", "err", err, "tokens", len(tokens))
	}
}

func (s *Service) resolveTargetUserIDs(ctx context.Context, tenant kernel.TenantID, policy domain.RecipientPolicy) ([]uuid.UUID, error) {
	if len(policy.UserIDs) > 0 {
		return policy.UserIDs, nil
	}
	if policy.EquipeID != nil {
		return s.resolver.ResolveEquipeUserIDs(ctx, tenant, *policy.EquipeID)
	}
	if policy.ApplicationID != nil {
		return s.resolver.ResolveApplicationUserIDs(ctx, tenant, *policy.ApplicationID)
	}
	if policy.ServiceID != nil {
		return s.resolver.ResolveServiceUserIDs(ctx, tenant, *policy.ServiceID)
	}
	return nil, nil
}
