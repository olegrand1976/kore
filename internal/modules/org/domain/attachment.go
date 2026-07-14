package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/kore/kore/pkg/kernel"
)

var (
	ErrAttachmentNotFound      = errors.New("attachment not found")
	ErrInvalidAttachmentTarget = errors.New("invalid attachment resource type")
)

const (
	ResourceTypeTmaDemand              = "tma_demand"
	ResourceTypeSupportTicket          = "support_ticket"
	ResourceTypeMaintenanceWorkRequest = "maintenance_work_request"
)

func ValidResourceType(resourceType string) bool {
	switch resourceType {
	case ResourceTypeTmaDemand, ResourceTypeSupportTicket, ResourceTypeMaintenanceWorkRequest:
		return true
	default:
		return false
	}
}

func ResourceModule(resourceType string) string {
	switch resourceType {
	case ResourceTypeTmaDemand:
		return "tma"
	case ResourceTypeSupportTicket:
		return "support"
	case ResourceTypeMaintenanceWorkRequest:
		return "maintenance"
	default:
		return ""
	}
}

type RequestAttachment struct {
	ID           uuid.UUID       `json:"id"`
	TenantID     kernel.TenantID `json:"tenantId"`
	ResourceType string          `json:"resourceType"`
	ResourceID   uuid.UUID       `json:"resourceId"`
	FileName     string          `json:"fileName"`
	MimeType     string          `json:"mimeType"`
	SizeBytes    int64           `json:"sizeBytes"`
	StoragePath  string          `json:"-"`
	UploadedBy   uuid.UUID       `json:"uploadedBy"`
	CreatedAt    time.Time       `json:"createdAt"`
}
