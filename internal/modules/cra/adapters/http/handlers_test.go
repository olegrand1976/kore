package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kore/kore/internal/modules/cra/domain"
	"github.com/kore/kore/internal/platform/httpx"
)

func TestWriteCRAError_BusinessCodes(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantCode   httpx.ErrorCode
	}{
		{
			name:       "already validated",
			err:        domain.ErrCRAAlreadyValidated,
			wantStatus: http.StatusConflict,
			wantCode:   httpx.ErrCodeCRAAlreadyValidated,
		},
		{
			name:       "commercial info",
			err:        domain.ErrCommercialInfoRequired,
			wantStatus: http.StatusUnprocessableEntity,
			wantCode:   httpx.ErrCodeCommercialInfoRequired,
		},
		{
			name:       "day capacity",
			err:        domain.ErrDayCapacityExceeded,
			wantStatus: http.StatusUnprocessableEntity,
			wantCode:   httpx.ErrCodeDayCapacityExceeded,
		},
		{
			name:       "conflict absence",
			err:        domain.ErrCRAConflictAbsence,
			wantStatus: http.StatusConflict,
			wantCode:   httpx.ErrCodeCRAConflictAbsence,
		},
		{
			name:       "week incomplete",
			err:        domain.ErrWeekIncomplete,
			wantStatus: http.StatusUnprocessableEntity,
			wantCode:   httpx.ErrCodeWeekIncomplete,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			writeCRAError(rec, tc.err)
			if rec.Code != tc.wantStatus {
				t.Fatalf("status: got %d want %d", rec.Code, tc.wantStatus)
			}
			var env httpx.Envelope
			if err := json.NewDecoder(rec.Body).Decode(&env); err != nil {
				t.Fatalf("decode: %v", err)
			}
			if env.Error == nil {
				t.Fatal("expected error envelope")
			}
			if env.Error.Code != tc.wantCode {
				t.Fatalf("code: got %s want %s", env.Error.Code, tc.wantCode)
			}
		})
	}
}

func TestWriteCRAError_InternalFallback(t *testing.T) {
	rec := httptest.NewRecorder()
	writeCRAError(rec, errors.New("boom"))
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}
