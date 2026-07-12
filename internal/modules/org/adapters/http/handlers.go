package http

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/org/app"
	"github.com/kore/kore/internal/modules/org/domain"
	"github.com/kore/kore/internal/modules/org/ports"
	"github.com/kore/kore/internal/platform/authx"
	"github.com/kore/kore/internal/platform/httpx"
	"github.com/kore/kore/pkg/kernel"
)

func RegisterRoutes(
	r chi.Router,
	org ports.OrganizationService,
	users ports.UserService,
	clients ports.ClientService,
	tokens *authx.TokenIssuer,
	authorizer authx.Authorizer,
) {
	r.Post("/auth/login", loginHandler(users))
	r.Post("/auth/refresh", refreshHandler(tokens))
	r.Post("/auth/logout", logoutHandler())

	r.Group(func(pr chi.Router) {
		pr.Use(httpx.AuthMiddleware(tokens))
		pr.Get("/societes", listSocietes(org))
		pr.Post("/societes", createSociete(org, authorizer))
		pr.Post("/sites", createSite(org, authorizer))
		pr.Post("/services", createService(org, authorizer))
		pr.Post("/applications", createApplication(org, authorizer))
		pr.Post("/users", createUser(users, authorizer))
		pr.Get("/clients", listClients(clients))
		pr.Post("/clients", createClient(clients, authorizer))
	})
}

type loginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

func loginHandler(users ports.UserService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req loginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid body")
			return
		}
		result, err := users.Authenticate(r.Context(), req.Login, req.Password)
		if err != nil {
			switch {
			case errors.Is(err, domain.ErrInvalidCredentials):
				httpx.WriteError(w, http.StatusUnauthorized, httpx.ErrCodeUnauthorized, err.Error())
			case errors.Is(err, domain.ErrAccountExpired):
				httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, err.Error())
			default:
				httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, err.Error())
			}
			return
		}
		httpx.WriteData(w, http.StatusOK, result)
	}
}

func refreshHandler(tokens *authx.TokenIssuer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			RefreshToken string `json:"refreshToken"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid body")
			return
		}
		identity, err := tokens.ParseRefreshToken(req.RefreshToken)
		if err != nil {
			httpx.WriteError(w, http.StatusUnauthorized, httpx.ErrCodeUnauthorized, "invalid refresh token")
			return
		}
		pair, err := tokens.Issue(identity)
		if err != nil {
			httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, err.Error())
			return
		}
		httpx.WriteData(w, http.StatusOK, pair)
	}
}

func logoutHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		httpx.WriteData(w, http.StatusOK, map[string]string{"status": "logged_out"})
	}
}

func listSocietes(org ports.OrganizationService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		identity, _ := authx.FromContext(r.Context())
		items, err := org.ListSocietes(r.Context(), identity.TenantID)
		if err != nil {
			httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, err.Error())
			return
		}
		httpx.WriteData(w, http.StatusOK, items)
	}
}

func createSociete(org ports.OrganizationService, authorizer authx.Authorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !authorizer.Can(r.Context(), "org", authx.ActionWrite) {
			httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "forbidden")
			return
		}
		var req struct {
			RaisonSociale string `json:"raisonSociale"`
			Devise        string `json:"devise"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid body")
			return
		}
		identity, _ := authx.FromContext(r.Context())
		s, err := org.CreateSociete(r.Context(), ports.CreateSocieteCommand{
			TenantID:      identity.TenantID,
			RaisonSociale: req.RaisonSociale,
			Devise:        req.Devise,
		})
		if err != nil {
			httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, err.Error())
			return
		}
		httpx.WriteData(w, http.StatusCreated, s)
	}
}

