package http

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	integrationdomain "github.com/kore/kore/internal/modules/integrations/domain"
	integrationports "github.com/kore/kore/internal/modules/integrations/ports"
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
	tenantAccess ports.TenantAccessService,
	tokens *authx.TokenIssuer,
	authorizer authx.Authorizer,
	uploadsDir string,
	attachments ports.AttachmentService,
	entitlements authx.EntitlementReader,
	leaveBootstrap ports.LeaveTypeBootstrapper,
	requestSettings ports.RequestSettingsService,
	taigaBridge TaigaApplicationBridge,
) {
	r.Post("/auth/login", loginHandler(users))
	r.Post("/auth/2fa/verify", verify2FAHandler(users))
	r.Post("/auth/2fa/enrollment/setup", setup2FAEnrollmentHandler(users))
	r.Post("/auth/2fa/enrollment/confirm", verify2FAEnrollmentHandler(users))
	r.Post("/auth/refresh", refreshHandler(users))
	r.Post("/auth/logout", logoutHandler())
	r.Post("/auth/tenant-discovery/request", tenantDiscoveryRequestHandler(tenantAccess))
	r.Get("/auth/tenant-discovery/resolve", tenantDiscoveryResolveHandler(tenantAccess))
	r.Get("/public/invitations/resolve", invitationResolveHandler(tenantAccess))

	r.Group(func(pr chi.Router) {
		pr.Use(httpx.AuthStack(tokens, entitlements))
		pr.Get("/societes", listSocietes(org))
		pr.Post("/societes", createSociete(org, authorizer, leaveBootstrap))
		pr.Put("/societes/{id}/branding", updateSocieteBranding(org, authorizer))
		pr.Put("/societes/{id}/settings", updateSocieteSettings(org, authorizer))
		pr.Get("/branding/logo/{tenantId}", serveTenantLogo(org, uploadsDir))
		pr.Post("/sites", createSite(org, authorizer))
		pr.Get("/sites", listSites(org, authorizer))
		pr.Put("/sites/{id}", updateSite(org, authorizer))
		pr.Post("/services", createService(org, authorizer))
		pr.Get("/services", listServices(org, authorizer))
		pr.Put("/services/{id}", updateService(org, authorizer))
		pr.Post("/applications", createApplication(org, authorizer, taigaBridge))
		pr.Get("/applications", listApplications(org, authorizer))
		pr.Post("/applications/merge", mergeApplications(org, authorizer))
		pr.Get("/applications/{id}", getApplication(org, authorizer))
		pr.Put("/applications/{id}", updateApplication(org, authorizer, taigaBridge))
		pr.Patch("/applications/{id}/deactivate", deactivateApplication(org, authorizer))
		pr.Patch("/applications/{id}/activate", activateApplication(org, authorizer))
		pr.Post("/equipes", createEquipe(org, authorizer))
		pr.Get("/equipes", listEquipes(org, authorizer))
		pr.Put("/equipes/{id}", updateEquipe(org, authorizer))
		pr.Get("/users", listUsers(users, authorizer))
		pr.Get("/users/{id}", getUser(users, authorizer))
		pr.Get("/users/me/2fa", get2FAStatusHandler(users))
		pr.Post("/users/me/2fa/setup", setup2FAHandler(users))
		pr.Post("/users/me/2fa/confirm", confirm2FAHandler(users))
		pr.Post("/users/me/2fa/disable", disable2FAHandler(users))
		pr.Get("/users/me/release-notes", getReleaseNotesPreferences(users))
		pr.Get("/users/me/profile", getMeProfile(users))
		pr.Get("/users/me/calendar-settings", getUserCalendarSettings(org))
		pr.Post("/users/me/release-notes/auto-show", setReleaseNotesAutoShow(users))
		pr.Post("/users/me/release-notes/seen", markReleaseNotesSeen(users))
		pr.Post("/users", createUser(users, authorizer))
		pr.Put("/users/{id}", updateUser(users, authorizer))
		pr.Patch("/users/{id}/deactivate", deactivateUser(users, authorizer))
		pr.Delete("/users/{id}", deleteUser(users, authorizer))
		pr.Get("/clients", listClients(clients, authorizer))
		pr.Get("/clients/{id}", getClient(clients, authorizer))
		pr.Post("/clients", createClient(clients, authorizer))
		pr.Put("/clients/{id}", updateClient(clients, authorizer))
		pr.Put("/clients/{id}/contacts", replaceClientContacts(clients, authorizer))

		pr.Post("/admin/invitations", createInvitationHandler(tenantAccess, authorizer))
		registerAttachmentRoutes(pr, attachments, authorizer, uploadsDir)
		if requestSettings != nil {
			registerRequestSettingsRoutes(pr, requestSettings, authorizer)
		}
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

func updateSite(org ports.OrganizationService, authorizer authx.Authorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !authorizer.Can(r.Context(), "org", authx.ActionWrite) {
			httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "forbidden")
			return
		}
		siteID, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid site id")
			return
		}
		var req struct {
			Libelle string `json:"libelle"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid body")
			return
		}
		identity, _ := authx.FromContext(r.Context())
		item, err := org.UpdateSite(r.Context(), ports.UpdateSiteCommand{
			TenantID: identity.TenantID,
			SiteID:   siteID,
			Libelle:  req.Libelle,
		})
		if err != nil {
			if errors.Is(err, domain.ErrSiteNotFound) {
				httpx.WriteError(w, http.StatusNotFound, httpx.ErrCodeNotFound, "site not found")
				return
			}
			if errors.Is(err, domain.ErrInvalidSiteLibelle) {
				httpx.WriteError(w, http.StatusUnprocessableEntity, httpx.ErrCodeValidation, err.Error())
				return
			}
			httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, err.Error())
			return
		}
		httpx.WriteData(w, http.StatusOK, item)
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
			Libelle       string    `json:"libelle"`
			Type          string    `json:"type"`
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
			Libelle:       req.Libelle,
			Type:          req.Type,
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

func updateService(org ports.OrganizationService, authorizer authx.Authorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !authorizer.Can(r.Context(), "org", authx.ActionWrite) {
			httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "forbidden")
			return
		}
		serviceID, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid service id")
			return
		}
		var req struct {
			Libelle string `json:"libelle"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid body")
			return
		}
		identity, _ := authx.FromContext(r.Context())
		item, err := org.UpdateService(r.Context(), ports.UpdateServiceCommand{
			TenantID:  identity.TenantID,
			ServiceID: serviceID,
			Libelle:   req.Libelle,
		})
		if err != nil {
			if errors.Is(err, domain.ErrServiceNotFound) {
				httpx.WriteError(w, http.StatusNotFound, httpx.ErrCodeNotFound, "service not found")
				return
			}
			if errors.Is(err, domain.ErrInvalidServiceLibelle) {
				httpx.WriteError(w, http.StatusUnprocessableEntity, httpx.ErrCodeValidation, err.Error())
				return
			}
			httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, err.Error())
			return
		}
		httpx.WriteData(w, http.StatusOK, item)
	}
}

