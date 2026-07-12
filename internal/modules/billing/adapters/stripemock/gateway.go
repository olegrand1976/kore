package stripemock

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/billing/domain"
	"github.com/kore/kore/internal/modules/billing/ports"
)

// Gateway is a dev/test PaymentGateway compatible with stripe-mock semantics.
// It works standalone without a running stripe-mock server.
type Gateway struct {
	WebhookSecret string
	BaseURL       string
	TrialDays     int
}

func New(webhookSecret string, trialDays int) *Gateway {
	return &Gateway{
		WebhookSecret: webhookSecret,
		BaseURL:       "http://localhost:12111",
		TrialDays:     trialDays,
	}
}

func (g *Gateway) CreateCheckoutSession(_ context.Context, req ports.CheckoutRequest) (domain.CheckoutSession, error) {
	id := "cs_mock_" + uuid.NewString()
	q := url.Values{}
	q.Set("tenant", req.TenantID.String())
	q.Set("seats", fmt.Sprintf("%d", req.Seats))
	if req.PrimaryColor != "" {
		q.Set("primary_color", req.PrimaryColor)
	}
	if req.BackgroundColor != "" {
		q.Set("background_color", req.BackgroundColor)
	}
	if req.LogoURL != "" {
		q.Set("logo_url", req.LogoURL)
	}
	checkoutURL := fmt.Sprintf("%s/v1/checkout/sessions/%s/mock_redirect?%s", g.BaseURL, id, q.Encode())
	return domain.CheckoutSession{ID: id, URL: checkoutURL}, nil
}

func (g *Gateway) CreatePortalSession(_ context.Context, customerID, returnURL string) (domain.PortalSession, error) {
	url := fmt.Sprintf("%s/v1/billing_portal/sessions/mock?customer=%s&return_url=%s",
		g.BaseURL, customerID, returnURL)
	return domain.PortalSession{URL: url}, nil
}

func (g *Gateway) VerifyWebhook(payload []byte, signature string) (domain.StripeEvent, error) {
	if g.WebhookSecret != "" && !g.verifySignature(payload, signature) {
		return domain.StripeEvent{}, domain.ErrInvalidWebhookSignature
	}
	var envelope struct {
		ID   string         `json:"id"`
		Type string         `json:"type"`
		Data map[string]any `json:"data"`
	}
	if err := json.Unmarshal(payload, &envelope); err != nil {
		return domain.StripeEvent{}, fmt.Errorf("invalid webhook payload: %w", err)
	}
	if envelope.ID == "" {
		envelope.ID = "evt_mock_" + uuid.NewString()
	}
	return domain.StripeEvent{
		ID:   envelope.ID,
		Type: envelope.Type,
		Data: envelope.Data,
	}, nil
}

func (g *Gateway) CancelSubscription(_ context.Context, subscriptionID string) error {
	if subscriptionID == "" {
		return fmt.Errorf("subscription id required")
	}
	return nil
}

func (g *Gateway) ListPrices(_ context.Context) (domain.PricingCatalog, error) {
	return domain.PricingCatalog{
		Currency:  "EUR",
		TrialDays: g.TrialDays,
		Modules: []domain.ModulePrice{
			{Code: domain.ModuleOrg, Name: "Organisation", Description: "Identité, tenant, RBAC", PriceID: "price_mock_org", UnitAmount: 1500, Interval: "month"},
			{Code: domain.ModuleCRA, Name: "CRA", Description: "Compte-rendu d'activité", PriceID: "price_mock_cra", UnitAmount: 1200, Interval: "month"},
			{Code: domain.ModuleConges, Name: "Congés", Description: "Gestion des absences", PriceID: "price_mock_conges", UnitAmount: 800, Interval: "month"},
			{Code: domain.ModuleBudget, Name: "Budget UO", Description: "Suivi budgétaire", PriceID: "price_mock_budget", UnitAmount: 1000, Interval: "month"},
			{Code: domain.ModuleTMA, Name: "TMA", Description: "Maintenance applicative", PriceID: "price_mock_tma", UnitAmount: 1800, Interval: "month"},
			{Code: domain.ModuleWorkflow, Name: "Workflow", Description: "Moteur de validation", PriceID: "price_mock_workflow", UnitAmount: 900, Interval: "month"},
		},
	}, nil
}

func (g *Gateway) verifySignature(payload []byte, signatureHeader string) bool {
	if signatureHeader == "" {
		return g.WebhookSecret == "whsec_test"
	}
	var timestamp, sigV1 string
	for _, part := range strings.Split(signatureHeader, ",") {
		kv := strings.SplitN(strings.TrimSpace(part), "=", 2)
		if len(kv) != 2 {
			continue
		}
		switch kv[0] {
		case "t":
			timestamp = kv[1]
		case "v1":
			sigV1 = kv[1]
		}
	}
	if sigV1 == "" {
		return false
	}
	signed := fmt.Sprintf("%s.%s", timestamp, string(payload))
	mac := hmac.New(sha256.New, []byte(g.WebhookSecret))
	_, _ = mac.Write([]byte(signed))
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(sigV1))
}

// BuildTestWebhookPayload builds a signed webhook payload for tests.
func BuildTestWebhookPayload(secret, eventType string, data map[string]any) ([]byte, string, error) {
	eventID := "evt_test_" + uuid.NewString()
	body, err := json.Marshal(map[string]any{
		"id":   eventID,
		"type": eventType,
		"data": map[string]any{"object": data},
	})
	if err != nil {
		return nil, "", err
	}
	ts := fmt.Sprintf("%d", time.Now().Unix())
	signed := fmt.Sprintf("%s.%s", ts, string(body))
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(signed))
	sig := hex.EncodeToString(mac.Sum(nil))
	header := fmt.Sprintf("t=%s,v1=%s", ts, sig)
	return body, header, nil
}

var _ ports.PaymentGateway = (*Gateway)(nil)
