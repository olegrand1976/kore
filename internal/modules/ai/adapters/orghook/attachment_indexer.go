package orghook

import (
	"context"

	"github.com/google/uuid"
	aiports "github.com/kore/kore/internal/modules/ai/ports"
	orgdomain "github.com/kore/kore/internal/modules/org/domain"
	orgports "github.com/kore/kore/internal/modules/org/ports"
	"github.com/kore/kore/pkg/kernel"
)

// AttachmentIndexer indexes TMA demand attachments for RAG after upload/delete.
type AttachmentIndexer struct {
	ai aiports.AIService
}

func NewAttachmentIndexer(ai aiports.AIService) *AttachmentIndexer {
	return &AttachmentIndexer{ai: ai}
}

func (h *AttachmentIndexer) OnCreated(ctx context.Context, att orgdomain.RequestAttachment) error {
	if h == nil || h.ai == nil {
		return nil
	}
	if att.ResourceType != orgdomain.ResourceTypeTmaDemand {
		return nil
	}
	return h.ai.IndexRequestAttachment(ctx, aiports.IndexAttachmentCommand{
		TenantID:     att.TenantID,
		AttachmentID: att.ID,
		DemandID:     att.ResourceID,
		FileName:     att.FileName,
		MimeType:     att.MimeType,
		StoragePath:  att.StoragePath,
	})
}

func (h *AttachmentIndexer) OnDeleted(ctx context.Context, tenant kernel.TenantID, attachmentID uuid.UUID) error {
	if h == nil || h.ai == nil {
		return nil
	}
	return h.ai.RemoveRequestAttachmentChunks(ctx, tenant, attachmentID)
}

var _ orgports.AttachmentLifecycleHook = (*AttachmentIndexer)(nil)