func createApplication(org ports.OrganizationService, authorizer authx.Authorizer, taigaBridge TaigaApplicationBridge) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !authorizer.Can(r.Context(), "org", authx.ActionWrite) {
			httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "forbidden")
			return
		}
		var req struct {
			Libelle            string      `json:"libelle"`
			Proprietaire       string      `json:"proprietaire"`
			ModeFacturation    string      `json:"modeFacturation"`
			UOActivee          bool        `json:"uoActivee"`
			ChefUtilisateurID  *uuid.UUID  `json:"chefUtilisateurId"`
			BudgetDefautID     *uuid.UUID  `json:"budgetDefautId"`
			DefaultTJMCents    int64       `json:"defaultTjmCents"`
			SiteIDs            []uuid.UUID `json:"siteIds"`
			ServiceIDs         []uuid.UUID `json:"serviceIds"`
			EquipeIDs          []uuid.UUID `json:"equipeIds"`
			MethodologyProfile string      `json:"methodologyProfile"`
			TaigaProjectID     *int        `json:"taigaProjectId"`
			// Legacy single serviceId still accepted and merged into serviceIds.
			ServiceID uuid.UUID `json:"serviceId"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid body")
			return
		}
		serviceIDs := append([]uuid.UUID{}, req.ServiceIDs...)
		if req.ServiceID != uuid.Nil {
			serviceIDs = append(serviceIDs, req.ServiceID)
		}
		identity, _ := authx.FromContext(r.Context())
		if req.TaigaProjectID != nil && *req.TaigaProjectID > 0 {
			if taigaBridge == nil {
				httpx.WriteError(w, http.StatusServiceUnavailable, httpx.ErrCodeInternal, "taiga not configured")
				return
			}
			a, err := taigaBridge.CreateApplicationWithTaiga(r.Context(), integrationports.CreateApplicationInput{
				TenantID:           identity.TenantID,
				Libelle:            req.Libelle,
				Proprietaire:       req.Proprietaire,
				ModeFacturation:    req.ModeFacturation,
				UOActivee:          req.UOActivee,
				ChefUtilisateurID:  req.ChefUtilisateurID,
				DefaultTJMCents:    req.DefaultTJMCents,
				SiteIDs:            req.SiteIDs,
				ServiceIDs:         serviceIDs,
				EquipeIDs:          req.EquipeIDs,
				MethodologyProfile: req.MethodologyProfile,
			}, *req.TaigaProjectID)
			if err != nil {
				writeApplicationTaigaError(w, err)
				return
			}
			app, err := org.GetApplication(r.Context(), identity.TenantID, a.ID)
			if err != nil {
				httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, err.Error())
				return
			}
			httpx.WriteData(w, http.StatusCreated, app)
			return
		}
		a, err := org.CreateApplication(r.Context(), ports.CreateApplicationCommand{
			TenantID:           identity.TenantID,
			Libelle:            req.Libelle,
			Proprietaire:       req.Proprietaire,
			ModeFacturation:    req.ModeFacturation,
			UOActivee:          req.UOActivee,
			ChefUtilisateurID:  req.ChefUtilisateurID,
			BudgetDefautID:     req.BudgetDefautID,
			DefaultTJMCents:    req.DefaultTJMCents,
			SiteIDs:            req.SiteIDs,
			ServiceIDs:         serviceIDs,
			EquipeIDs:          req.EquipeIDs,
			MethodologyProfile: req.MethodologyProfile,
		})
		if err != nil {
			if errors.Is(err, domain.ErrInvalidModeFacturation) ||
				errors.Is(err, domain.ErrInvalidMethodologyProfile) {
				httpx.WriteError(w, http.StatusUnprocessableEntity, httpx.ErrCodeValidation, err.Error())
				return
			}
			if errors.Is(err, domain.ErrUserNotFound) ||
				errors.Is(err, domain.ErrInvalidApplicationLibelle) ||
				errors.Is(err, domain.ErrBudgetNotFound) ||
				errors.Is(err, domain.ErrBudgetNotAllowedOnCreate) ||
				errors.Is(err, domain.ErrApplicationWithoutShare) ||
				errors.Is(err, domain.ErrInvalidApplicationShare) {
				httpx.WriteError(w, http.StatusUnprocessableEntity, httpx.ErrCodeValidation, err.Error())
				return
			}
			httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, err.Error())
			return
		}
		httpx.WriteData(w, http.StatusCreated, a)
	}
}

func writeApplicationTaigaError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrInvalidModeFacturation),
		errors.Is(err, domain.ErrInvalidMethodologyProfile),
		errors.Is(err, domain.ErrUserNotFound),
		errors.Is(err, domain.ErrInvalidApplicationLibelle),
		errors.Is(err, domain.ErrBudgetNotFound),
		errors.Is(err, domain.ErrBudgetNotAllowedOnCreate),
		errors.Is(err, domain.ErrApplicationWithoutShare),
		errors.Is(err, domain.ErrInvalidApplicationShare),
		errors.Is(err, integrationdomain.ErrTaigaProjectNotFound),
		errors.Is(err, integrationdomain.ErrTaigaProjectLinked),
		errors.Is(err, integrationdomain.ErrTaigaApplicationAlreadyLinked):
		httpx.WriteError(w, http.StatusUnprocessableEntity, httpx.ErrCodeValidation, err.Error())
	case errors.Is(err, integrationdomain.ErrTaigaNotConfigured),
		errors.Is(err, integrationdomain.ErrTaigaUnavailable):
		httpx.WriteError(w, http.StatusServiceUnavailable, httpx.ErrCodeInternal, err.Error())
	default:
		httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, err.Error())
	}
}

func canReadApplications(ctx context.Context, authorizer authx.Authorizer) bool {
	return authorizer.Can(ctx, "org", authx.ActionRead) ||
		authorizer.Can(ctx, "budget", authx.ActionRead) ||
		authorizer.Can(ctx, "tma", authx.ActionRead) ||
		authorizer.Can(ctx, "cra", authx.ActionRead)
}

func parseApplicationActiveFilter(raw string) ports.ApplicationListFilter {
	switch raw {
	case "", "true":
		active := true
		return ports.ApplicationListFilter{Active: &active}
	case "false":
		active := false
		return ports.ApplicationListFilter{Active: &active}
	case "all":
		return ports.ApplicationListFilter{}
	default:
		active := true
		return ports.ApplicationListFilter{Active: &active}
	}
}

func listApplications(org ports.OrganizationService, authorizer authx.Authorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !canReadApplications(r.Context(), authorizer) {
			httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "forbidden")
			return
		}
		identity, _ := authx.FromContext(r.Context())
		filter := parseApplicationActiveFilter(r.URL.Query().Get("active"))
		items, err := org.ListApplications(r.Context(), identity.TenantID, filter)
		if err != nil {
			httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, err.Error())
			return
		}
		httpx.WriteData(w, http.StatusOK, items)
	}
}

func mergeApplications(org ports.OrganizationService, authorizer authx.Authorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !authorizer.Can(r.Context(), "org", authx.ActionWrite) {
			httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "forbidden")
			return
		}
		var req struct {
			SourceApplicationID uuid.UUID `json:"sourceApplicationId"`
			TargetApplicationID uuid.UUID `json:"targetApplicationId"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid body")
			return
		}
		if req.SourceApplicationID == uuid.Nil || req.TargetApplicationID == uuid.Nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid application ids")
			return
		}
		identity, _ := authx.FromContext(r.Context())
		app, err := org.MergeApplications(r.Context(), ports.MergeApplicationsCommand{
			TenantID:            identity.TenantID,
			ActorUserID:         identity.UserID,
			SourceApplicationID: req.SourceApplicationID,
			TargetApplicationID: req.TargetApplicationID,
		})
		if err != nil {
			switch {
			case errors.Is(err, domain.ErrApplicationsMergeBothTaigaLinked):
				httpx.WriteError(w, http.StatusUnprocessableEntity, httpx.ErrCodeApplicationsMergeBothTaiga, err.Error())
			case errors.Is(err, domain.ErrApplicationsMergeActiveSprintConflict):
				httpx.WriteError(w, http.StatusUnprocessableEntity, httpx.ErrCodeApplicationsMergeActiveSprint, err.Error())
			case errors.Is(err, domain.ErrApplicationsMergeMethodologyConflict):
				httpx.WriteError(w, http.StatusUnprocessableEntity, httpx.ErrCodeApplicationsMergeMethodology, err.Error())
			case errors.Is(err, domain.ErrApplicationsMergeDuplicateDefaultBudget):
				httpx.WriteError(w, http.StatusUnprocessableEntity, httpx.ErrCodeApplicationsMergeDefaultBudget, err.Error())
			case errors.Is(err, domain.ErrApplicationsMergeInactiveApplication):
				httpx.WriteError(w, http.StatusUnprocessableEntity, httpx.ErrCodeApplicationsMergeInactive, err.Error())
			case errors.Is(err, domain.ErrApplicationsMergeInvalid):
				httpx.WriteError(w, http.StatusUnprocessableEntity, httpx.ErrCodeValidation, err.Error())
			default:
				httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, err.Error())
			}
			return
		}
		httpx.WriteData(w, http.StatusOK, app)
	}
}

