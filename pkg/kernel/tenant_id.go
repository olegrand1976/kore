package kernel

import (
	"encoding/json"
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

// MarshalJSON encodes TenantID as a UUID string (not an empty object).
func (t TenantID) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.value.String())
}

// UnmarshalJSON accepts a UUID string.
func (t *TenantID) UnmarshalJSON(data []byte) error {
	var raw string
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidTenantID, err)
	}
	parsed, err := ParseTenantID(raw)
	if err != nil {
		return err
	}
	*t = parsed
	return nil
}
