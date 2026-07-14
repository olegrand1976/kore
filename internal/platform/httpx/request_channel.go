package httpx

import (
	"context"
	"net/http"

	"github.com/kore/kore/internal/platform/authx"
	"github.com/kore/kore/pkg/kernel"
)

// RequireRequestChannel blocks mutations when the tenant disabled the channel.
func RequireRequestChannel(w http.ResponseWriter, r *http.Request, reader kernel.RequestChannelReader, channel kernel.RequestChannel) bool {
	if reader == nil {
		return true
	}
	identity, ok := authx.FromContext(r.Context())
	if !ok {
		WriteError(w, http.StatusUnauthorized, ErrCodeUnauthorized, "unauthorized")
		return false
	}
	enabled, err := reader.IsChannelEnabled(r.Context(), identity.TenantID, channel)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, ErrCodeInternal, err.Error())
		return false
	}
	if !enabled {
		WriteError(w, http.StatusForbidden, ErrCodeForbidden, "request channel disabled")
		return false
	}
	return true
}

// ChannelEnabled is a test helper for handlers.
func ChannelEnabled(ctx context.Context, reader kernel.RequestChannelReader, tenant kernel.TenantID, channel kernel.RequestChannel) (bool, error) {
	if reader == nil {
		return true, nil
	}
	return reader.IsChannelEnabled(ctx, tenant, channel)
}