func listSites(org ports.OrganizationService, authorizer authx.Authorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !authorizer.Can(r.Context(), "org", authx.ActionRead) {
			httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "forbidden")
			return
		}
		identity, _ := authx.FromContext(r.Context())
		items, err := org.ListSites(r.Context(), identity.TenantID)
		if err != nil {
			httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, err.Error())
			return
		}
		httpx.WriteData(w, http.StatusOK, items)
	}
}

func createEquipe(org ports.OrganizationService, authorizer authx.Authorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !authorizer.Can(r.Context(), "org", authx.ActionWrite) {
			httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "forbidden")
			return
		}
		var req struct {
			ApplicationID uuid.UUID  `json:"applicationId"`
			Libelle       string     `json:"libelle"`
			ResponsableID *uuid.UUID `json:"responsableId"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid body")
			return
		}
		identity, _ := authx.FromContext(r.Context())
		e, err := org.CreateEquipe(r.Context(), ports.CreateEquipeCommand{
			TenantID:      identity.TenantID,
			ApplicationID: req.ApplicationID,
			Libelle:       req.Libelle,
			ResponsableID: req.ResponsableID,
		})
		if err != nil {
			if errors.Is(err, domain.ErrEquipeWithoutApplication) {
				httpx.WriteError(w, http.StatusUnprocessableEntity, httpx.ErrCodeValidation, err.Error())
				return
			}
			if errors.Is(err, domain.ErrInvalidEquipeLibelle) {
				httpx.WriteError(w, http.StatusUnprocessableEntity, httpx.ErrCodeValidation, err.Error())
				return
			}
			if errors.Is(err, domain.ErrUserNotFound) {
				httpx.WriteError(w, http.StatusUnprocessableEntity, httpx.ErrCodeValidation, err.Error())
				return
			}
			httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, err.Error())
			return
		}
		httpx.WriteData(w, http.StatusCreated, e)
	}
}

func updateEquipe(org ports.OrganizationService, authorizer authx.Authorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !authorizer.Can(r.Context(), "org", authx.ActionWrite) {
			httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "forbidden")
			return
		}
		equipeID, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid equipe id")
			return
		}
		var req struct {
			Libelle       string     `json:"libelle"`
			ResponsableID *uuid.UUID `json:"responsableId"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid body")
			return
		}
		identity, _ := authx.FromContext(r.Context())
		item, err := org.UpdateEquipe(r.Context(), ports.UpdateEquipeCommand{
			TenantID:      identity.TenantID,
			EquipeID:      equipeID,
			Libelle:       req.Libelle,
			ResponsableID: req.ResponsableID,
		})
		if err != nil {
			if errors.Is(err, domain.ErrEquipeNotFound) {
				httpx.WriteError(w, http.StatusNotFound, httpx.ErrCodeNotFound, "equipe not found")
				return
			}
			if errors.Is(err, domain.ErrInvalidEquipeLibelle) {
				httpx.WriteError(w, http.StatusUnprocessableEntity, httpx.ErrCodeValidation, err.Error())
				return
			}
			if errors.Is(err, domain.ErrUserNotFound) {
				httpx.WriteError(w, http.StatusUnprocessableEntity, httpx.ErrCodeValidation, err.Error())
				return
			}
			httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, err.Error())
			return
		}
		httpx.WriteData(w, http.StatusOK, item)
	}
}

