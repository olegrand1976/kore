package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/ai/domain"
	"github.com/kore/kore/internal/modules/ai/ports"
	"github.com/kore/kore/internal/platform/authx"
	"github.com/kore/kore/internal/platform/httpx"
)

func RegisterRoutes(r chi.Router, ai ports.AIService, tokens *authx.TokenIssuer, authorizer authx.Authorizer, entitlements authx.EntitlementReader) {
	r.Group(func(pr chi.Router) {
		pr.Use(httpx.AuthStack(tokens, entitlements))
		pr.Post("/ai/tma/analysis-draft", analysisDraft(ai))
		pr.Post("/ai/tma/classify", classifyDemand(ai))
		pr.Get("/ai/tma/similar", similarDemands(ai))
		pr.Post("/ai/tma/suggest-assignee", suggestAssignee(ai))
		pr.Post("/ai/tma/executive-summary", executiveSummary(ai))
		pr.Post("/ai/cra/prefill-suggest", craPrefill(ai))
		pr.Get("/ai/cra/anomalies", craAnomalies(ai))
		pr.Post("/ai/cra/comment-summary", craCommentSummary(ai))
		pr.Post("/ai/budget/estimate-effort", budgetEstimate(ai))
		pr.Get("/ai/budget/demand-suggest", budgetDemandSuggest(ai))
		pr.Get("/ai/budget/overrun-forecast", budgetOverrunForecast(ai))
		pr.Get("/ai/dashboard/briefing", dashboardBriefing(ai))
		pr.Post("/ai/conges/manager-context", congesManagerContext(ai))
		pr.Post("/ai/conges/date-suggest", congesDateSuggest(ai))
		pr.Get("/ai/workflow/explain", workflowExplain(ai, authorizer))
		pr.Post("/ai/notifications/digest", notificationsDigest(ai))
		pr.Post("/ai/mobile/voice-cra", voiceCraTranscribe(ai))
		pr.Get("/ai/explain/{requestId}", explainRequest(ai))
		pr.Get("/ai/settings", getSettings(ai))
		pr.Post("/ai/settings/enable", enableAI(ai, authorizer))
	})
	r.Post("/ai/public/chat", publicChat(ai))
	r.Post("/ai/public/lead-scoring", leadScoring(ai))
}

func aiError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrAIDisabled), errors.Is(err, domain.ErrCapabilityOff):
		httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, err.Error())
	case errors.Is(err, domain.ErrRequestNotFound):
		httpx.WriteError(w, http.StatusNotFound, httpx.ErrCodeNotFound, err.Error())
	case errors.Is(err, domain.ErrPromptInjectionBlocked):
		httpx.WriteError(w, http.StatusUnprocessableEntity, httpx.ErrCodeValidation, err.Error())
	default:
		httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, err.Error())
	}
}

func analysisDraft(ai ports.AIService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		identity, _ := authx.FromContext(r.Context())
		var body struct {
			DemandID      string `json:"demandId"`
			Subject       string `json:"subject"`
			ApplicationID string `json:"applicationId"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid body")
			return
		}
		demandID, _ := uuid.Parse(body.DemandID)
		appID, _ := uuid.Parse(body.ApplicationID)
		result, err := ai.SuggestAnalysisDraft(r.Context(), ports.AnalysisDraftCommand{
			TenantID: identity.TenantID, UserID: identity.UserID,
			DemandID: demandID, Subject: body.Subject, ApplicationID: appID,
		})
		if err != nil {
			aiError(w, err)
			return
		}
		httpx.WriteData(w, http.StatusOK, result)
	}
}

func classifyDemand(ai ports.AIService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		identity, _ := authx.FromContext(r.Context())
		var body struct {
			Subject string `json:"subject"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid body")
			return
		}
		result, err := ai.ClassifyDemand(r.Context(), ports.ClassifyDemandCommand{
			TenantID: identity.TenantID, UserID: identity.UserID, Subject: body.Subject,
		})
		if err != nil {
			aiError(w, err)
			return
		}
		httpx.WriteData(w, http.StatusOK, result)
	}
}

