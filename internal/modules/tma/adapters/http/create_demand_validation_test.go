package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/platform/authx"
	"github.com/kore/kore/internal/platform/httpx"
	"github.com/kore/kore/pkg/kernel"
)

type allowAuthorizer struct{}

func (allowAuthorizer) Can(context.Context, authx.Module, authx.Action) bool { return true }

type enabledChannels struct{}

func (enabledChannels) IsChannelEnabled(context.Context, kernel.TenantID, kernel.RequestChannel) (bool, error) {
	return true, nil
}

func TestCreateDemand_validation(t *testing.T) {
	tenant := kernel.NewTenantID(uuid.New())
	userID := uuid.New()
	ctx := authx.WithIdentity(context.Background(), authx.Identity{
		TenantID: tenant,
		UserID:   userID,
		Profile:  "Administrateur",
	})
	h := createDemand(nil, allowAuthorizer{}, enabledChannels{})

	t.Run("missing subject", func(t *testing.T) {
		body, _ := json.Marshal(map[string]any{
			"applicationId": uuid.New().String(),
			"subject":       "   ",
		})
		req := httptest.NewRequest(http.MethodPost, "/demands", bytes.NewReader(body)).WithContext(ctx)
		rec := httptest.NewRecorder()
		h(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status: got %d want %d", rec.Code, http.StatusBadRequest)
		}
		var env httpx.Envelope
		if err := json.NewDecoder(rec.Body).Decode(&env); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if env.Error == nil || env.Error.Message != "subject required" {
			t.Fatalf("unexpected envelope: %+v", env.Error)
		}
	})

	t.Run("missing applicationId", func(t *testing.T) {
		body, _ := json.Marshal(map[string]any{
			"applicationId": "00000000-0000-0000-0000-000000000000",
			"subject":       "Suivi d'équipe",
		})
		req := httptest.NewRequest(http.MethodPost, "/demands", bytes.NewReader(body)).WithContext(ctx)
		rec := httptest.NewRecorder()
		h(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status: got %d want %d", rec.Code, http.StatusBadRequest)
		}
	})
}
