package app

import (
	"strings"
	"time"

	"github.com/kore/kore/internal/modules/integrations/adapters/pennylane"
	"github.com/kore/kore/internal/platform/config"
)

func NewPennylaneClient(cfg config.Config) *pennylane.Client {
	return pennylane.NewClient(pennylane.Config{
		BaseURL: cfg.PennylaneAPIBaseURL,
		Token:   cfg.PennylaneAPIToken,
		Timeout: 30 * time.Second,
	})
}

func PennylaneEnabled(cfg config.Config) bool {
	return pennylane.Enabled(pennylane.Config{Token: cfg.PennylaneAPIToken}) &&
		strings.ToLower(strings.TrimSpace(cfg.PennylaneAPIToken)) != "stub"
}
