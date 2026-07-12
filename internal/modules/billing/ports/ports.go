package ports

import (
	"context"

	"github.com/kore/kore/internal/modules/billing/domain"
	"github.com/kore/kore/internal/platform/authx"
	"github.com/kore/kore/pkg/kernel"
)

type CheckoutCommand struct {
	TenantID      kernel.TenantID
	Modules       []domain.ModuleCode
	Seats         int
	SuccessURL    string
	CancelURL     string
	CustomerEmail string
}

type CheckoutRequest struct {
	TenantID      kernel.TenantID
	CustomerID    string
	Modules       []domain.ModuleCode
	Seats         int
	SuccessURL    string
	CancelURL     string
	CustomerEmail string
	// Stripe Checkout branding (module 14)
	PrimaryColor    string // #c9a227 Kore gold
	BackgroundColor string // #1a1f2e dark / #f8f9fb light
	LogoURL         string // public Kore horizontal logo URL
}

type SubscriptionRepository interface {
	Save(ctx context.Context, s domain.Subscription) error
	GetByTenant(ctx context.Context, tenantID kernel.TenantID) (domain.Subscription, error)
	GetByStripeCustomer(ctx context.Context, customerID string) (domain.Subscription, error)
	ListEntitlements(ctx context.Context, tenantID kernel.TenantID) ([]domain.ModuleEntitlement, error)
	SaveEntitlements(ctx context.Context, tenantID kernel.TenantID, modules []domain.ModuleEntitlement) error
	MarkEventProcessed(ctx context.Context, eventID, eventType string) (firstTime bool, err error)
}

type PaymentGateway interface {
	CreateCheckoutSession(ctx context.Context, req CheckoutRequest) (domain.CheckoutSession, error)
	CreatePortalSession(ctx context.Context, customerID, returnURL string) (domain.PortalSession, error)
	VerifyWebhook(payload []byte, signature string) (domain.StripeEvent, error)
	CancelSubscription(ctx context.Context, subscriptionID string) error
	ListPrices(ctx context.Context) (domain.PricingCatalog, error)
}

type SubscriptionService interface {
	StartCheckout(ctx context.Context, cmd CheckoutCommand) (domain.CheckoutSession, error)
	OpenCustomerPortal(ctx context.Context, tenantID kernel.TenantID, returnURL string) (domain.PortalSession, error)
	HandleWebhook(ctx context.Context, payload []byte, signature string) error
	Cancel(ctx context.Context, tenantID kernel.TenantID) error
	GetSubscription(ctx context.Context, tenantID kernel.TenantID) (domain.Subscription, error)
	Catalog(ctx context.Context) (domain.PricingCatalog, error)
}

type EntitlementReader interface {
	IsModuleEnabled(ctx context.Context, tenantID kernel.TenantID, module authx.Module) (bool, error)
	GetSeatLimit(ctx context.Context, tenantID kernel.TenantID) (int, error)
}
