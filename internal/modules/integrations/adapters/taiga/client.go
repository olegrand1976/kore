package taiga

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/kore/kore/internal/modules/integrations/domain"
	"github.com/kore/kore/internal/modules/integrations/ports"
)

type Client struct {
	apiBase  string
	username string
	password string
	client   *http.Client
}

type Config struct {
	BaseURL  string
	Username string
	Password string
	Timeout  time.Duration
}

func NewClient(cfg Config) *Client {
	base := strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/")
	if base != "" && !strings.HasSuffix(base, "/api/v1") {
		base += "/api/v1"
	}
	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	return &Client{
		apiBase:  base,
		username: strings.TrimSpace(cfg.Username),
		password: cfg.Password,
		client:   &http.Client{Timeout: timeout},
	}
}

func Enabled(cfg Config) bool {
	return strings.TrimSpace(cfg.BaseURL) != "" &&
		strings.TrimSpace(cfg.Username) != "" &&
		cfg.Password != ""
}

func (c *Client) ListProjects(ctx context.Context) ([]ports.TaigaProject, error) {
	if c.apiBase == "" || c.username == "" || c.password == "" {
		return nil, domain.ErrTaigaNotConfigured
	}
	token, err := c.auth(ctx)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.apiBase+"/projects?member=me", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrTaigaUnavailable, err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("%w: status %d", domain.ErrTaigaUnavailable, resp.StatusCode)
	}
	var raw []struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
		Slug string `json:"slug"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("taiga decode: %w", err)
	}
	out := make([]ports.TaigaProject, 0, len(raw))
	for _, p := range raw {
		if p.ID <= 0 {
			continue
		}
		out = append(out, ports.TaigaProject{
			ID:   p.ID,
			Name: strings.TrimSpace(p.Name),
			Slug: strings.TrimSpace(p.Slug),
		})
	}
	return out, nil
}

func (c *Client) auth(ctx context.Context) (string, error) {
	payload, err := json.Marshal(map[string]string{
		"type":     "normal",
		"username": c.username,
		"password": c.password,
	})
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.apiBase+"/auth", bytes.NewReader(payload))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("%w: %v", domain.ErrTaigaUnavailable, err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("%w: auth status %d", domain.ErrTaigaUnavailable, resp.StatusCode)
	}
	var parsed struct {
		AuthToken string `json:"auth_token"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return "", fmt.Errorf("taiga auth decode: %w", err)
	}
	if strings.TrimSpace(parsed.AuthToken) == "" {
		return "", domain.ErrTaigaUnavailable
	}
	return parsed.AuthToken, nil
}
