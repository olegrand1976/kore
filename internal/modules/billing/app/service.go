package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	stripegw "github.com/kore/kore/internal/modules/billing/adapters/stripe"
	"github.com/kore/kore/internal/modules/billing/adapters/stripemock"
	"github.com/kore/kore/internal/modules/billing/domain"
	"github.com/kore/kore/internal/modules/billing/ports"
	"github.com/kore/kore/internal/platform/authx"
	"github.com/kore/kore/pkg/kernel"
)

type Service struct {
	repo      ports.SubscriptionRepository
	gateway   ports.PaymentGateway
	trialDays int
}

func NewService(repo ports.SubscriptionRepository, stripeSecretKey, webhookSecret string, trialDays int) *Service {
	if webhookSecret == "" {
		webhookSecret = "whsec_test"
	}
	var gateway ports.PaymentGateway
	if stripegw.Enabled(stripeSecretKey) {
		gateway = stripegw.New(stripeSecretKey, webhookSecret, trialDays)
	} else {
		gateway = stripemock.New(webhookSecret, trialDays)
	}
	return &Service{
		repo:      repo,
		gateway:   gateway,
		trialDays: trialDays,
	}
}

func NewServiceWithGateway(repo ports.SubscriptionRepository, gateway ports.PaymentGateway, trialDays int) *Service {
	return &Service{repo: repo, gateway: gateway, trialDays: trialDays}
}

func (s *Service) IsModuleEnabled(ctx context.Context, tenantID kernel.TenantID, module authx.Module) (bool, error) {
	sub, err := s.repo.GetByTenant(ctx, tenantID)
	if err != nil {
		if errors.Is(err, domain.ErrSubscriptionNotFound) {
			return true, nil
		}
		return false, err
	}
	if !sub.Status.AllowsAccess() {
		return false, nil
	}
	if len(sub.Modules) == 0 {
		return true, nil
	}
	return sub.IsModuleEnabled(domain.ModuleCode(module)), nil
}

func (s *Service) GetSeatLimit(ctx context.Context, tenantID kernel.TenantID) (int, error) {
	sub, err := s.repo.GetByTenant(ctx, tenantID)
	if err != nil {
		if errors.Is(err, domain.ErrSubscriptionNotFound) {
			return 0, nil
		}
		return 0, err
	}
	if !sub.Status.AllowsAccess() {
		return 0, nil
	}
	return sub.Seats, nil
}

func (s *Service) Catalog(ctx context.Context) (domain.PricingCatalog, error) {
	catalog, err := s.gateway.ListPrices(ctx)
	if err != nil {
		return domain.PricingCatalog{}, err
	}
	if catalog.TrialDays == 0 {
		catalog.TrialDays = s.trialDays
	}
	return catalog, nil
}

func (s *Service) GetSubscription(ctx context.Context, tenantID kernel.TenantID) (domain.Subscription, error) {
	return s.repo.GetByTenant(ctx, tenantID)
}

func (s *Service) StartCheckout(ctx context.Context, cmd ports.CheckoutCommand) (domain.CheckoutSession, error) {
	sub, err := s.repo.GetByTenant(ctx, cmd.TenantID)
	customerID := ""
	if err == nil {
		customerID = sub.StripeCustomerID
	} else if !errors.Is(err, domain.ErrSubscriptionNotFound) {
		return domain.CheckoutSession{}, err
	}
	seats := cmd.Seats
	if seats <= 0 {
		seats = 1
	}
	return s.gateway.CreateCheckoutSession(ctx, ports.CheckoutRequest{
		TenantID:        cmd.TenantID,
		CustomerID:      customerID,
		Modules:         cmd.Modules,
		Seats:           seats,
		SuccessURL:      cmd.SuccessURL,
		CancelURL:       cmd.CancelURL,
		CustomerEmail:   cmd.CustomerEmail,
		PrimaryColor:    "#c9a227",
		BackgroundColor: "#1a1f2e",
		LogoURL:         "/brand/kore-logo-horizontal.svg",
	})
}

