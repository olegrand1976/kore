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
