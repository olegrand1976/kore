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

// Client appelle l'API Pennylane External v1 lorsque PENNYLANE_API_TOKEN est renseigné.
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

// SyncAccounting synchronise les factures clients du mois (API Pennylane customer_invoices).
func (c *Client) SyncAccounting(ctx context.Context, tenant kernel.TenantID, period string) (int, error) {
	if c.token == "" {
		return 0, fmt.Errorf("pennylane: API token not configured")
	}
	start, end, err := monthBounds(period)
	if err != nil {
		return 0, err
	}
	url := fmt.Sprintf("%s/customer_invoices?filter[date][gte]=%s&filter[date][lte]=%s&page=1&per_page=1",
		c.baseURL, start, end)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
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
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return 0, fmt.Errorf("pennylane http: status %d: %s", resp.StatusCode, string(body))
	}
	var payload struct {
		Total int `json:"total"`
		Meta  struct {
			Total int `json:"total"`
		} `json:"meta"`
		Data []json.RawMessage `json:"data"`
	}
	if len(body) > 0 {
		_ = json.Unmarshal(body, &payload)
	}
	total := payload.Total
	if total == 0 {
		total = payload.Meta.Total
	}
	if total == 0 && len(payload.Data) > 0 {
		total = len(payload.Data)
	}
	return total, nil
}

func monthBounds(period string) (start, end string, err error) {
	period = strings.TrimSpace(period)
	if len(period) != 7 || period[4] != '-' {
		return "", "", fmt.Errorf("pennylane: invalid period %q (expected YYYY-MM)", period)
	}
	start = period + "-01"
	t, err := time.Parse("2006-01-02", start)
	if err != nil {
		return "", "", err
	}
	end = t.AddDate(0, 1, -1).Format("2006-01-02")
	return start, end, nil
}
