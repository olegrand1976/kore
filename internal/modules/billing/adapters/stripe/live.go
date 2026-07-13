package stripe

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/kore/kore/internal/modules/billing/domain"
	"github.com/kore/kore/internal/modules/billing/ports"
	"github.com/stripe/stripe-go/v81"
	billingportal "github.com/stripe/stripe-go/v81/billingportal/session"
	checkoutsession "github.com/stripe/stripe-go/v81/checkout/session"
	"github.com/stripe/stripe-go/v81/customer"
	"github.com/stripe/stripe-go/v81/price"
	"github.com/stripe/stripe-go/v81/subscription"
	"github.com/stripe/stripe-go/v81/webhook"
)

// Gateway implements ports.PaymentGateway against the live Stripe API.
type Gateway struct {
	WebhookSecret string
	TrialDays     int
}

// IsLiveKey reports whether the key looks like a live Stripe secret key.
func IsLiveKey(key string) bool {
	return strings.HasPrefix(key, "sk_live_")
}

// Enabled reports whether live Stripe should be used (STRIPE_LIVE=true or live secret key).
func Enabled(secretKey string) bool {
	if os.Getenv("STRIPE_LIVE") == "true" {
		return true
	}
	return IsLiveKey(secretKey)
}

func New(secretKey, webhookSecret string, trialDays int) *Gateway {
	if secretKey == "" {
		secretKey = os.Getenv("STRIPE_SECRET_KEY")
	}
	if secretKey == "" {
		secretKey = os.Getenv("STRIPE_API_KEY")
	}
	stripe.Key = secretKey
	return &Gateway{WebhookSecret: webhookSecret, TrialDays: trialDays}
}

func (g *Gateway) CreateCheckoutSession(ctx context.Context, req ports.CheckoutRequest) (domain.CheckoutSession, error) {
	_ = ctx
	customerID := req.CustomerID
	if customerID == "" && req.CustomerEmail != "" {
		c, err := customer.New(&stripe.CustomerParams{Email: stripe.String(req.CustomerEmail)})
		if err != nil {
			return domain.CheckoutSession{}, fmt.Errorf("stripe create customer: %w", err)
		}
		customerID = c.ID
	}
	params := &stripe.CheckoutSessionParams{
		Mode:              stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		ClientReferenceID: stripe.String(req.TenantID.String()),
		SuccessURL:        stripe.String(req.SuccessURL),
		CancelURL:         stripe.String(req.CancelURL),
		LineItems:         buildLineItems(req.Modules),
		Metadata: map[string]string{
			"tenant_id": req.TenantID.String(),
			"seats":     fmt.Sprintf("%d", maxSeats(req.Seats)),
			"modules":   modulesJSON(req.Modules),
		},
	}
	if customerID != "" {
		params.Customer = stripe.String(customerID)
	}
	if g.TrialDays > 0 {
		params.SubscriptionData = &stripe.CheckoutSessionSubscriptionDataParams{
			TrialPeriodDays: stripe.Int64(int64(g.TrialDays)),
		}
	}
	sess, err := checkoutsession.New(params)
	if err != nil {
		return domain.CheckoutSession{}, fmt.Errorf("stripe checkout session: %w", err)
	}
	return domain.CheckoutSession{ID: sess.ID, URL: sess.URL}, nil
}

func (g *Gateway) CreatePortalSession(_ context.Context, customerID, returnURL string) (domain.PortalSession, error) {
	sess, err := billingportal.New(&stripe.BillingPortalSessionParams{
		Customer:  stripe.String(customerID),
		ReturnURL: stripe.String(returnURL),
	})
	if err != nil {
		return domain.PortalSession{}, fmt.Errorf("stripe portal session: %w", err)
	}
	return domain.PortalSession{URL: sess.URL}, nil
}

func (g *Gateway) VerifyWebhook(payload []byte, signature string) (domain.StripeEvent, error) {
	event, err := webhook.ConstructEvent(payload, signature, g.WebhookSecret)
	if err != nil {
		return domain.StripeEvent{}, domain.ErrInvalidWebhookSignature
	}
	var data map[string]any
	if err := json.Unmarshal(event.Data.Raw, &data); err != nil {
		data = map[string]any{"object": json.RawMessage(event.Data.Raw)}
	}
	return domain.StripeEvent{
		ID:   event.ID,
		Type: string(event.Type),
		Data: data,
	}, nil
}

