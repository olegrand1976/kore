package app

import (
	"strings"

	"github.com/kore/kore/internal/modules/invoicing/adapters/pdp"
	"github.com/kore/kore/internal/modules/invoicing/ports"
	"github.com/kore/kore/internal/platform/config"
)

// NewPDPGateway sélectionne le connecteur PDP selon la configuration (stub par défaut).
func NewPDPGateway(cfg config.Config) ports.PDPGateway {
	httpCfg := pdp.HTTPConfig{
		BaseURL: cfg.PDPBaseURL,
		APIKey:  cfg.PDPAPIKey,
		Timeout: cfg.PDPTimeout,
	}
	provider := strings.ToLower(strings.TrimSpace(cfg.PDPProvider))
	if provider == "http" && pdp.EnabledHTTP(httpCfg) {
		return pdp.NewHTTPGateway(httpCfg)
	}
	return pdp.NewStubGateway()
}
