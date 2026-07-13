package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/kore/kore/pkg/kernel"
)

var (
	ErrParameterSetNotFound        = errors.New("parameter set not found")
	ErrTemplateNotFound            = errors.New("template not found")
	ErrPhoneDirectoryEntryNotFound = errors.New("phone directory entry not found")
)

type ParameterSet struct {
	ID        uuid.UUID
	TenantID  kernel.TenantID
	Code      string
	Payload   map[string]any
	UpdatedAt time.Time
}

type Template struct {
	ID        uuid.UUID
	TenantID  kernel.TenantID
	Type      string
	Name      string
	Content   map[string]any
	Active    bool
	CreatedAt time.Time
}

type PhoneDirectoryEntry struct {
	ID         uuid.UUID
	TenantID   kernel.TenantID
	UserID     *uuid.UUID
	Label      string
	Phone      string
	Visibility string
	CreatedAt  time.Time
}

func NewParameterSet(tenant kernel.TenantID, code string, payload map[string]any) ParameterSet {
	if payload == nil {
		payload = map[string]any{}
	}
	return ParameterSet{
		ID:        uuid.New(),
		TenantID:  tenant,
		Code:      code,
		Payload:   payload,
		UpdatedAt: time.Now().UTC(),
	}
}

func NewTemplate(tenant kernel.TenantID, tmplType, name string, content map[string]any) Template {
	if content == nil {
		content = map[string]any{}
	}
	return Template{
		ID:        uuid.New(),
		TenantID:  tenant,
		Type:      tmplType,
		Name:      name,
		Content:   content,
		Active:    true,
		CreatedAt: time.Now().UTC(),
	}
}

func NewPhoneDirectoryEntry(tenant kernel.TenantID, label, phone string) PhoneDirectoryEntry {
	return PhoneDirectoryEntry{
		ID:         uuid.New(),
		TenantID:   tenant,
		Label:      label,
		Phone:      phone,
		Visibility: "internal",
		CreatedAt:  time.Now().UTC(),
	}
}
