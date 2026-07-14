package app

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/org/domain"
	"github.com/kore/kore/internal/modules/org/ports"
	"github.com/kore/kore/pkg/kernel"
)

// NoopAttachmentService registers attachment routes in contract tests without a database.
type NoopAttachmentService struct{}

func (NoopAttachmentService) List(context.Context, kernel.TenantID, string, uuid.UUID) ([]domain.RequestAttachment, error) {
	return nil, nil
}

func (NoopAttachmentService) Create(context.Context, ports.CreateAttachmentCommand) (domain.RequestAttachment, error) {
	return domain.RequestAttachment{ID: uuid.New(), CreatedAt: time.Now().UTC()}, nil
}

func (NoopAttachmentService) Get(context.Context, kernel.TenantID, uuid.UUID) (domain.RequestAttachment, error) {
	return domain.RequestAttachment{}, domain.ErrAttachmentNotFound
}

var _ ports.AttachmentService = NoopAttachmentService{}
