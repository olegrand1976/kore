package ports

import "context"

type TaigaProject struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type TaigaGateway interface {
	ListProjects(ctx context.Context) ([]TaigaProject, error)
}
