package http

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/org/app"
	"github.com/kore/kore/internal/modules/org/domain"
	"github.com/kore/kore/internal/modules/org/ports"
	"github.com/kore/kore/internal/platform/authx"
	"github.com/kore/kore/internal/platform/httpx"
	"github.com/kore/kore/internal/platform/uploads"
	"github.com/kore/kore/pkg/kernel"
)

func RegisterRoutes(
	r chi.Router,
	org ports.OrganizationService,
	users ports.UserService,
	clients ports.ClientService,
	tokens *authx.TokenIssuer,
	authorizer authx.Authorizer,
	uploadsDir string,
	entitlements authx.EntitlementReader,
	leaveBootstrap ports.LeaveTypeBootstrapper,
) {
	r.Post("/auth/login", loginHandler(users))
	r.Post("/auth/refresh", refreshHandler(users))
	r.Post("/auth/logout", logoutHandler())

	r.Group(func(pr chi.Router) {
		pr.Use(httpx.AuthStack(tokens, entitlements))
		pr.Get("/societes", listSocietes(org))
		pr.Post("/societes", createSociete(org, authorizer, leaveBootstrap))
		pr.Put("/societes/{id}/branding", updateSocieteBranding(org, authorizer, uploadsDir))
		pr.Get("/branding/logo/{tenantId}", serveTenantLogo(uploadsDir))
		pr.Post("/sites", createSite(org, authorizer))
		pr.Post("/services", createService(org, authorizer))
		pr.Post("/applications", createApplication(org, authorizer))
		pr.Get("/applications", listApplications(org, authorizer))
		pr.Get("/applications/{id}", getApplication(org, authorizer))
		pr.Get("/users", listUsers(users, authorizer))
		pr.Get("/users/{id}", getUser(users, authorizer))
		pr.Post("/users", createUser(users, authorizer))
		pr.Put("/users/{id}", updateUser(users, authorizer))
		pr.Patch("/users/{id}/deactivate", deactivateUser(users, authorizer))
		pr.Delete("/users/{id}", deleteUser(users, authorizer))
		pr.Get("/clients", listClients(clients))
		pr.Get("/clients/{id}", getClient(clients, authorizer))
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

func refreshHandler(users ports.UserService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			RefreshToken string `json:"refreshToken"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid body")
			return
		}
		pair, err := users.RefreshSession(r.Context(), req.RefreshToken)
		if err != nil {
			switch {
			case errors.Is(err, domain.ErrAccountExpired):
				httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, err.Error())
			case errors.Is(err, domain.ErrInvalidCredentials):
				httpx.WriteError(w, http.StatusUnauthorized, httpx.ErrCodeUnauthorized, "invalid refresh token")
			default:
				httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, err.Error())
			}
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

func createSociete(org ports.OrganizationService, authorizer authx.Authorizer, leaveBootstrap ports.LeaveTypeBootstrapper) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !authorizer.Can(r.Context(), "org", authx.ActionWrite) {
			httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "forbidden")
			return
		}
		var req struct {
			RaisonSociale string `json:"raisonSociale"`
			Devise        string `json:"devise"`
			Pays          string `json:"pays"`
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
			Pays:          req.Pays,
		})
		if err != nil {
			httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, err.Error())
			return
		}
		if leaveBootstrap != nil {
			if err := leaveBootstrap.BootstrapDefaults(r.Context(), identity.TenantID, s.ID); err != nil {
				httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, err.Error())
				return
			}
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

func canReadApplications(ctx context.Context, authorizer authx.Authorizer) bool {
	return authorizer.Can(ctx, "org", authx.ActionRead) ||
		authorizer.Can(ctx, "budget", authx.ActionRead) ||
		authorizer.Can(ctx, "tma", authx.ActionRead) ||
		authorizer.Can(ctx, "cra", authx.ActionRead)
}

func listApplications(org ports.OrganizationService, authorizer authx.Authorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !canReadApplications(r.Context(), authorizer) {
			httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "forbidden")
			return
		}
		identity, _ := authx.FromContext(r.Context())
		items, err := org.ListApplications(r.Context(), identity.TenantID)
		if err != nil {
			httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, err.Error())
			return
		}
		httpx.WriteData(w, http.StatusOK, items)
	}
}

func getApplication(org ports.OrganizationService, authorizer authx.Authorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !canReadApplications(r.Context(), authorizer) {
			httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "forbidden")
			return
		}
		appID, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid application id")
			return
		}
		identity, _ := authx.FromContext(r.Context())
		item, err := org.GetApplication(r.Context(), identity.TenantID, appID)
		if err != nil {
			httpx.WriteError(w, http.StatusNotFound, httpx.ErrCodeNotFound, "application not found")
			return
		}
		httpx.WriteData(w, http.StatusOK, item)
	}
}

func createUser(users ports.UserService, authorizer authx.Authorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !authorizer.Can(r.Context(), "org", authx.ActionWrite) {
			httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "forbidden")
			return
		}
		var req struct {
			Login    string         `json:"login"`
			Password string         `json:"password"`
			Profile  domain.Profile `json:"profil"`
			EquipeID *uuid.UUID     `json:"equipeId"`
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

func getClient(clients ports.ClientService, authorizer authx.Authorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !authorizer.Can(r.Context(), "org", authx.ActionRead) &&
			!authorizer.Can(r.Context(), "cra", authx.ActionRead) &&
			!authorizer.Can(r.Context(), "ssii", authx.ActionRead) {
			httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "forbidden")
			return
		}
		clientID, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid client id")
			return
		}
		identity, _ := authx.FromContext(r.Context())
		item, err := clients.GetClient(r.Context(), identity.TenantID, clientID)
		if err != nil {
			httpx.WriteError(w, http.StatusNotFound, httpx.ErrCodeNotFound, "client not found")
			return
		}
		httpx.WriteData(w, http.StatusOK, item)
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

func updateSocieteBranding(org ports.OrganizationService, authorizer authx.Authorizer, uploadsDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !authorizer.Can(r.Context(), "org", authx.ActionWrite) {
			httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "forbidden")
			return
		}
		societeID, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid societe id")
			return
		}
		identity, _ := authx.FromContext(r.Context())
		if err := r.ParseMultipartForm(512 << 10); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid multipart form")
			return
		}
		cmd := ports.UpdateSocieteBrandingCommand{
			TenantID:      identity.TenantID,
			SocieteID:     societeID,
			RaisonSociale: r.FormValue("raisonSociale"),
			Adresse:       r.FormValue("adresse"),
			Siret:         r.FormValue("siret"),
			URLTenant:     r.FormValue("urlTenant"),
		}
		if file, header, err := r.FormFile("logo"); err == nil {
			defer file.Close()
			if err := uploads.ValidateLogoFilename(header.Filename); err != nil {
				httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, err.Error())
				return
			}
			logoURL, err := uploads.Store(uploadsDir, identity.TenantID.UUID(), societeID, header.Filename, file)
			if err != nil {
				writeUploadError(w, err)
				return
			}
			cmd.Logo = logoURL
		}
		societe, err := org.UpdateSocieteBranding(r.Context(), cmd)
		if err != nil {
			httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, err.Error())
			return
		}
		httpx.WriteData(w, http.StatusOK, societe)
	}
}

func serveTenantLogo(uploadsDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		identity, ok := authx.FromContext(r.Context())
		if !ok {
			httpx.WriteError(w, http.StatusUnauthorized, httpx.ErrCodeUnauthorized, "unauthorized")
			return
		}
		tenantID, err := uuid.Parse(chi.URLParam(r, "tenantId"))
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid tenant id")
			return
		}
		if identity.TenantID.UUID() != tenantID {
			httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "forbidden")
			return
		}
		path, ok := uploads.Path(uploadsDir, tenantID)
		if !ok {
			httpx.WriteError(w, http.StatusNotFound, httpx.ErrCodeNotFound, "logo not found")
			return
		}
		f, err := os.Open(path)
		if err != nil {
			httpx.WriteError(w, http.StatusNotFound, httpx.ErrCodeNotFound, "logo not found")
			return
		}
		defer f.Close()
		w.Header().Set("Content-Type", uploads.ContentTypeForExt(path))
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.WriteHeader(http.StatusOK)
		_, _ = io.Copy(w, f)
	}
}

func writeUploadError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, uploads.ErrInvalidLogo),
		errors.Is(err, uploads.ErrLogoTooLarge),
		errors.Is(err, uploads.ErrUnsupportedExt):
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, err.Error())
	default:
		httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, err.Error())
	}
}

func listUsers(users ports.UserService, authorizer authx.Authorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		canRead := authorizer.Can(r.Context(), "org", authx.ActionRead)
		canValidateConges := authorizer.Can(r.Context(), "conges", authx.ActionValidate)
		canValidateCra := authorizer.Can(r.Context(), "cra", authx.ActionValidate)
		if !canRead && !canValidateConges && !canValidateCra {
			httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "forbidden")
			return
		}
		identity, _ := authx.FromContext(r.Context())
		items, err := users.ListUsers(r.Context(), identity.TenantID)
		if err != nil {
			httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, err.Error())
			return
		}
		httpx.WriteData(w, http.StatusOK, items)
	}
}

func getUser(users ports.UserService, authorizer authx.Authorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		canRead := authorizer.Can(r.Context(), "org", authx.ActionRead)
		canValidateConges := authorizer.Can(r.Context(), "conges", authx.ActionValidate)
		canValidateCra := authorizer.Can(r.Context(), "cra", authx.ActionValidate)
		canReadCra := authorizer.Can(r.Context(), "cra", authx.ActionRead)
		if !canRead && !canValidateConges && !canValidateCra && !canReadCra {
			httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "forbidden")
			return
		}
		userID, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid user id")
			return
		}
		identity, _ := authx.FromContext(r.Context())
		item, err := users.GetUser(r.Context(), identity.TenantID, userID)
		if err != nil {
			if errors.Is(err, domain.ErrUserNotFound) {
				httpx.WriteError(w, http.StatusNotFound, httpx.ErrCodeNotFound, err.Error())
				return
			}
			httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, err.Error())
			return
		}
		httpx.WriteData(w, http.StatusOK, item)
	}
}

func updateUser(users ports.UserService, authorizer authx.Authorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !authorizer.Can(r.Context(), "org", authx.ActionWrite) {
			httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "forbidden")
			return
		}
		userID, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid user id")
			return
		}
		var req struct {
			Profile  *domain.Profile `json:"profil"`
			Password string          `json:"password"`
			Active   *bool           `json:"active"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid body")
			return
		}
		identity, _ := authx.FromContext(r.Context())
		summary, err := users.UpdateUser(r.Context(), ports.UpdateUserCommand{
			TenantID:    identity.TenantID,
			UserID:      userID,
			ActorUserID: identity.UserID,
			Profile:     req.Profile,
			Password:    req.Password,
			Active:      req.Active,
		})
		if err != nil {
			writeUserMutationError(w, err)
			return
		}
		httpx.WriteData(w, http.StatusOK, summary)
	}
}

func deactivateUser(users ports.UserService, authorizer authx.Authorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !authorizer.Can(r.Context(), "org", authx.ActionWrite) {
			httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "forbidden")
			return
		}
		userID, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid user id")
			return
		}
		identity, _ := authx.FromContext(r.Context())
		err = users.DeactivateUser(r.Context(), ports.DeleteUserCommand{
			TenantID:    identity.TenantID,
			UserID:      userID,
			ActorUserID: identity.UserID,
		})
		if err != nil {
			writeUserMutationError(w, err)
			return
		}
		httpx.WriteData(w, http.StatusOK, map[string]string{"status": "deactivated"})
	}
}

func deleteUser(users ports.UserService, authorizer authx.Authorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !authorizer.Can(r.Context(), "org", authx.ActionWrite) {
			httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "forbidden")
			return
		}
		userID, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid user id")
			return
		}
		identity, _ := authx.FromContext(r.Context())
		err = users.DeleteUser(r.Context(), ports.DeleteUserCommand{
			TenantID:    identity.TenantID,
			UserID:      userID,
			ActorUserID: identity.UserID,
		})
		if err != nil {
			writeUserMutationError(w, err)
			return
		}
		httpx.WriteData(w, http.StatusOK, map[string]string{"status": "deleted"})
	}
}

func writeUserMutationError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrUserNotFound):
		httpx.WriteError(w, http.StatusNotFound, httpx.ErrCodeNotFound, err.Error())
	case errors.Is(err, domain.ErrCannotModifySelf):
		httpx.WriteError(w, http.StatusUnprocessableEntity, httpx.ErrCodeValidation, err.Error())
	case errors.Is(err, domain.ErrSeatLimitReached):
		httpx.WriteError(w, http.StatusConflict, httpx.ErrCodeConflict, err.Error())
	default:
		httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, err.Error())
	}
}

var _ = app.DefaultPermissions
var _ = kernel.TenantID{}