func listEquipes(org ports.OrganizationService, authorizer authx.Authorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !authorizer.Can(r.Context(), "org", authx.ActionRead) &&
			!authorizer.Can(r.Context(), "workflow", authx.ActionRead) {
			httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "forbidden")
			return
		}
		identity, _ := authx.FromContext(r.Context())
		filter := ports.EquipeListFilter{}
		if raw := r.URL.Query().Get("applicationId"); raw != "" {
			appID, err := uuid.Parse(raw)
			if err != nil {
				httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid applicationId")
				return
			}
			filter.ApplicationID = &appID
		}
		items, err := org.ListEquipes(r.Context(), identity.TenantID, filter)
		if err != nil {
			httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, err.Error())
			return
		}
		httpx.WriteData(w, http.StatusOK, items)
	}
}

func listServices(org ports.OrganizationService, authorizer authx.Authorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !authorizer.Can(r.Context(), "org", authx.ActionRead) &&
			!authorizer.Can(r.Context(), "workflow", authx.ActionRead) {
			httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "forbidden")
			return
		}
		identity, _ := authx.FromContext(r.Context())
		items, err := org.ListServices(r.Context(), identity.TenantID)
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

func updateApplication(org ports.OrganizationService, authorizer authx.Authorizer, taigaBridge TaigaApplicationBridge) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !authorizer.Can(r.Context(), "org", authx.ActionWrite) {
			httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "forbidden")
			return
		}
		appID, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid application id")
			return
		}
		var raw map[string]json.RawMessage
		if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid body")
			return
		}
		cmd := ports.UpdateApplicationCommand{ApplicationID: appID}
		if v, ok := raw["libelle"]; ok {
			var s string
			if err := json.Unmarshal(v, &s); err != nil {
				httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid libelle")
				return
			}
			cmd.Libelle = &s
		}
		if v, ok := raw["active"]; ok {
			var b bool
			if err := json.Unmarshal(v, &b); err != nil {
				httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid active")
				return
			}
			cmd.Active = &b
		}
		if v, ok := raw["proprietaire"]; ok {
			var s string
			if err := json.Unmarshal(v, &s); err != nil {
				httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid proprietaire")
				return
			}
			cmd.Proprietaire = &s
		}
		if v, ok := raw["modeFacturation"]; ok {
			var s string
			if err := json.Unmarshal(v, &s); err != nil {
				httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid modeFacturation")
				return
			}
			cmd.ModeFacturation = &s
		}
		if v, ok := raw["uoActivee"]; ok {
			var b bool
			if err := json.Unmarshal(v, &b); err != nil {
				httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid uoActivee")
				return
			}
			cmd.UOActivee = &b
		}
		if v, ok := raw["chefUtilisateurId"]; ok {
			ptr, err := parseOptionalUUIDField(v)
			if err != nil {
				httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid chefUtilisateurId")
				return
			}
			cmd.ChefUtilisateurID = &ptr
		}
		if v, ok := raw["budgetDefautId"]; ok {
			ptr, err := parseOptionalUUIDField(v)
			if err != nil {
				httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid budgetDefautId")
				return
			}
			cmd.BudgetDefautID = &ptr
		}
		if v, ok := raw["defaultTjmCents"]; ok {
			var n int64
			if err := json.Unmarshal(v, &n); err != nil {
				httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid defaultTjmCents")
				return
			}
			cmd.DefaultTJMCents = &n
		}
		if v, ok := raw["methodologyProfile"]; ok {
			var s string
			if err := json.Unmarshal(v, &s); err != nil {
				httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid methodologyProfile")
				return
			}
			cmd.MethodologyProfile = &s
		}
		sharesTouched := false
		if v, ok := raw["siteIds"]; ok {
			ids, err := parseUUIDSliceField(v)
			if err != nil {
				httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid siteIds")
				return
			}
			cmd.SiteIDs = &ids
			sharesTouched = true
		}
		if v, ok := raw["serviceIds"]; ok {
			ids, err := parseUUIDSliceField(v)
			if err != nil {
				httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid serviceIds")
				return
			}
			cmd.ServiceIDs = &ids
			sharesTouched = true
		}
		if v, ok := raw["equipeIds"]; ok {
			ids, err := parseUUIDSliceField(v)
			if err != nil {
				httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid equipeIds")
				return
			}
			cmd.EquipeIDs = &ids
			sharesTouched = true
		}
		var taigaProjectID *int
		if v, ok := raw["taigaProjectId"]; ok {
			var n int
			if err := json.Unmarshal(v, &n); err != nil {
				httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid taigaProjectId")
				return
			}
			taigaProjectID = &n
		}
		taigaTouched := taigaProjectID != nil
		if cmd.Libelle == nil && cmd.Active == nil && cmd.Proprietaire == nil &&
			cmd.ModeFacturation == nil && cmd.UOActivee == nil &&
			cmd.ChefUtilisateurID == nil && cmd.BudgetDefautID == nil &&
			cmd.DefaultTJMCents == nil && cmd.MethodologyProfile == nil &&
			!sharesTouched && !taigaTouched {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "at least one field required")
			return
		}
		identity, _ := authx.FromContext(r.Context())
		cmd.TenantID = identity.TenantID
		// Liaison Taiga avant la mise à jour org : si la liaison échoue, aucun champ n'est modifié.
		if taigaProjectID != nil && *taigaProjectID > 0 {
			if taigaBridge == nil {
				httpx.WriteError(w, http.StatusServiceUnavailable, httpx.ErrCodeInternal, "taiga not configured")
				return
			}
			if err := taigaBridge.LinkExistingApplication(r.Context(), identity.TenantID, appID, *taigaProjectID); err != nil {
				writeApplicationTaigaError(w, err)
				return
			}
		}
		item, err := org.UpdateApplication(r.Context(), cmd)
		if err != nil {
			if errors.Is(err, domain.ErrApplicationNotFound) {
				httpx.WriteError(w, http.StatusNotFound, httpx.ErrCodeNotFound, "application not found")
				return
			}
			if errors.Is(err, domain.ErrInvalidModeFacturation) ||
				errors.Is(err, domain.ErrInvalidMethodologyProfile) ||
				errors.Is(err, domain.ErrMethodologyProfileLocked) ||
				errors.Is(err, domain.ErrUserNotFound) ||
				errors.Is(err, domain.ErrInvalidApplicationLibelle) ||
				errors.Is(err, domain.ErrBudgetNotFound) ||
				errors.Is(err, domain.ErrApplicationWithoutShare) ||
				errors.Is(err, domain.ErrInvalidApplicationShare) {
				httpx.WriteError(w, http.StatusUnprocessableEntity, httpx.ErrCodeValidation, err.Error())
				return
			}
			httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, err.Error())
			return
		}
		httpx.WriteData(w, http.StatusOK, item)
	}
}

