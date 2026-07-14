package app

import (
	"bytes"
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/org/domain"
	"github.com/kore/kore/internal/modules/org/ports"
	"github.com/kore/kore/internal/platform/uploads"
	"github.com/kore/kore/pkg/kernel"
)

type attachmentService struct {
	repo     ports.AttachmentRepository
	checker  ports.AttachmentResourceChecker
}

func NewAttachmentService(repo ports.AttachmentRepository, checker ports.AttachmentResourceChecker) ports.AttachmentService {
	return &attachmentService{repo: repo, checker: checker}
}

func (s *attachmentService) ensureResource(ctx context.Context, tenant kernel.TenantID, resourceType string, resourceID uuid.UUID) error {
	if s.checker == nil {
		return nil
	}
	exists, err := s.checker.Exists(ctx, tenant, resourceType, resourceID)
	if err != nil {
		return err
	}
	if !exists {
		return domain.ErrAttachmentResourceNotFound
	}
	return nil
}

func (s *attachmentService) List(ctx context.Context, tenant kernel.TenantID, resourceType string, resourceID uuid.UUID) ([]domain.RequestAttachment, error) {
	if !domain.ValidResourceType(resourceType) {
		return nil, domain.ErrInvalidAttachmentTarget
	}
	if err := s.ensureResource(ctx, tenant, resourceType, resourceID); err != nil {
		return nil, err
	}
	return s.repo.ListByResource(ctx, tenant, resourceType, resourceID)
}

func (s *attachmentService) Create(ctx context.Context, cmd ports.CreateAttachmentCommand) (domain.RequestAttachment, error) {
	if !domain.ValidResourceType(cmd.ResourceType) {
		return domain.RequestAttachment{}, domain.ErrInvalidAttachmentTarget
	}
	if err := s.ensureResource(ctx, cmd.TenantID, cmd.ResourceType, cmd.ResourceID); err != nil {
		return domain.RequestAttachment{}, err
	}
	data, err := uploads.ReadAndValidateAttachment(cmd.Content, cmd.FileName)
	if err != nil {
		return domain.RequestAttachment{}, err
	}
	id := uuid.New()
	storagePath, err := uploads.StoreAttachment(cmd.UploadsDir, cmd.TenantID.UUID(), id, cmd.FileName, bytes.NewReader(data))
	if err != nil {
		return domain.RequestAttachment{}, err
	}
	mimeType := cmd.MimeType
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}
	att := domain.RequestAttachment{
		ID:           id,
		TenantID:     cmd.TenantID,
		ResourceType: cmd.ResourceType,
		ResourceID:   cmd.ResourceID,
		FileName:     cmd.FileName,
		MimeType:     mimeType,
		SizeBytes:    int64(len(data)),
		StoragePath:  storagePath,
		UploadedBy:   cmd.UploadedBy,
		CreatedAt:    time.Now().UTC(),
	}
	return att, s.repo.Save(ctx, att)
}

func (s *attachmentService) Get(ctx context.Context, tenant kernel.TenantID, id uuid.UUID) (domain.RequestAttachment, error) {
	return s.repo.Get(ctx, tenant, id)
}

var _ ports.AttachmentService = (*attachmentService)(nil)
