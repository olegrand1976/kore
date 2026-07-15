package domain

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/kore/kore/pkg/kernel"
)

var ErrInvalidDevicePlatform = errors.New("invalid device platform")
var ErrEmptyDeviceToken = errors.New("device token required")

type DevicePlatform string

const (
	DevicePlatformIOS     DevicePlatform = "ios"
	DevicePlatformAndroid DevicePlatform = "android"
	DevicePlatformWeb     DevicePlatform = "web"
)

func ParseDevicePlatform(raw string) (DevicePlatform, error) {
	p := DevicePlatform(strings.ToLower(strings.TrimSpace(raw)))
	switch p {
	case DevicePlatformIOS, DevicePlatformAndroid, DevicePlatformWeb:
		return p, nil
	default:
		return "", fmt.Errorf("%w: %s", ErrInvalidDevicePlatform, raw)
	}
}

type DeviceToken struct {
	ID        uuid.UUID
	TenantID  kernel.TenantID
	UserID    uuid.UUID
	Platform  DevicePlatform
	Token     string
	CreatedAt time.Time
	UpdatedAt time.Time
}
