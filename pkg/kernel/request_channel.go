package kernel

import "context"

// RequestChannel identifies a demand routing module (TMA, support, maintenance).
type RequestChannel string

const (
	RequestChannelTMA         RequestChannel = "tma"
	RequestChannelSupport     RequestChannel = "support"
	RequestChannelMaintenance RequestChannel = "maintenance"
)

// RequestChannelReader exposes tenant-level request routing settings.
type RequestChannelReader interface {
	IsChannelEnabled(ctx context.Context, tenant TenantID, channel RequestChannel) (bool, error)
}

// InvoicingEnabledReader exposes the org-level invoicing module toggle.
type InvoicingEnabledReader interface {
	IsInvoicingEnabled(ctx context.Context, tenant TenantID) (bool, error)
}
