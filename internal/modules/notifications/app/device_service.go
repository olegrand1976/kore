package app

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/notifications/domain"
	"github.com/kore/kore/internal/modules/notifications/ports"
)

type DeviceService struct {
	repo  ports.DeviceRepository
	clock ports.Clock
}

func NewDeviceService(repo ports.DeviceRepository, opts ...DeviceOption) *DeviceService {
	s := &DeviceService{
		repo:  repo,
		clock: realClock{},
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

type DeviceOption func(*DeviceService)

func WithDeviceClock(clock ports.Clock) DeviceOption {
	return func(s *DeviceService) {
		if clock != nil {
			s.clock = clock
		}
	}
}

func (s *DeviceService) RegisterDevice(ctx context.Context, cmd ports.RegisterDeviceCommand) error {
	platform, err := domain.ParseDevicePlatform(cmd.Platform)
	if err != nil {
		return err
	}
	token := strings.TrimSpace(cmd.Token)
	if token == "" {
		return domain.ErrEmptyDeviceToken
	}
	now := s.clock.Now()
	return s.repo.UpsertDeviceToken(ctx, domain.DeviceToken{
		ID:        uuid.New(),
		TenantID:  cmd.TenantID,
		UserID:    cmd.UserID,
		Platform:  platform,
		Token:     token,
		CreatedAt: now,
		UpdatedAt: now,
	})
}

func (s *DeviceService) UnregisterDevice(ctx context.Context, cmd ports.UnregisterDeviceCommand) error {
	token := strings.TrimSpace(cmd.Token)
	if token == "" {
		return domain.ErrEmptyDeviceToken
	}
	return s.repo.DeleteDeviceToken(ctx, cmd.TenantID, cmd.UserID, token)
}

var _ ports.DeviceService = (*DeviceService)(nil)