func (s *Service) OpenCustomerPortal(ctx context.Context, tenantID kernel.TenantID, returnURL string) (domain.PortalSession, error) {
	sub, err := s.repo.GetByTenant(ctx, tenantID)
	if err != nil {
		return domain.PortalSession{}, err
	}
	if sub.StripeCustomerID == "" {
		return domain.PortalSession{}, fmt.Errorf("no stripe customer for tenant")
	}
	return s.gateway.CreatePortalSession(ctx, sub.StripeCustomerID, returnURL)
}

func (s *Service) Cancel(ctx context.Context, tenantID kernel.TenantID) error {
	sub, err := s.repo.GetByTenant(ctx, tenantID)
	if err != nil {
		return err
	}
	if sub.StripeSubscriptionID != "" {
		if err := s.gateway.CancelSubscription(ctx, sub.StripeSubscriptionID); err != nil {
			return err
		}
	}
	sub.Status = domain.StatusCanceled
	return s.repo.Save(ctx, sub)
}

func (s *Service) HandleWebhook(ctx context.Context, payload []byte, signature string) error {
	event, err := s.gateway.VerifyWebhook(payload, signature)
	if err != nil {
		return err
	}
	firstTime, err := s.repo.MarkEventProcessed(ctx, event.ID, event.Type)
	if err != nil {
		return err
	}
	if !firstTime {
		return nil
	}
	switch event.Type {
	case "checkout.session.completed":
		return s.handleCheckoutCompleted(ctx, event)
	case "customer.subscription.updated":
		return s.handleSubscriptionUpdated(ctx, event)
	case "customer.subscription.deleted":
		return s.handleSubscriptionDeleted(ctx, event)
	case "invoice.payment_failed":
		return s.handlePaymentFailed(ctx, event)
	default:
		return nil
	}
}

func (s *Service) handleCheckoutCompleted(ctx context.Context, event domain.StripeEvent) error {
	obj := objectFromEvent(event)
	tenantRaw := stringField(obj, "client_reference_id")
	if tenantRaw == "" {
		tenantRaw = stringField(obj, "metadata", "tenant_id")
	}
	tenantID, err := kernel.ParseTenantID(tenantRaw)
	if err != nil {
		return fmt.Errorf("webhook missing tenant_id: %w", err)
	}
	customerID := stringField(obj, "customer")
	subscriptionID := stringField(obj, "subscription")
	seats := intField(obj, "metadata", "seats")
	if seats <= 0 {
		seats = 1
	}
	end := time.Now().UTC().Add(30 * 24 * time.Hour)
	sub := domain.Subscription{
		ID:                   uuid.New(),
		TenantID:             tenantID,
		StripeCustomerID:     customerID,
		StripeSubscriptionID: subscriptionID,
		Status:               domain.StatusActive,
		Seats:                seats,
		CurrentPeriodEnd:     &end,
	}
	if err := s.repo.Save(ctx, sub); err != nil {
		return err
	}
	modules := parseModulesFromMetadata(obj)
	if len(modules) == 0 {
		modules = defaultModules()
	}
	return s.repo.SaveEntitlements(ctx, tenantID, modules)
}

func (s *Service) handleSubscriptionUpdated(ctx context.Context, event domain.StripeEvent) error {
	obj := objectFromEvent(event)
	customerID := stringField(obj, "customer")
	sub, err := s.repo.GetByStripeCustomer(ctx, customerID)
	if err != nil {
		return err
	}
	status := mapStripeStatus(stringField(obj, "status"))
	sub.Status = status
	sub.StripeSubscriptionID = stringField(obj, "id")
	if qty := intField(obj, "items", "data", "0", "quantity"); qty > 0 {
		sub.Seats = qty
	}
	if endUnix := int64Field(obj, "current_period_end"); endUnix > 0 {
		end := time.Unix(endUnix, 0).UTC()
		sub.CurrentPeriodEnd = &end
	}
	return s.repo.Save(ctx, sub)
}

