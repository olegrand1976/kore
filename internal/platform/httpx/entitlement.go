package httpx

import (
	"net/http"
	"strings"

	"github.com/kore/kore/internal/platform/authx"
)

var projectApplicationSubresources = map[string]struct{}{
	"epics":         {},
	"sprints":       {},
	"backlog":       {},
	"kanban-config": {},
	"velocity":      {},
}

func isProjectApplicationPath(parts []string) bool {
	if len(parts) < 3 || parts[0] != "applications" {
		return false
	}
	_, ok := projectApplicationSubresources[parts[2]]
	return ok
}

// ModuleFromPath maps API path prefixes to entitlement module codes.
func ModuleFromPath(path string) (authx.Module, bool) {
	path = strings.TrimPrefix(path, "/api/v1")
	path = strings.TrimPrefix(path, "/")
	if path == "" {
		return "", false
	}
	parts := strings.Split(path, "/")
	segment := parts[0]
	switch segment {
	case "project":
		return "project", true
	case "applications":
		if isProjectApplicationPath(parts) {
			return "project", true
		}
		return "org", true
	case "societes", "sites", "services", "users", "clients", "branding":
		return "org", true
	case "notification-rules", "notifications":
		return "notifications", true
	case "workflows", "workflow-instances":
		return "workflow", true
	case "timesheets":
		return "cra", true
	case "leave-requests", "leave-balances":
		return "conges", true
	case "budgets":
		return "budget", true
	case "demands":
		return "tma", true
	case "billing":
		return "billing", true
	default:
		return "", false
	}
}

func EntitlementMiddleware(reader authx.EntitlementReader) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if reader == nil {
				next.ServeHTTP(w, r)
				return
			}
			module, ok := ModuleFromPath(r.URL.Path)
			if !ok {
				next.ServeHTTP(w, r)
				return
			}
			identity, ok := authx.FromContext(r.Context())
			if !ok {
				next.ServeHTTP(w, r)
				return
			}
			enabled, err := reader.IsModuleEnabled(r.Context(), identity.TenantID, module)
			if err != nil {
				WriteError(w, http.StatusInternalServerError, ErrCodeInternal, err.Error())
				return
			}
			if !enabled && module != "org" && module != "billing" {
				WriteError(w, http.StatusPaymentRequired, ErrCodePaymentRequired, "module not subscribed")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
