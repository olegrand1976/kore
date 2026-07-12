package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/kore/kore/pkg/kernel"
)

var (
	ErrSubscriptionNotFound    = errors.New("subscription not found")
	ErrInvalidWebhookSignature = errors.New("invalid webhook signature")
	ErrPaymentRequired         = errors.New("payment required")
	ErrModuleNotSubscribed     = errors.New("module not subscribed")
)

type SubscriptionStatus string

const (
	StatusTrial     SubscriptionStatus = "trial"
	StatusActive    SubscriptionStatus = "active"
	StatusPastDue   SubscriptionStatus = "past_due"
	StatusSuspended SubscriptionStatus = "suspended"
	StatusCanceled  SubscriptionStatus = "canceled"
)

type ModuleCode string

const (
	ModuleOrg           ModuleCode = "org"
	ModuleCRA           ModuleCode = "cra"
	ModuleConges        ModuleCode = "conges"
	ModuleBudget        ModuleCode = "budget"
	ModuleTMA           ModuleCode = "tma"
	ModuleWorkflow      ModuleCode = "workflow"
	ModuleNotifications ModuleCode = "notifications"
	ModuleBilling       ModuleCode = "billing"
)

type Subscription struct {
	ID                   uuid.UUID
	TenantID             kernel.TenantID
	StripeCustomerID     string
	StripeSubscriptionID string
	Status               SubscriptionStatus
	Seats                int
	CurrentPeriodEnd     *time.Time
	Modules              []ModuleEntitlement
}

type ModuleEntitlement struct {
	TenantID   kernel.TenantID
	ModuleCode ModuleCode
	Enabled    bool
}

type PricingCatalog struct {
	Currency string        `json:"currency"`
	Modules  []ModulePrice `json:"modules"`
	TrialDays int          `json:"trialDays"`
}

type ModulePrice struct {
	Code        ModuleCode `json:"code"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	PriceID     string     `json:"priceId"`
	UnitAmount  int64      `json:"unitAmount"`
	Interval    string     `json:"interval"`
}

type CheckoutSession struct {
	ID  string `json:"id"`
	URL string `json:"url"`
}

type PortalSession struct {
	URL string `json:"url"`
}

type StripeEvent struct {
	ID   string
	Type string
	Data map[string]any
}

func (s SubscriptionStatus) AllowsAccess() bool {
	switch s {
	case StatusTrial, StatusActive, StatusPastDue:
		return true
	default:
		return false
	}
}

func (s Subscription) IsModuleEnabled(code ModuleCode) bool {
	if !s.Status.AllowsAccess() {
		return false
	}
	for _, m := range s.Modules {
		if m.ModuleCode == code && m.Enabled {
			return true
		}
	}
	return false
}
