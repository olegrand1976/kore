package kernel

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
)

var ErrInvalidTenantID = errors.New("invalid tenant id")

type TenantID struct {
	value uuid.UUID
}

func NewTenantID(id uuid.UUID) TenantID {
	return TenantID{value: id}
}

func ParseTenantID(raw string) (TenantID, error) {
	id, err := uuid.Parse(raw)
	if err != nil {
		return TenantID{}, fmt.Errorf("%w: %s", ErrInvalidTenantID, raw)
	}
	return TenantID{value: id}, nil
}

func (t TenantID) UUID() uuid.UUID { return t.value }
func (t TenantID) String() string  { return t.value.String() }
func (t TenantID) IsZero() bool    { return t.value == uuid.Nil }
