package pdp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/kore/kore/internal/modules/invoicing/domain"
	"github.com/kore/kore/internal/modules/invoicing/ports"
	"github.com/kore/kore/pkg/kernel"
)

// HTTPGateway transmet les factures EN 16931 vers une PDP/PA via API REST générique.
// Contrat attendu : POST {baseURL}/invoices → {"id":"..."} ; GET {baseURL}/invoices/{id} → {"status":"..."}.
type HTTPGateway struct {
	baseURL string
	apiKey  string
	client  *http.Client
}

type HTTPConfig struct {
	BaseURL string
	APIKey  string
	Timeout time.Duration
}

func NewHTTPGateway(cfg HTTPConfig) *HTTPGateway {
	base := strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/")
	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	return &HTTPGateway{
		baseURL: base,
		apiKey:  strings.TrimSpace(cfg.APIKey),
		client:  &http.Client{Timeout: timeout},
	}
}

func EnabledHTTP(cfg HTTPConfig) bool {
	return strings.TrimSpace(cfg.BaseURL) != "" && strings.TrimSpace(cfg.APIKey) != ""
}

func (g *HTTPGateway) Transmit(ctx context.Context, tenant kernel.TenantID, doc ports.En16931Document) (ports.PDPReceipt, error) {
	if doc == nil {
		return ports.PDPReceipt{}, fmt.Errorf("empty en16931 document")
	}
	body, err := json.Marshal(map[string]any{
		"tenant_id": tenant.String(),
		"document":  doc,
	})
	if err != nil {
		return ports.PDPReceipt{}, err
	}
	var out struct {
		ID string `json:"id"`
	}
	if err := g.doJSON(ctx, http.MethodPost, g.baseURL+"/invoices", body, &out); err != nil {
		return ports.PDPReceipt{}, err
	}
	if out.ID == "" {
		return ports.PDPReceipt{}, fmt.Errorf("pdp: empty receipt id")
	}
	return ports.PDPReceipt{ID: out.ID}, nil
}

func (g *HTTPGateway) SyncStatus(ctx context.Context, receiptID string) (domain.InvoiceStatus, error) {
	var out struct {
		Status string `json:"status"`
	}
	if err := g.doJSON(ctx, http.MethodGet, g.baseURL+"/invoices/"+receiptID, nil, &out); err != nil {
		return "", err
	}
	if out.Status == "" {
		return domain.InvoiceStatusTransmise, nil
	}
	return domain.InvoiceStatus(out.Status), nil
}

func (g *HTTPGateway) doJSON(ctx context.Context, method, url string, body []byte, dest any) error {
	var reader io.Reader
	if len(body) > 0 {
		reader = bytes.NewReader(body)
	}
	req, err := http.NewRequestWithContext(ctx, method, url, reader)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	if g.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+g.apiKey)
	}
	resp, err := g.client.Do(req)
	if err != nil {
		return fmt.Errorf("pdp http: %w", err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("pdp http: status %d: %s", resp.StatusCode, string(raw))
	}
	if dest == nil || len(raw) == 0 {
		return nil
	}
	return json.Unmarshal(raw, dest)
}

var _ ports.PDPGateway = (*HTTPGateway)(nil)