func similarDemands(ai ports.AIService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		identity, _ := authx.FromContext(r.Context())
		appID, _ := uuid.Parse(r.URL.Query().Get("applicationId"))
		var appPtr *uuid.UUID
		if appID != uuid.Nil {
			appPtr = &appID
		}
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		items, err := ai.FindSimilarDemands(r.Context(), ports.SimilarDemandsCommand{
			TenantID: identity.TenantID, UserID: identity.UserID,
			Subject: r.URL.Query().Get("subject"), ApplicationID: appPtr, Limit: limit,
		})
		if err != nil {
			aiError(w, err)
			return
		}
		httpx.WriteData(w, http.StatusOK, items)
	}
}

func craPrefill(ai ports.AIService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		identity, _ := authx.FromContext(r.Context())
		var body struct {
			TimesheetID string `json:"timesheetId"`
			WeekNumber  int    `json:"weekNumber"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid body")
			return
		}
		tsID, err := uuid.Parse(body.TimesheetID)
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid timesheetId")
			return
		}
		result, err := ai.SuggestCraPrefill(r.Context(), ports.CraPrefillCommand{
			TenantID: identity.TenantID, UserID: identity.UserID,
			TimesheetID: tsID, WeekNumber: body.WeekNumber,
		})
		if err != nil {
			aiError(w, err)
			return
		}
		httpx.WriteData(w, http.StatusOK, result)
	}
}

func craAnomalies(ai ports.AIService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		identity, _ := authx.FromContext(r.Context())
		tsID, err := uuid.Parse(r.URL.Query().Get("timesheetId"))
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid timesheetId")
			return
		}
		items, err := ai.ListCraAnomalies(r.Context(), ports.CraAnomaliesCommand{
			TenantID: identity.TenantID, UserID: identity.UserID, TimesheetID: tsID,
		})
		if err != nil {
			aiError(w, err)
			return
		}
		httpx.WriteData(w, http.StatusOK, items)
	}
}

func budgetEstimate(ai ports.AIService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		identity, _ := authx.FromContext(r.Context())
		var body struct {
			DemandID string `json:"demandId"`
			BudgetID string `json:"budgetId"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid body")
			return
		}
		demandID, _ := uuid.Parse(body.DemandID)
		budgetID, _ := uuid.Parse(body.BudgetID)
		result, err := ai.EstimateBudgetEffort(r.Context(), ports.BudgetEstimateCommand{
			TenantID: identity.TenantID, UserID: identity.UserID,
			DemandID: demandID, BudgetID: budgetID,
		})
		if err != nil {
			aiError(w, err)
			return
		}
		httpx.WriteData(w, http.StatusOK, result)
	}
}

func budgetDemandSuggest(ai ports.AIService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		identity, _ := authx.FromContext(r.Context())
		budgetID, _ := uuid.Parse(r.URL.Query().Get("budgetId"))
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		items, err := ai.SuggestBudgetDemands(r.Context(), ports.BudgetDemandSuggestCommand{
			TenantID: identity.TenantID, UserID: identity.UserID,
			BudgetID: budgetID, Query: r.URL.Query().Get("q"), Limit: limit,
		})
		if err != nil {
			aiError(w, err)
			return
		}
		httpx.WriteData(w, http.StatusOK, items)
	}
}

func dashboardBriefing(ai ports.AIService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		identity, _ := authx.FromContext(r.Context())
		q := r.URL.Query()
		leavePending, _ := strconv.Atoi(q.Get("leavePending"))
		tmaOpen, _ := strconv.Atoi(q.Get("tmaOpen"))
		budgetOverrun, _ := strconv.Atoi(q.Get("budgetOverrun"))
		pendingValidations, _ := strconv.Atoi(q.Get("pendingValidations"))
		budgetConsumption, _ := strconv.ParseFloat(q.Get("budgetConsumption"), 64)
		result, err := ai.DashboardBriefing(r.Context(), ports.DashboardBriefingCommand{
			TenantID: identity.TenantID, UserID: identity.UserID,
			Profile:   string(identity.Profile),
			CraStatus: q.Get("craStatus"), LeavePending: leavePending, TmaOpen: tmaOpen,
			BudgetConsumption: budgetConsumption, BudgetOverrun: budgetOverrun,
			PendingValidations: pendingValidations,
		})
		if err != nil {
			aiError(w, err)
			return
		}
		httpx.WriteData(w, http.StatusOK, result)
	}
}