// parseOptionalUUIDField decodes JSON null → nil pointer, or a UUID string → *uuid.
func parseOptionalUUIDField(raw json.RawMessage) (*uuid.UUID, error) {
	if string(raw) == "null" {
		return nil, nil
	}
	var id uuid.UUID
	if err := json.Unmarshal(raw, &id); err != nil {
		return nil, err
	}
	return &id, nil
}

func parseUUIDSliceField(raw json.RawMessage) ([]uuid.UUID, error) {
	if string(raw) == "null" {
		return []uuid.UUID{}, nil
	}
	var ids []uuid.UUID
	if err := json.Unmarshal(raw, &ids); err != nil {
		return nil, err
	}
	if ids == nil {
		ids = []uuid.UUID{}
	}
	return ids, nil
}

func deactivateApplication(org ports.OrganizationService, authorizer authx.Authorizer) http.HandlerFunc {
	return setApplicationActiveHandler(org, authorizer, false, "deactivated")
}

func activateApplication(org ports.OrganizationService, authorizer authx.Authorizer) http.HandlerFunc {
	return setApplicationActiveHandler(org, authorizer, true, "activated")
}

func setApplicationActiveHandler(
	org ports.OrganizationService,
	authorizer authx.Authorizer,
	active bool,
	status string,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !authorizer.Can(r.Context(), "org", authx.ActionWrite) {
			httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "forbidden")
			return
		}
		appID, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid application id")
			return
		}
		identity, _ := authx.FromContext(r.Context())
		_, err = org.SetApplicationActive(r.Context(), ports.SetApplicationActiveCommand{
			TenantID:      identity.TenantID,
			ApplicationID: appID,
			Active:        active,
		})
		if err != nil {
			if errors.Is(err, domain.ErrApplicationNotFound) {
				httpx.WriteError(w, http.StatusNotFound, httpx.ErrCodeNotFound, "application not found")
				return
			}
			httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, err.Error())
			return
		}
		httpx.WriteData(w, http.StatusOK, map[string]string{"status": status})
	}
}

