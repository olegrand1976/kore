package taiga

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kore/kore/internal/modules/integrations/domain"
	"github.com/stretchr/testify/require"
)

func TestClient_ListProjects(t *testing.T) {
	t.Parallel()
	var authCalls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/auth":
			authCalls++
			_ = json.NewEncoder(w).Encode(map[string]string{"auth_token": "tok"})
		case "/api/v1/projects":
			require.Equal(t, "Bearer tok", r.Header.Get("Authorization"))
			_ = json.NewEncoder(w).Encode([]map[string]any{
				{"id": 1, "name": "Kore TMA", "slug": "kore-tma"},
				{"id": 2, "name": "Other", "slug": "other"},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	client := NewClient(Config{BaseURL: srv.URL, Username: "svc", Password: "secret"})
	projects, err := client.ListProjects(context.Background())
	require.NoError(t, err)
	require.Equal(t, 1, authCalls)
	require.Len(t, projects, 2)
	require.Equal(t, "Kore TMA", projects[0].Name)
}

func TestClient_ListProjects_NotConfigured(t *testing.T) {
	t.Parallel()
	client := NewClient(Config{})
	_, err := client.ListProjects(context.Background())
	require.ErrorIs(t, err, domain.ErrTaigaNotConfigured)
}

func TestClient_ListProjects_Unavailable(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/auth" {
			_ = json.NewEncoder(w).Encode(map[string]string{"auth_token": "tok"})
			return
		}
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer srv.Close()

	client := NewClient(Config{BaseURL: srv.URL, Username: "svc", Password: "secret"})
	_, err := client.ListProjects(context.Background())
	require.ErrorIs(t, err, domain.ErrTaigaUnavailable)
}

func TestClient_CreateAndListIssues(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/v1/auth":
			_ = json.NewEncoder(w).Encode(map[string]string{"auth_token": "tok"})
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/issues":
			require.Equal(t, "Bearer tok", r.Header.Get("Authorization"))
			var body map[string]any
			require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
			require.Equal(t, float64(7), body["project"])
			require.Equal(t, "Bug export", body["subject"])
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id": 99, "ref": 12, "project": 7, "subject": "Bug export",
				"description": "desc", "version": 1,
				"external_reference": []any{"kore", "11111111-1111-4111-8111-111111111111"},
				"permalink":          "https://taiga.example/issue/12",
			})
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/issues":
			require.Equal(t, "7", r.URL.Query().Get("project"))
			_ = json.NewEncoder(w).Encode([]map[string]any{
				{"id": 99, "ref": 12, "project": 7, "subject": "Bug export", "version": 1},
			})
		case r.Method == http.MethodPatch && r.URL.Path == "/api/v1/issues/99":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id": 99, "ref": 12, "project": 7, "subject": "Bug export", "version": 2,
				"external_reference": []any{"kore", "11111111-1111-4111-8111-111111111111"},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	client := NewClient(Config{BaseURL: srv.URL, Username: "svc", Password: "secret"})
	created, err := client.CreateIssue(context.Background(), 7, "Bug export", "desc", []string{
		"kore", "11111111-1111-4111-8111-111111111111",
	})
	require.NoError(t, err)
	require.Equal(t, 99, created.ID)
	require.Equal(t, 12, created.Ref)

	listed, err := client.ListProjectIssues(context.Background(), 7)
	require.NoError(t, err)
	require.Len(t, listed, 1)

	updated, err := client.UpdateIssueExternalReference(context.Background(), 99, 1, []string{
		"kore", "11111111-1111-4111-8111-111111111111",
	})
	require.NoError(t, err)
	require.Equal(t, 2, updated.Version)
}