func congesManagerContext(ai ports.AIService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		identity, _ := authx.FromContext(r.Context())
		var body struct {
			LeaveRequestID string `json:"leaveRequestId"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid body")
			return
		}
		leaveID, err := uuid.Parse(body.LeaveRequestID)
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid leaveRequestId")
			return
		}
		result, err := ai.CongesManagerContext(r.Context(), ports.CongesManagerCommand{
			TenantID: identity.TenantID, UserID: identity.UserID, LeaveRequestID: leaveID,
		})
		if err != nil {
			aiError(w, err)
			return
		}
		httpx.WriteData(w, http.StatusOK, result)
	}
}

func workflowExplain(ai ports.AIService, _ authx.Authorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		identity, _ := authx.FromContext(r.Context())
		instanceID, err := uuid.Parse(r.URL.Query().Get("instanceId"))
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid instanceId")
			return
		}
		result, err := ai.ExplainWorkflow(r.Context(), ports.WorkflowExplainCommand{
			TenantID: identity.TenantID, UserID: identity.UserID, InstanceID: instanceID,
		})
		if err != nil {
			aiError(w, err)
			return
		}
		httpx.WriteData(w, http.StatusOK, result)
	}
}

func explainRequest(ai ports.AIService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		identity, _ := authx.FromContext(r.Context())
		requestID, err := uuid.Parse(chi.URLParam(r, "requestId"))
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid requestId")
			return
		}
		result, err := ai.ExplainRequest(r.Context(), identity.TenantID, requestID)
		if err != nil {
			aiError(w, err)
			return
		}
		httpx.WriteData(w, http.StatusOK, result)
	}
}

func getSettings(ai ports.AIService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		identity, _ := authx.FromContext(r.Context())
		settings, err := ai.GetTenantSettings(r.Context(), identity.TenantID)
		if err != nil {
			aiError(w, err)
			return
		}
		httpx.WriteData(w, http.StatusOK, settings)
	}
}

func enableAI(ai ports.AIService, authorizer authx.Authorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !authorizer.Can(r.Context(), "org", authx.ActionWrite) {
			httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "admin required")
			return
		}
		identity, _ := authx.FromContext(r.Context())
		var body struct {
			NoticeAccepted  bool `json:"noticeAccepted"`
			WorkersInformed bool `json:"workersInformed"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid body")
			return
		}
		if err := ai.EnableAI(r.Context(), ports.EnableAICommand{
			TenantID: identity.TenantID, UserID: identity.UserID,
			NoticeAccepted: body.NoticeAccepted, WorkersInformed: body.WorkersInformed,
		}); err != nil {
			aiError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func publicChat(ai ports.AIService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Message   string `json:"message"`
			SessionID string `json:"sessionId"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid body")
			return
		}
		result, err := ai.PublicChat(r.Context(), ports.PublicChatCommand{
			Message: body.Message, SessionID: body.SessionID,
		})
		if err != nil {
			aiError(w, err)
			return
		}
		httpx.WriteData(w, http.StatusOK, result)
	}
}

func suggestAssignee(ai ports.AIService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		identity, _ := authx.FromContext(r.Context())
		var body struct {
			DemandID      string `json:"demandId"`
			ApplicationID string `json:"applicationId"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid body")
			return
		}
		demandID, _ := uuid.Parse(body.DemandID)
		appID, _ := uuid.Parse(body.ApplicationID)
		result, err := ai.SuggestAssignee(r.Context(), ports.SuggestAssigneeCommand{
			TenantID: identity.TenantID, UserID: identity.UserID,
			DemandID: demandID, ApplicationID: appID,
		})
		if err != nil {
			aiError(w, err)
			return
		}
		httpx.WriteData(w, http.StatusOK, result)
	}
}