func (g *Gateway) CancelSubscription(_ context.Context, subscriptionID string) error {
	_, err := subscription.Cancel(subscriptionID, nil)
	if err != nil {
		return fmt.Errorf("stripe cancel subscription: %w", err)
	}
	return nil
}

func (g *Gateway) ListPrices(_ context.Context) (domain.PricingCatalog, error) {
	iter := price.List(&stripe.PriceListParams{Active: stripe.Bool(true), Expand: []*string{stripe.String("data.product")}})
	catalog := domain.PricingCatalog{Currency: "EUR", TrialDays: g.TrialDays}
	for iter.Next() {
		p := iter.Price()
		if p.UnitAmount == 0 {
			continue
		}
		name := p.Nickname
		if name == "" && p.Product != nil {
			name = p.Product.Name
		}
		catalog.Modules = append(catalog.Modules, domain.ModulePrice{
			Code:        domain.ModuleCode(p.LookupKey),
			Name:        name,
			PriceID:     p.ID,
			UnitAmount:  p.UnitAmount,
			Interval:    string(p.Recurring.Interval),
			Description: p.Nickname,
		})
	}
	if err := iter.Err(); err != nil {
		return domain.PricingCatalog{}, fmt.Errorf("stripe list prices: %w", err)
	}
	if len(catalog.Modules) == 0 {
		return fallbackCatalog(g.TrialDays), nil
	}
	return catalog, nil
}

func buildLineItems(modules []domain.ModuleCode) []*stripe.CheckoutSessionLineItemParams {
	if len(modules) == 0 {
		modules = []domain.ModuleCode{domain.ModuleOrg, domain.ModuleCRA, domain.ModuleConges, domain.ModuleBudget}
	}
	items := make([]*stripe.CheckoutSessionLineItemParams, 0, len(modules))
	for _, mod := range modules {
		priceID := os.Getenv("STRIPE_PRICE_" + strings.ToUpper(string(mod)))
		if priceID == "" {
			continue
		}
		items = append(items, &stripe.CheckoutSessionLineItemParams{
			Price:    stripe.String(priceID),
			Quantity: stripe.Int64(1),
		})
	}
	if len(items) == 0 {
		if defaultPrice := os.Getenv("STRIPE_DEFAULT_PRICE_ID"); defaultPrice != "" {
			items = append(items, &stripe.CheckoutSessionLineItemParams{
				Price:    stripe.String(defaultPrice),
				Quantity: stripe.Int64(1),
			})
		}
	}
	return items
}

func modulesJSON(modules []domain.ModuleCode) string {
	codes := make([]string, 0, len(modules))
	for _, m := range modules {
		codes = append(codes, string(m))
	}
	raw, _ := json.Marshal(codes)
	return string(raw)
}

func maxSeats(seats int) int {
	if seats <= 0 {
		return 1
	}
	return seats
}

func fallbackCatalog(trialDays int) domain.PricingCatalog {
	return domain.PricingCatalog{
		Currency:  "EUR",
		TrialDays: trialDays,
		Modules: []domain.ModulePrice{
			{Code: domain.ModuleOrg, Name: "Organisation", PriceID: "price_live_org", UnitAmount: 1500, Interval: "month"},
			{Code: domain.ModuleCRA, Name: "CRA", PriceID: "price_live_cra", UnitAmount: 1200, Interval: "month"},
			{Code: domain.ModuleConges, Name: "Congés", PriceID: "price_live_conges", UnitAmount: 800, Interval: "month"},
			{Code: domain.ModuleBudget, Name: "Budget UO", PriceID: "price_live_budget", UnitAmount: 1000, Interval: "month"},
			{Code: domain.ModuleTMA, Name: "TMA", PriceID: "price_live_tma", UnitAmount: 1800, Interval: "month"},
		},
	}
}

var _ ports.PaymentGateway = (*Gateway)(nil)
