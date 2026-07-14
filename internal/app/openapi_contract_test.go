package app

import (
	"net/http"
	"os"
	"sort"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	adminhttp "github.com/kore/kore/internal/modules/admin/adapters/http"
	billinghttp "github.com/kore/kore/internal/modules/billing/adapters/http"
	budgethttp "github.com/kore/kore/internal/modules/budget/adapters/http"
	congeshttp "github.com/kore/kore/internal/modules/conges/adapters/http"
	crahttp "github.com/kore/kore/internal/modules/cra/adapters/http"
	etthttp "github.com/kore/kore/internal/modules/ett/adapters/http"
	integrationshttp "github.com/kore/kore/internal/modules/integrations/adapters/http"
	invoicinghttp "github.com/kore/kore/internal/modules/invoicing/adapters/http"
	maintenancehttp "github.com/kore/kore/internal/modules/maintenance/adapters/http"
	notifhttp "github.com/kore/kore/internal/modules/notifications/adapters/http"
	orghttp "github.com/kore/kore/internal/modules/org/adapters/http"
	orgapp "github.com/kore/kore/internal/modules/org/app"
	publichttp "github.com/kore/kore/internal/modules/publicsite/adapters/http"
	reportinghttp "github.com/kore/kore/internal/modules/reporting/adapters/http"
	ssiihttp "github.com/kore/kore/internal/modules/ssii/adapters/http"
	supporthttp "github.com/kore/kore/internal/modules/support/adapters/http"
	tmahttp "github.com/kore/kore/internal/modules/tma/adapters/http"
	wfhttp "github.com/kore/kore/internal/modules/workflow/adapters/http"
	"gopkg.in/yaml.v3"
)

// buildAPIRouter registers every module's routes with nil dependencies.
// Registration only wires closures, so handlers are never invoked here — we
// only need the route tree to compare it against the OpenAPI spec.
func buildAPIRouter() chi.Router {
	r := chi.NewRouter()
	r.Route("/api/v1", func(r chi.Router) {
		orghttp.RegisterRoutes(r, nil, nil, nil, nil, nil, nil, "", orgapp.NoopAttachmentService{}, nil, nil)
		orghttp.RegisterOIDCRoutes(r, nil, nil, nil)
		notifhttp.RegisterRoutes(r, nil, nil, nil, nil)
		wfhttp.RegisterRoutes(r, nil, nil, nil, nil)
		crahttp.RegisterRoutes(r, nil, nil, nil, nil)
		congeshttp.RegisterRoutes(r, nil, nil, nil, nil, nil)
		budgethttp.RegisterRoutes(r, nil, nil, nil, nil)
		tmahttp.RegisterRoutes(r, nil, nil, nil, nil)
		billinghttp.RegisterRoutes(r, nil, nil, nil, "", nil)
		publichttp.RegisterRoutes(r, nil, nil, nil)
		integrationshttp.RegisterRoutes(r, nil, nil, nil, nil, nil)
		invoicinghttp.RegisterRoutes(r, nil, nil, nil, nil)
		adminhttp.RegisterRoutes(r, nil, nil, nil, nil)
		reportinghttp.RegisterRoutes(r, nil, nil, nil, nil)
		ssiihttp.RegisterRoutes(r, nil, nil, nil, nil)
		etthttp.RegisterRoutes(r, nil, nil, nil, nil)
		supporthttp.RegisterRoutes(r, nil, nil, nil, nil)
		maintenancehttp.RegisterRoutes(r, nil, nil, nil, nil)
	})
	return r
}

// normalizeRoute strips the /api/v1 prefix and any trailing slash so router
// patterns line up with the server-relative paths declared in the spec.
func normalizeRoute(route string) string {
	route = strings.TrimPrefix(route, "/api/v1")
	if len(route) > 1 {
		route = strings.TrimSuffix(route, "/")
	}
	return route
}

func routerEndpoints(t *testing.T) map[string]bool {
	t.Helper()
	endpoints := map[string]bool{}
	err := chi.Walk(buildAPIRouter(), func(method, route string, _ http.Handler, _ ...func(http.Handler) http.Handler) error {
		endpoints[strings.ToUpper(method)+" "+normalizeRoute(route)] = true
		return nil
	})
	if err != nil {
		t.Fatalf("walk router: %v", err)
	}
	return endpoints
}

func specEndpoints(t *testing.T) map[string]bool {
	t.Helper()
	data, err := os.ReadFile("openapi.yaml")
	if err != nil {
		t.Fatalf("read openapi.yaml: %v", err)
	}
	var doc struct {
		Paths map[string]map[string]any `yaml:"paths"`
	}
	if err := yaml.Unmarshal(data, &doc); err != nil {
		t.Fatalf("parse openapi.yaml: %v", err)
	}
	endpoints := map[string]bool{}
	for path, ops := range doc.Paths {
		// /health and /ready are mounted at the root, not under /api/v1.
		if path == "/health" || path == "/ready" {
			continue
		}
		for method := range ops {
			m := strings.ToUpper(method)
			switch m {
			case "GET", "POST", "PUT", "PATCH", "DELETE":
				endpoints[m+" "+path] = true
			}
		}
	}
	return endpoints
}

func sortedKeys(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// TestOpenAPIContract detects drift between the chi router and openapi.yaml:
// every documented path must exist as a route and every route must be
// documented. Health/readiness probes are handled separately.
func TestOpenAPIContract(t *testing.T) {
	routes := routerEndpoints(t)
	spec := specEndpoints(t)

	var undocumented []string
	for route := range routes {
		if !spec[route] {
			undocumented = append(undocumented, route)
		}
	}
	sort.Strings(undocumented)

	var missingRoute []string
	for path := range spec {
		if !routes[path] {
			missingRoute = append(missingRoute, path)
		}
	}
	sort.Strings(missingRoute)

	if len(undocumented) > 0 {
		t.Errorf("routes without an OpenAPI entry (add them to openapi.yaml):\n  %s", strings.Join(undocumented, "\n  "))
	}
	if len(missingRoute) > 0 {
		t.Errorf("OpenAPI paths without a matching route (remove or fix in openapi.yaml):\n  %s", strings.Join(missingRoute, "\n  "))
	}

	if t.Failed() {
		t.Logf("router endpoints:\n  %s", strings.Join(sortedKeys(routes), "\n  "))
		t.Logf("spec endpoints:\n  %s", strings.Join(sortedKeys(spec), "\n  "))
	}
}