func createUser(users ports.UserService, authorizer authx.Authorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !authorizer.Can(r.Context(), "org", authx.ActionWrite) {
			httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "forbidden")
			return
		}
		var req struct {
			Login     string           `json:"login"`
			Password  string           `json:"password"`
			Profile   domain.Profile   `json:"profil"`
			Profiles  []domain.Profile `json:"profils"`
			EquipeID  *uuid.UUID       `json:"equipeId"`
			EquipeIDs []uuid.UUID      `json:"equipeIds"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid body")
			return
		}
		identity, _ := authx.FromContext(r.Context())
		u, err := users.CreateUser(r.Context(), ports.CreateUserCommand{
			TenantID:  identity.TenantID,
			Login:     req.Login,
			Password:  req.Password,
			Profile:   req.Profile,
			Profiles:  req.Profiles,
			EquipeID:  req.EquipeID,
			EquipeIDs: req.EquipeIDs,
		})
		if err != nil {
			switch {
			case errors.Is(err, domain.ErrLoginAlreadyExists):
				httpx.WriteError(w, http.StatusConflict, httpx.ErrCodeConflict, err.Error())
			case errors.Is(err, domain.ErrInvalidLogin), errors.Is(err, domain.ErrWeakPassword),
				errors.Is(err, domain.ErrProfilesRequired), errors.Is(err, domain.ErrInvalidProfile),
				errors.Is(err, domain.ErrEquipeNotFound):
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

func canReadClients(ctx context.Context, authorizer authx.Authorizer) bool {
	return authorizer.Can(ctx, "org", authx.ActionRead) ||
		authorizer.Can(ctx, "cra", authx.ActionRead) ||
		authorizer.Can(ctx, "ssii", authx.ActionRead)
}

func listClients(clients ports.ClientService, authorizer authx.Authorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !canReadClients(r.Context(), authorizer) {
			httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "forbidden")
			return
		}
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
		if !canReadClients(r.Context(), authorizer) {
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
			writeClientError(w, err)
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
		var req clientWriteRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid body")
			return
		}
		identity, _ := authx.FromContext(r.Context())
		c, err := clients.CreateClient(r.Context(), req.toCreateCommand(identity.TenantID))
		if err != nil {
			writeClientError(w, err)
			return
		}
		httpx.WriteData(w, http.StatusCreated, c)
	}
}

func updateClient(clients ports.ClientService, authorizer authx.Authorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !authorizer.Can(r.Context(), "org", authx.ActionWrite) {
			httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "forbidden")
			return
		}
		clientID, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid client id")
			return
		}
		var req clientWriteRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid body")
			return
		}
		identity, _ := authx.FromContext(r.Context())
		c, err := clients.UpdateClient(r.Context(), req.toUpdateCommand(identity.TenantID, clientID))
		if err != nil {
			writeClientError(w, err)
			return
		}
		httpx.WriteData(w, http.StatusOK, c)
	}
}

// clientWriteRequest is the full-replace payload for create/update.
// Omitted optional JSON fields decode as empty strings and clear stored values on update.
type clientWriteRequest struct {
	RaisonSociale string `json:"raisonSociale"`
	TVA           string `json:"tva"`
	Pays          string `json:"pays"`
	Adresse       string `json:"adresse"`
	AdresseNumero string `json:"adresseNumero"`
	AdresseBoite  string `json:"adresseBoite"`
	CodePostal    string `json:"codePostal"`
	Ville         string `json:"ville"`
	Siret         string `json:"siret"`
}

func (req clientWriteRequest) toCreateCommand(tenant kernel.TenantID) ports.CreateClientCommand {
	return ports.CreateClientCommand{
		TenantID:      tenant,
		RaisonSociale: req.RaisonSociale,
		TVA:           req.TVA,
		Pays:          req.Pays,
		Adresse:       req.Adresse,
		AdresseNumero: req.AdresseNumero,
		AdresseBoite:  req.AdresseBoite,
		CodePostal:    req.CodePostal,
		Ville:         req.Ville,
		Siret:         req.Siret,
	}
}

func (req clientWriteRequest) toUpdateCommand(tenant kernel.TenantID, clientID uuid.UUID) ports.UpdateClientCommand {
	return ports.UpdateClientCommand{
		TenantID:      tenant,
		ClientID:      clientID,
		RaisonSociale: req.RaisonSociale,
		TVA:           req.TVA,
		Pays:          req.Pays,
		Adresse:       req.Adresse,
		AdresseNumero: req.AdresseNumero,
		AdresseBoite:  req.AdresseBoite,
		CodePostal:    req.CodePostal,
		Ville:         req.Ville,
		Siret:         req.Siret,
	}
}

func writeClientError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrInvalidClientName), errors.Is(err, domain.ErrInvalidPays):
		httpx.WriteError(w, http.StatusUnprocessableEntity, httpx.ErrCodeValidation, err.Error())
	case errors.Is(err, domain.ErrClientNotFound):
		httpx.WriteError(w, http.StatusNotFound, httpx.ErrCodeNotFound, "client not found")
	default:
		httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, "internal error")
	}
}

func replaceClientContacts(clients ports.ClientService, authorizer authx.Authorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !authorizer.Can(r.Context(), "org", authx.ActionWrite) {
			httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "forbidden")
			return
		}
		clientID, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid client id")
			return
		}
		var req struct {
			Contacts []domain.ClientContact `json:"contacts"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid body")
			return
		}
		identity, _ := authx.FromContext(r.Context())
		client, err := clients.ReplaceClientContacts(r.Context(), ports.ReplaceClientContactsCommand{
			TenantID: identity.TenantID,
			ClientID: clientID,
			Contacts: req.Contacts,
		})
		if err != nil {
			writeClientError(w, err)
			return
		}
		httpx.WriteData(w, http.StatusOK, client)
	}
}

