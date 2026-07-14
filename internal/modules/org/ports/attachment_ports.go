package ports

import (
	"context"
	"io"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/org/domain"
	"github.com/kore/kore/pkg/kernel"
)

type CreateAttachmentCommand struct {
	TenantID     kernel.TenantID
	ResourceType string
	ResourceID   uuid.UUID
	FileName     string
	MimeType     string
	Content      io.Reader
	UploadedBy   uuid.UUID
	UploadsDir   string
}

type AttachmentService interface {
	List(ctx context.Context, tenant kernel.TenantID, resourceType string, resourceID uuid.UUID) ([]domain.RequestAttachment, error)
	Create(ctx context.Context, cmd CreateAttachmentCommand) (domain.RequestAttachment, error)
	Get(ctx context.Context, tenant kernel.TenantID, id uuid.UUID) (domain.RequestAttachment, error)
}

type AttachmentRepository interface {
	Save(ctx context.Context, att domain.RequestAttachment) error
	Get(ctx context.Context, tenant kernel.TenantID, id uuid.UUID) (domain.RequestAttachment, error)
	ListByResource(ctx context.Context, tenant kernel.TenantID, resourceType string, resourceID uuid.UUID) ([]domain.RequestAttachment, error)
}
