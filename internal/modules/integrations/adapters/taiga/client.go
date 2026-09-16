package taiga

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
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

func (c *Client) ensureConfigured() error {
	if c.apiBase == "" || c.username == "" || c.password == "" {
		return domain.ErrTaigaNotConfigured
	}
	return nil
}

func (c *Client) ListProjects(ctx context.Context) ([]ports.TaigaProject, error) {
	if err := c.ensureConfigured(); err != nil {
		return nil, err
	}
	body, err := c.doJSON(ctx, http.MethodGet, "/projects", nil)
	if err != nil {
		return nil, err
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

func (c *Client) CreateIssue(
	ctx context.Context,
	projectID int,
	subject, description string,
	externalRef []string,
) (ports.TaigaIssue, error) {
	if err := c.ensureConfigured(); err != nil {
		return ports.TaigaIssue{}, err
	}
	if projectID <= 0 {
		return ports.TaigaIssue{}, domain.ErrTaigaProjectNotFound
	}
	payload := map[string]any{
		"project":     projectID,
		"subject":     strings.TrimSpace(subject),
		"description": strings.TrimSpace(description),
	}
	if len(externalRef) > 0 {
		payload["external_reference"] = externalRef
	}
	body, err := c.doJSON(ctx, http.MethodPost, "/issues", payload)
	if err != nil {
		return ports.TaigaIssue{}, err
	}
	return decodeIssue(body)
}

func (c *Client) ListProjectIssues(ctx context.Context, projectID int) ([]ports.TaigaIssue, error) {
	if err := c.ensureConfigured(); err != nil {
		return nil, err
	}
	if projectID <= 0 {
		return nil, domain.ErrTaigaProjectNotFound
	}
	path := "/issues?" + url.Values{"project": {strconv.Itoa(projectID)}}.Encode()
	body, err := c.doJSON(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	var raw []json.RawMessage
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("taiga decode issues: %w", err)
	}
	out := make([]ports.TaigaIssue, 0, len(raw))
	for _, item := range raw {
		issue, err := decodeIssue(item)
		if err != nil {
			continue
		}
		if issue.ID <= 0 {
			continue
		}
		out = append(out, issue)
	}
	return out, nil
}

func (c *Client) UpdateIssueExternalReference(
	ctx context.Context,
	issueID int,
	version int,
	externalRef []string,
) (ports.TaigaIssue, error) {
	if err := c.ensureConfigured(); err != nil {
		return ports.TaigaIssue{}, err
	}
	if issueID <= 0 {
		return ports.TaigaIssue{}, domain.ErrTaigaUnavailable
	}
	payload := map[string]any{
		"version":            version,
		"external_reference": externalRef,
	}
	body, err := c.doJSON(ctx, http.MethodPatch, fmt.Sprintf("/issues/%d", issueID), payload)
	if err != nil {
		return ports.TaigaIssue{}, err
	}
	return decodeIssue(body)
}

func decodeIssue(body []byte) (ports.TaigaIssue, error) {
	var raw struct {
		ID                int    `json:"id"`
		Ref               int    `json:"ref"`
		Project           int    `json:"project"`
		Subject           string `json:"subject"`
		Description       string `json:"description"`
		Version           int    `json:"version"`
		ExternalReference []any  `json:"external_reference"`
		Permalink         string `json:"permalink"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return ports.TaigaIssue{}, fmt.Errorf("taiga decode issue: %w", err)
	}
	refs := make([]string, 0, len(raw.ExternalReference))
	for _, v := range raw.ExternalReference {
		refs = append(refs, fmt.Sprint(v))
	}
	return ports.TaigaIssue{
		ID:                raw.ID,
		Ref:               raw.Ref,
		ProjectID:         raw.Project,
		Subject:           strings.TrimSpace(raw.Subject),
		Description:       strings.TrimSpace(raw.Description),
		Version:           raw.Version,
		ExternalReference: refs,
		Permalink:         strings.TrimSpace(raw.Permalink),
	}, nil
}

func (c *Client) doJSON(ctx context.Context, method, path string, payload any) ([]byte, error) {
	token, err := c.auth(ctx)
	if err != nil {
		return nil, err
	}
	var reader io.Reader
	if payload != nil {
		b, err := json.Marshal(payload)
		if err != nil {
			return nil, err
		}
		reader = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.apiBase+path, reader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrTaigaUnavailable, err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("%w: status %d", domain.ErrTaigaUnavailable, resp.StatusCode)
	}
	return body, nil
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

var _ ports.TaigaGateway = (*Client)(nil)
