package app

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/kore/kore/internal/modules/integrations/ports"
)

const maxWebhookAttempts = 5

var webhookRetryDelays = []time.Duration{
	time.Minute,
	5 * time.Minute,
	30 * time.Minute,
	2 * time.Hour,
	24 * time.Hour,
}

type WebhookDispatcher struct {
	repo   ports.IntegrationRepository
	client *http.Client
}

func NewWebhookDispatcher(repo ports.IntegrationRepository) *WebhookDispatcher {
	return &WebhookDispatcher{
		repo:   repo,
		client: &http.Client{Timeout: 15 * time.Second},
	}
}

func (d *WebhookDispatcher) Dispatch(ctx context.Context, evt ports.OutboundEvent) error {
	subs, err := d.repo.ListWebhookSubscriptions(ctx, evt.TenantID)
	if err != nil {
		return err
	}
	payload, err := json.Marshal(map[string]any{
		"id":          evt.ID.String(),
		"type":        evt.Type,
		"tenant_id":   evt.TenantID.String(),
		"occurred_at": evt.OccurredAt.UTC().Format(time.RFC3339),
		"data":        evt.Data,
	})
	if err != nil {
		return err
	}
	for _, sub := range subs {
		if !sub.Active || !matchesEvent(sub.Events, evt.Type) {
			continue
		}
		if err := d.deliverWithRetry(ctx, sub.URL, sub.SecretRef, payload); err != nil {
			return err
		}
	}
	return nil
}

func (d *WebhookDispatcher) deliverWithRetry(ctx context.Context, url, secret string, payload []byte) error {
	var lastErr error
	for attempt := 0; attempt < maxWebhookAttempts; attempt++ {
		if attempt > 0 {
			delay := webhookRetryDelays[attempt-1]
			if delay > 0 {
				timer := time.NewTimer(delay)
				select {
				case <-ctx.Done():
					timer.Stop()
					return ctx.Err()
				case <-timer.C:
				}
			}
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Kore-Signature", signPayload(secret, payload))
		resp, err := d.client.Do(req)
		if err != nil {
			lastErr = err
			continue
		}
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		_ = resp.Body.Close()
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return nil
		}
		lastErr = fmt.Errorf("webhook delivery failed: status=%d body=%s", resp.StatusCode, string(body))
	}
	return lastErr
}

func signPayload(secret string, payload []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(payload)
	return hex.EncodeToString(mac.Sum(nil))
}

func matchesEvent(events []string, eventType string) bool {
	if len(events) == 0 {
		return true
	}
	for _, e := range events {
		if e == eventType || e == "*" {
			return true
		}
	}
	return false
}

var _ ports.WebhookDispatcher = (*WebhookDispatcher)(nil)