func (s *Service) handleSubscriptionDeleted(ctx context.Context, event domain.StripeEvent) error {
	obj := objectFromEvent(event)
	customerID := stringField(obj, "customer")
	sub, err := s.repo.GetByStripeCustomer(ctx, customerID)
	if err != nil {
		return err
	}
	sub.Status = domain.StatusCanceled
	return s.repo.Save(ctx, sub)
}

func (s *Service) handlePaymentFailed(ctx context.Context, event domain.StripeEvent) error {
	obj := objectFromEvent(event)
	customerID := stringField(obj, "customer")
	sub, err := s.repo.GetByStripeCustomer(ctx, customerID)
	if err != nil {
		return nil
	}
	sub.Status = domain.StatusPastDue
	return s.repo.Save(ctx, sub)
}

func objectFromEvent(event domain.StripeEvent) map[string]any {
	if obj, ok := event.Data["object"].(map[string]any); ok {
		return obj
	}
	return event.Data
}

func stringField(obj map[string]any, path ...string) string {
	cur := any(obj)
	for _, p := range path {
		m, ok := cur.(map[string]any)
		if !ok {
			return ""
		}
		cur = m[p]
	}
	if cur == nil {
		return ""
	}
	return fmt.Sprint(cur)
}

func intField(obj map[string]any, path ...string) int {
	s := stringField(obj, path...)
	if s == "" {
		return 0
	}
	var n int
	_, _ = fmt.Sscan(s, &n)
	return n
}

func int64Field(obj map[string]any, path ...string) int64 {
	s := stringField(obj, path...)
	if s == "" {
		return 0
	}
	var n int64
	_, _ = fmt.Sscan(s, &n)
	return n
}

func mapStripeStatus(stripeStatus string) domain.SubscriptionStatus {
	switch strings.ToLower(stripeStatus) {
	case "trialing":
		return domain.StatusTrial
	case "active":
		return domain.StatusActive
	case "past_due":
		return domain.StatusPastDue
	case "canceled", "unpaid":
		return domain.StatusCanceled
	case "paused":
		return domain.StatusSuspended
	default:
		return domain.StatusActive
	}
}

func parseModulesFromMetadata(obj map[string]any) []domain.ModuleEntitlement {
	raw := stringField(obj, "metadata", "modules")
	if raw == "" {
		return nil
	}
	var codes []string
	if err := json.Unmarshal([]byte(raw), &codes); err != nil {
		codes = strings.Split(raw, ",")
	}
	tenantRaw := stringField(obj, "client_reference_id")
	if tenantRaw == "" {
		tenantRaw = stringField(obj, "metadata", "tenant_id")
	}
	tenantID, err := kernel.ParseTenantID(tenantRaw)
	if err != nil {
		return nil
	}
	out := make([]domain.ModuleEntitlement, 0, len(codes))
	for _, code := range codes {
		code = strings.TrimSpace(code)
		if code == "" {
			continue
		}
		out = append(out, domain.ModuleEntitlement{
			TenantID:   tenantID,
			ModuleCode: domain.ModuleCode(code),
			Enabled:    true,
		})
	}
	return out
}

func defaultModules() []domain.ModuleEntitlement {
	codes := []domain.ModuleCode{
		domain.ModuleOrg, domain.ModuleCRA, domain.ModuleConges,
		domain.ModuleBudget, domain.ModuleTMA, domain.ModuleWorkflow,
	}
	out := make([]domain.ModuleEntitlement, 0, len(codes))
	for _, code := range codes {
		out = append(out, domain.ModuleEntitlement{
			ModuleCode: code,
			Enabled:    true,
		})
	}
	return out
}

var (
	_ ports.SubscriptionService = (*Service)(nil)
	_ ports.EntitlementReader   = (*Service)(nil)
	_ authx.EntitlementReader   = (*Service)(nil)
)
