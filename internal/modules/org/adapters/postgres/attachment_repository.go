package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/kore/kore/internal/modules/org/domain"
	"github.com/kore/kore/internal/modules/org/ports"
	"github.com/kore/kore/internal/platform/db"
	"github.com/kore/kore/pkg/kernel"
)

type AttachmentRepository struct {
	pool *db.Pool
}

func NewAttachmentRepository(pool *db.Pool) *AttachmentRepository {
	return &AttachmentRepository{pool: pool}
}

func (r *AttachmentRepository) Save(ctx context.Context, att domain.RequestAttachment) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO org.request_attachments (
			id, tenant_id, resource_type, resource_id, file_name, mime_type, size_bytes, storage_path, uploaded_by, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`, att.ID, att.TenantID.UUID(), att.ResourceType, att.ResourceID, att.FileName, att.MimeType,
		att.SizeBytes, att.StoragePath, att.UploadedBy, att.CreatedAt)
	return err
}

func (r *AttachmentRepository) Get(ctx context.Context, tenant kernel.TenantID, id uuid.UUID) (domain.RequestAttachment, error) {
	return r.scanAttachment(r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, resource_type, resource_id, file_name, mime_type, size_bytes, storage_path, uploaded_by, created_at
		FROM org.request_attachments WHERE tenant_id = $1 AND id = $2
	`, tenant.UUID(), id))
}

func (r *AttachmentRepository) ListByResource(ctx context.Context, tenant kernel.TenantID, resourceType string, resourceID uuid.UUID) ([]domain.RequestAttachment, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, resource_type, resource_id, file_name, mime_type, size_bytes, storage_path, uploaded_by, created_at
		FROM org.request_attachments
		WHERE tenant_id = $1 AND resource_type = $2 AND resource_id = $3
		ORDER BY created_at
	`, tenant.UUID(), resourceType, resourceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.RequestAttachment
	for rows.Next() {
		att, err := r.scanAttachment(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, att)
	}
	return out, rows.Err()
}

func (r *AttachmentRepository) scanAttachment(row pgx.Row) (domain.RequestAttachment, error) {
	var att domain.RequestAttachment
	var tenantID uuid.UUID
	err := row.Scan(
		&att.ID, &tenantID, &att.ResourceType, &att.ResourceID, &att.FileName, &att.MimeType,
		&att.SizeBytes, &att.StoragePath, &att.UploadedBy, &att.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.RequestAttachment{}, domain.ErrAttachmentNotFound
		}
		return domain.RequestAttachment{}, err
	}
	att.TenantID = kernel.NewTenantID(tenantID)
	return att, nil
}

var _ ports.AttachmentRepository = (*AttachmentRepository)(nil)