func createSite(org ports.OrganizationService, authorizer authx.Authorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !authorizer.Can(r.Context(), "org", authx.ActionWrite) {
			httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "forbidden")
			return
		}
		var req struct {
			SocieteID uuid.UUID `json:"societeId"`
			Libelle   string    `json:"libelle"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid body")
			return
		}
		identity, _ := authx.FromContext(r.Context())
		s, err := org.CreateSite(r.Context(), ports.CreateSiteCommand{
			TenantID:  identity.TenantID,
			SocieteID: req.SocieteID,
			Libelle:   req.Libelle,
		})
		if err != nil {
			httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, err.Error())
			return
		}
		httpx.WriteData(w, http.StatusCreated, s)
	}
}

func createService(org ports.OrganizationService, authorizer authx.Authorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !authorizer.Can(r.Context(), "org", authx.ActionWrite) {
			httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "forbidden")
			return
		}
		var req struct {
			SiteID        uuid.UUID `json:"siteId"`
			ResponsableID uuid.UUID `json:"responsableId"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid body")
			return
		}
		identity, _ := authx.FromContext(r.Context())
		s, err := org.CreateService(r.Context(), ports.CreateServiceCommand{
			TenantID:      identity.TenantID,
			SiteID:        req.SiteID,
			ResponsableID: req.ResponsableID,
		})
		if err != nil {
			if errors.Is(err, domain.ErrServiceWithoutResponsible) {
				httpx.WriteError(w, http.StatusUnprocessableEntity, httpx.ErrCodeValidation, err.Error())
				return
			}
			httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, err.Error())
			return
		}
		httpx.WriteData(w, http.StatusCreated, s)
	}
}

func createApplication(org ports.OrganizationService, authorizer authx.Authorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !authorizer.Can(r.Context(), "org", authx.ActionWrite) {
			httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "forbidden")
			return
		}
		var req struct {
			ServiceID uuid.UUID `json:"serviceId"`
			Libelle   string    `json:"libelle"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid body")
			return
		}
		identity, _ := authx.FromContext(r.Context())
		a, err := org.CreateApplication(r.Context(), ports.CreateApplicationCommand{
			TenantID:  identity.TenantID,
			ServiceID: req.ServiceID,
			Libelle:   req.Libelle,
		})
		if err != nil {
			httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, err.Error())
			return
		}
		httpx.WriteData(w, http.StatusCreated, a)
	}
}

func createUser(users ports.UserService, authorizer authx.Authorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !authorizer.Can(r.Context(), "org", authx.ActionWrite) {
			httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "forbidden")
			return
		}
		var req struct {
			Login    string          `json:"login"`
			Password string          `json:"password"`
			Profile  domain.Profile  `json:"profil"`
			EquipeID *uuid.UUID      `json:"equipeId"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid body")
			return
		}
		identity, _ := authx.FromContext(r.Context())
		u, err := users.CreateUser(r.Context(), ports.CreateUserCommand{
			TenantID: identity.TenantID,
			Login:    req.Login,
			Password: req.Password,
			Profile:  req.Profile,
			EquipeID: req.EquipeID,
		})
		if err != nil {
			switch {
			case errors.Is(err, domain.ErrLoginAlreadyExists):
				httpx.WriteError(w, http.StatusConflict, httpx.ErrCodeConflict, err.Error())
			case errors.Is(err, domain.ErrInvalidLogin):
				httpx.WriteError(w, http.StatusUnprocessableEntity, httpx.ErrCodeValidation, err.Error())
			case errors.Is(err, domain.ErrSeatLimitReached):
				httpx.WriteError(w, http.StatusConflict, httpx.ErrCodeConflict, err.Error())
			default:
				httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, err.Error())
			}
			return
		}
		httpx.WriteData(w, http.StatusCreated, map[string]any{"id": u.ID, "login": u.Login})
	}
}

func listClients(clients ports.ClientService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		identity, _ := authx.FromContext(r.Context())
		items, err := clients.ListClients(r.Context(), identity.TenantID)
		if err != nil {
			httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, err.Error())
			return
		}
		httpx.WriteData(w, http.StatusOK, items)
	}
}

func createClient(clients ports.ClientService, authorizer authx.Authorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !authorizer.Can(r.Context(), "org", authx.ActionWrite) {
			httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "forbidden")
			return
		}
		var req struct {
			RaisonSociale string `json:"raisonSociale"`
			TVA           string `json:"tva"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid body")
			return
		}
		identity, _ := authx.FromContext(r.Context())
		c, err := clients.CreateClient(r.Context(), ports.CreateClientCommand{
			TenantID:      identity.TenantID,
			RaisonSociale: req.RaisonSociale,
			TVA:           req.TVA,
		})
		if err != nil {
			httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, err.Error())
			return
		}
		httpx.WriteData(w, http.StatusCreated, c)
	}
}

var _ = app.DefaultPermissions
var _ = kernel.TenantID{}