func executiveSummary(ai ports.AIService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		identity, _ := authx.FromContext(r.Context())
		var body struct {
			ApplicationID string `json:"applicationId"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		var appPtr *uuid.UUID
		if appID, err := uuid.Parse(body.ApplicationID); err == nil && appID != uuid.Nil {
			appPtr = &appID
		}
		result, err := ai.ExecutiveSummary(r.Context(), ports.ExecutiveSummaryCommand{
			TenantID: identity.TenantID, UserID: identity.UserID, ApplicationID: appPtr,
		})
		if err != nil {
			aiError(w, err)
			return
		}
		httpx.WriteData(w, http.StatusOK, result)
	}
}

func craCommentSummary(ai ports.AIService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		identity, _ := authx.FromContext(r.Context())
		var body struct {
			TimesheetID string `json:"timesheetId"`
			WeekNumber  int    `json:"weekNumber"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid body")
			return
		}
		tsID, err := uuid.Parse(body.TimesheetID)
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid timesheetId")
			return
		}
		result, err := ai.SummarizeCraComments(r.Context(), ports.CommentSummaryCommand{
			TenantID: identity.TenantID, UserID: identity.UserID,
			TimesheetID: tsID, WeekNumber: body.WeekNumber,
		})
		if err != nil {
			aiError(w, err)
			return
		}
		httpx.WriteData(w, http.StatusOK, result)
	}
}

func budgetOverrunForecast(ai ports.AIService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		identity, _ := authx.FromContext(r.Context())
		budgetID, err := uuid.Parse(r.URL.Query().Get("budgetId"))
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid budgetId")
			return
		}
		result, err := ai.ForecastBudgetOverrun(r.Context(), ports.OverrunForecastCommand{
			TenantID: identity.TenantID, UserID: identity.UserID, BudgetID: budgetID,
		})
		if err != nil {
			aiError(w, err)
			return
		}
		httpx.WriteData(w, http.StatusOK, result)
	}
}

func congesDateSuggest(ai ports.AIService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		identity, _ := authx.FromContext(r.Context())
		var body struct {
			LeaveType    string `json:"leaveType"`
			DurationDays int    `json:"durationDays"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid body")
			return
		}
		result, err := ai.SuggestLeaveDates(r.Context(), ports.DateSuggestCommand{
			TenantID: identity.TenantID, UserID: identity.UserID,
			LeaveType: body.LeaveType, DurationDays: body.DurationDays,
		})
		if err != nil {
			aiError(w, err)
			return
		}
		httpx.WriteData(w, http.StatusOK, result)
	}
}

func notificationsDigest(ai ports.AIService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		identity, _ := authx.FromContext(r.Context())
		var body struct {
			Period         string `json:"period"`
			UnreadTma      int    `json:"unreadTma"`
			UnreadLeave    int    `json:"unreadLeave"`
			UnreadCra      int    `json:"unreadCra"`
			UnreadWorkflow int    `json:"unreadWorkflow"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid body")
			return
		}
		result, err := ai.DigestNotifications(r.Context(), ports.NotificationsDigestCommand{
			TenantID: identity.TenantID, UserID: identity.UserID, Period: body.Period,
			UnreadTma: body.UnreadTma, UnreadLeave: body.UnreadLeave,
			UnreadCra: body.UnreadCra, UnreadWorkflow: body.UnreadWorkflow,
		})
		if err != nil {
			aiError(w, err)
			return
		}
		httpx.WriteData(w, http.StatusOK, result)
	}
}

func voiceCraTranscribe(ai ports.AIService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		identity, _ := authx.FromContext(r.Context())
		var body struct {
			TimesheetID string `json:"timesheetId"`
			WeekNumber  int    `json:"weekNumber"`
			Transcript  string `json:"transcript"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid body")
			return
		}
		tsID, _ := uuid.Parse(body.TimesheetID)
		result, err := ai.TranscribeVoiceCra(r.Context(), ports.VoiceCraCommand{
			TenantID: identity.TenantID, UserID: identity.UserID,
			TimesheetID: tsID, WeekNumber: body.WeekNumber, Transcript: body.Transcript,
		})
		if err != nil {
			aiError(w, err)
			return
		}
		httpx.WriteData(w, http.StatusOK, result)
	}
}

func leadScoring(ai ports.AIService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			CompanySize string   `json:"companySize"`
			Modules     []string `json:"modules"`
			Email       string   `json:"email"`
			Message     string   `json:"message"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid body")
			return
		}
		result, err := ai.ScoreLead(r.Context(), ports.LeadScoringCommand{
			CompanySize: body.CompanySize, Modules: body.Modules, Email: body.Email, Message: body.Message,
		})
		if err != nil {
			aiError(w, err)
			return
		}
		httpx.WriteData(w, http.StatusOK, result)
	}
}
