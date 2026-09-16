package ports

import "context"

type TaigaProject struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

// TaigaIssue is a Taiga issue (incident) used for TMA demand sync.
type TaigaIssue struct {
	ID                int
	Ref               int
	ProjectID         int
	Subject           string
	Description       string
	Version           int
	ExternalReference []string
	Permalink         string
}

type TaigaGateway interface {
	ListProjects(ctx context.Context) ([]TaigaProject, error)
	CreateIssue(ctx context.Context, projectID int, subject, description string, externalRef []string) (TaigaIssue, error)
	ListProjectIssues(ctx context.Context, projectID int) ([]TaigaIssue, error)
	UpdateIssueExternalReference(ctx context.Context, issueID int, version int, externalRef []string) (TaigaIssue, error)
}