func updateSocieteBranding(org ports.OrganizationService, authorizer authx.Authorizer) http.HandlerFunc {
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
		if err := r.ParseMultipartForm(uploads.MaxLogoBytes + (1 << 20)); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid multipart form")
			return
		}
		cmd := ports.UpdateSocieteBrandingCommand{
			TenantID:      identity.TenantID,
			SocieteID:     societeID,
			RaisonSociale: r.FormValue("raisonSociale"),
			Adresse:       r.FormValue("adresse"),
			AdresseNumero: r.FormValue("adresseNumero"),
			AdresseBoite:  r.FormValue("adresseBoite"),
			CodePostal:    r.FormValue("codePostal"),
			Ville:         r.FormValue("ville"),
			Pays:          r.FormValue("pays"),
			Siret:         r.FormValue("siret"),
		}
		if file, header, err := r.FormFile("logo"); err == nil {
			defer file.Close()
			data, err := uploads.ReadAndValidateLogo(file, header.Filename)
			if err != nil {
				writeUploadError(w, err)
				return
			}
			cmd.Logo = fmt.Sprintf("/api/v1/branding/logo/%s", identity.TenantID.UUID().String())
			cmd.LogoContent = data
			cmd.LogoContentType = uploads.ContentTypeForExt(header.Filename)
		}
		societe, err := org.UpdateSocieteBranding(r.Context(), cmd)
		if err != nil {
			httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, err.Error())
			return
		}
		httpx.WriteData(w, http.StatusOK, societe)
	}
}

