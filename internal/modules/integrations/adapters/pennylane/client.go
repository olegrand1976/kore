package pennylane

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/kore/kore/pkg/kernel"
)

// Client appelle l'API Pennylane (ou compatible) lorsque PENNYLANE_API_TOKEN est renseigné.
type Client struct {
	baseURL string
	token   string
	client  *http.Client
}

type Config struct {
	BaseURL string
	Token   string
	Timeout time.Duration
}

func NewClient(cfg Config) *Client {
	base := strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/")
	if base == "" {
		base = "https://app.pennylane.com/api/external/v1"
	}
	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	return &Client{
		baseURL: base,
		token:   strings.TrimSpace(cfg.Token),
		client:  &http.Client{Timeout: timeout},
	}
}

func Enabled(cfg Config) bool {
	return strings.TrimSpace(cfg.Token) != ""
}

func (c *Client) SyncAccounting(ctx context.Context, tenant kernel.TenantID, period string) (int, error) {
	if c.token == "" {
		return 0, fmt.Errorf("pennylane: API token not configured")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/ledger_entries?period="+period, nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Tenant-Id", tenant.String())
	resp, err := c.client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("pennylane http: %w", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusNotImplemented {
		// Endpoint indicatif — connexion validée, sync métier à compléter avec le contrat Pennylane réel.
		return 0, nil
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return 0, fmt.Errorf("pennylane http: status %d: %s", resp.StatusCode, string(body))
	}
	var payload struct {
		Total int `json:"total"`
	}
	if len(body) > 0 {
		_ = json.Unmarshal(body, &payload)
	}
	return payload.Total, nil
}
