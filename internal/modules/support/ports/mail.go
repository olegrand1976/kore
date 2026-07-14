package ports

import (
	"context"

	"github.com/google/uuid"
)

type InboundEmail struct {
	From       string
	Subject    string
	Body       string
	ReporterID *uuid.UUID
}

type InboundMailGateway interface {
	Poll(ctx context.Context) ([]InboundEmail, error)
}