func updateSocieteSettings(org ports.OrganizationService, authorizer authx.Authorizer) http.HandlerFunc {
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
		var req struct {
			WeekStartDay         *int      `json:"weekStartDay"`
			DayCapacityMinutes   *int      `json:"dayCapacityMinutes"`
			CraMailAuto          *bool     `json:"craMailAuto"`
			CraMailRecipients    *[]string `json:"craMailRecipients"`
			WeekSubmitPolicy     *string   `json:"weekSubmitPolicy"`
			CraGateMode          *string   `json:"craGateMode"`
			TaskTypesEnabled     *[]string `json:"taskTypesEnabled"`
			TotpDefaultEnabled   *bool     `json:"totpDefaultEnabled"`
			TotpUserConfigurable *bool     `json:"totpUserConfigurable"`
			DefaultTJMCents      *int64    `json:"defaultTjmCents"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid body")
			return
		}
		identity, _ := authx.FromContext(r.Context())
		societe, err := org.UpdateSocieteSettings(r.Context(), ports.UpdateSocieteSettingsCommand{
			TenantID:             identity.TenantID,
			SocieteID:            societeID,
			WeekStartDay:         req.WeekStartDay,
			DayCapacityMinutes:   req.DayCapacityMinutes,
			CraMailAuto:          req.CraMailAuto,
			CraMailRecipients:    req.CraMailRecipients,
			WeekSubmitPolicy:     req.WeekSubmitPolicy,
			CraGateMode:          req.CraGateMode,
			TaskTypesEnabled:     req.TaskTypesEnabled,
			TotpDefaultEnabled:   req.TotpDefaultEnabled,
			TotpUserConfigurable: req.TotpUserConfigurable,
			DefaultTJMCents:      req.DefaultTJMCents,
		})
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, err.Error())
			return
		}
		httpx.WriteData(w, http.StatusOK, societe)
	}
}

func serveTenantLogo(org ports.OrganizationService, uploadsDir string) http.HandlerFunc {
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
		content, contentType, err := org.GetTenantLogo(r.Context(), identity.TenantID)
		if err == nil {
			w.Header().Set("Content-Type", contentType)
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("Cache-Control", "private, max-age=3600")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(content)
			return
		}
		if !errors.Is(err, domain.ErrLogoNotFound) {
			httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, err.Error())
			return
		}
		// Fallback: legacy filesystem logos (local docker volume).
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
			Profile   *domain.Profile   `json:"profil"`
			Profiles  *[]domain.Profile `json:"profils"`
			Password  string            `json:"password"`
			Active    *bool             `json:"active"`
			EquipeID  *string           `json:"equipeId"`
			EquipeIDs *[]uuid.UUID      `json:"equipeIds"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid body")
			return
		}
		// equipeId absent = rattachement inchangé ; chaîne vide = détachement.
		// Prefer equipeIds when present.
		var equipeID **uuid.UUID
		if req.EquipeIDs == nil && req.EquipeID != nil {
			var parsed *uuid.UUID
			if *req.EquipeID != "" {
				id, err := uuid.Parse(*req.EquipeID)
				if err != nil {
					httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid equipe id")
					return
				}
				parsed = &id
			}
			equipeID = &parsed
		}
		identity, _ := authx.FromContext(r.Context())
		summary, err := users.UpdateUser(r.Context(), ports.UpdateUserCommand{
			TenantID:    identity.TenantID,
			UserID:      userID,
			ActorUserID: identity.UserID,
			Profile:     req.Profile,
			Profiles:    req.Profiles,
			Password:    req.Password,
			Active:      req.Active,
			EquipeID:    equipeID,
			EquipeIDs:   req.EquipeIDs,
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

func getMeProfile(users ports.UserService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		identity, ok := authx.FromContext(r.Context())
		if !ok {
			httpx.WriteError(w, http.StatusUnauthorized, httpx.ErrCodeUnauthorized, "unauthorized")
			return
		}
		detail, err := users.GetUser(r.Context(), identity.TenantID, identity.UserID)
		if err != nil {
			writeUserMutationError(w, err)
			return
		}
		httpx.WriteData(w, http.StatusOK, map[string]any{
			"id":         detail.ID,
			"login":      detail.Login,
			"prenom":     detail.Prenom,
			"nom":        detail.Nom,
			"craRequis":  detail.CraRequis,
			"salarieEtt": detail.SalarieETT,
		})
	}
}

func getUserCalendarSettings(org ports.OrganizationService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		identity, _ := authx.FromContext(r.Context())
		settings, err := org.CalendarSettingsForUser(r.Context(), identity.TenantID, identity.UserID)
		if err != nil {
			httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, err.Error())
			return
		}
		httpx.WriteData(w, http.StatusOK, settings)
	}
}

func getReleaseNotesPreferences(users ports.UserService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		identity, _ := authx.FromContext(r.Context())
		prefs, err := users.GetReleaseNotesPreferences(r.Context(), identity.TenantID, identity.UserID)
		if err != nil {
			httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, err.Error())
			return
		}
		httpx.WriteData(w, http.StatusOK, prefs)
	}
}

func setReleaseNotesAutoShow(users ports.UserService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		identity, _ := authx.FromContext(r.Context())
		var req struct {
			Enabled bool `json:"enabled"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid body")
			return
		}
		if err := users.SetReleaseNotesAutoShow(r.Context(), identity.TenantID, identity.UserID, req.Enabled); err != nil {
			httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, err.Error())
			return
		}
		httpx.WriteData(w, http.StatusOK, map[string]any{"status": "ok"})
	}
}

func markReleaseNotesSeen(users ports.UserService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		identity, _ := authx.FromContext(r.Context())
		var req struct {
			Version string `json:"version"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid body")
			return
		}
		if err := users.MarkReleaseNotesSeen(r.Context(), identity.TenantID, identity.UserID, req.Version); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, err.Error())
			return
		}
		httpx.WriteData(w, http.StatusOK, map[string]any{"status": "ok"})
	}
}

func writeUserMutationError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrUserNotFound):
		httpx.WriteError(w, http.StatusNotFound, httpx.ErrCodeNotFound, err.Error())
	case errors.Is(err, domain.ErrCannotModifySelf),
		errors.Is(err, domain.ErrCannotDemoteSelf),
		errors.Is(err, domain.ErrLastAdmin):
		httpx.WriteError(w, http.StatusUnprocessableEntity, httpx.ErrCodeValidation, err.Error())
	case errors.Is(err, domain.ErrWeakPassword), errors.Is(err, domain.ErrProfilesRequired),
		errors.Is(err, domain.ErrInvalidProfile), errors.Is(err, domain.ErrEquipeNotFound):
		httpx.WriteError(w, http.StatusUnprocessableEntity, httpx.ErrCodeValidation, err.Error())
	case errors.Is(err, domain.ErrSeatLimitReached):
		httpx.WriteError(w, http.StatusConflict, httpx.ErrCodeConflict, err.Error())
	default:
		httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, err.Error())
	}
}

var _ = app.DefaultPermissions
var _ = kernel.TenantID{}
