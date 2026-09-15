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
	Delete(ctx context.Context, tenant kernel.TenantID, id uuid.UUID) error
}

// AttachmentLifecycleHook is notified after attachment create/delete (e.g. RAG index).
type AttachmentLifecycleHook interface {
	OnCreated(ctx context.Context, att domain.RequestAttachment) error
	OnDeleted(ctx context.Context, tenant kernel.TenantID, attachmentID uuid.UUID) error
}

type AttachmentResourceChecker interface {
	Exists(ctx context.Context, tenant kernel.TenantID, resourceType string, resourceID uuid.UUID) (bool, error)
}

type AttachmentRepository interface {
	Save(ctx context.Context, att domain.RequestAttachment) error
	Get(ctx context.Context, tenant kernel.TenantID, id uuid.UUID) (domain.RequestAttachment, error)
	ListByResource(ctx context.Context, tenant kernel.TenantID, resourceType string, resourceID uuid.UUID) ([]domain.RequestAttachment, error)
	Delete(ctx context.Context, tenant kernel.TenantID, id uuid.UUID) error
}
