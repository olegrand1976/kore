package http

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/kore/kore/internal/modules/org/domain"
	"github.com/kore/kore/internal/modules/org/ports"
	"github.com/kore/kore/internal/platform/authx"
	"github.com/kore/kore/internal/platform/httpx"
)

func verify2FAHandler(users ports.UserService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			ChallengeToken string `json:"challengeToken"`
			Code           string `json:"code"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid body")
			return
		}
		result, err := users.Verify2FAChallenge(r.Context(), req.ChallengeToken, req.Code)
		if err != nil {
			write2FAError(w, err)
			return
		}
		httpx.WriteData(w, http.StatusOK, result)
	}
}

func verify2FAEnrollmentHandler(users ports.UserService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			EnrollmentToken string `json:"enrollmentToken"`
			Code            string `json:"code"`
			Password        string `json:"password"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid body")
			return
		}
		result, err := users.Verify2FAEnrollment(r.Context(), req.EnrollmentToken, req.Code, req.Password)
		if err != nil {
			write2FAError(w, err)
			return
		}
		httpx.WriteData(w, http.StatusOK, result)
	}
}

func setup2FAEnrollmentHandler(users ports.UserService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			EnrollmentToken string `json:"enrollmentToken"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid body")
			return
		}
		result, err := users.Setup2FAWithEnrollmentToken(r.Context(), req.EnrollmentToken)
		if err != nil {
			write2FAError(w, err)
			return
		}
		httpx.WriteData(w, http.StatusOK, result)
	}
}

func get2FAStatusHandler(users ports.UserService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		identity, ok := authx.FromContext(r.Context())
		if !ok {
			httpx.WriteError(w, http.StatusUnauthorized, httpx.ErrCodeUnauthorized, "unauthorized")
			return
		}
		status, err := users.Get2FAStatus(r.Context(), identity.TenantID, identity.UserID)
		if err != nil {
			httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, err.Error())
			return
		}
		httpx.WriteData(w, http.StatusOK, status)
	}
}

func setup2FAHandler(users ports.UserService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		identity, ok := authx.FromContext(r.Context())
		if !ok {
			httpx.WriteError(w, http.StatusUnauthorized, httpx.ErrCodeUnauthorized, "unauthorized")
			return
		}
		result, err := users.Setup2FA(r.Context(), identity.TenantID, identity.UserID)
		if err != nil {
			write2FAError(w, err)
			return
		}
		httpx.WriteData(w, http.StatusOK, result)
	}
}

func confirm2FAHandler(users ports.UserService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		identity, ok := authx.FromContext(r.Context())
		if !ok {
			httpx.WriteError(w, http.StatusUnauthorized, httpx.ErrCodeUnauthorized, "unauthorized")
			return
		}
		var req struct {
			Code     string `json:"code"`
			Password string `json:"password"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid body")
			return
		}
		result, err := users.Confirm2FA(r.Context(), ports.Confirm2FACommand{
			TenantID: identity.TenantID,
			UserID:   identity.UserID,
			Code:     req.Code,
			Password: req.Password,
		})
		if err != nil {
			write2FAError(w, err)
			return
		}
		httpx.WriteData(w, http.StatusOK, result)
	}
}

func disable2FAHandler(users ports.UserService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		identity, ok := authx.FromContext(r.Context())
		if !ok {
			httpx.WriteError(w, http.StatusUnauthorized, httpx.ErrCodeUnauthorized, "unauthorized")
			return
		}
		var req struct {
			Password string `json:"password"`
			Code     string `json:"code"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid body")
			return
		}
		err := users.Disable2FA(r.Context(), ports.Disable2FACommand{
			TenantID: identity.TenantID,
			UserID:   identity.UserID,
			Password: req.Password,
			Code:     req.Code,
		})
		if err != nil {
			write2FAError(w, err)
			return
		}
		httpx.WriteData(w, http.StatusOK, map[string]string{"status": "disabled"})
	}
}

func write2FAError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrInvalidCredentials):
		httpx.WriteError(w, http.StatusUnauthorized, httpx.ErrCodeUnauthorized, err.Error())
	case errors.Is(err, domain.Err2FAInvalidCode):
		httpx.WriteError(w, http.StatusUnauthorized, httpx.ErrCodeUnauthorized, err.Error())
	case errors.Is(err, domain.Err2FAChallengeExpired), errors.Is(err, domain.Err2FAEnrollmentTokenInvalid):
		httpx.WriteError(w, http.StatusUnauthorized, httpx.ErrCodeUnauthorized, err.Error())
	case errors.Is(err, domain.Err2FAPolicyForbidden):
		httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, err.Error())
	case errors.Is(err, domain.Err2FAAlreadyEnabled), errors.Is(err, domain.Err2FANotEnabled):
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, err.Error())
	case errors.Is(err, domain.Err2FAPasswordRequired):
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, err.Error())
	case errors.Is(err, domain.Err2FARateLimited):
		httpx.WriteError(w, http.StatusTooManyRequests, httpx.ErrCodeForbidden, err.Error())
	default:
		httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, err.Error())
	}
}
