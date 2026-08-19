package httpx

import "testing"

func TestModuleFromPath(t *testing.T) {
	tests := []struct {
		path   string
		module string
		ok     bool
	}{
		{"/api/v1/project/applications", "project", true},
		{"/api/v1/applications/00000000-0000-0000-0000-000000000001/epics", "project", true},
		{"/api/v1/applications/00000000-0000-0000-0000-000000000001/sprints/00000000-0000-0000-0000-000000000002/start", "project", true},
		{"/api/v1/applications/00000000-0000-0000-0000-000000000001/backlog/reorder", "project", true},
		{"/api/v1/applications/00000000-0000-0000-0000-000000000001/velocity", "project", true},
		{"/api/v1/applications/00000000-0000-0000-0000-000000000001/kanban-config", "project", true},
		{"/api/v1/applications", "org", true},
		{"/api/v1/applications/00000000-0000-0000-0000-000000000001", "org", true},
		{"/api/v1/demands", "tma", true},
		{"/api/v1/unknown", "", false},
	}
	for _, tc := range tests {
		got, ok := ModuleFromPath(tc.path)
		if ok != tc.ok {
			t.Fatalf("path %q: ok = %v, want %v", tc.path, ok, tc.ok)
		}
		if string(got) != tc.module {
			t.Fatalf("path %q: module = %q, want %q", tc.path, got, tc.module)
		}
	}
}
