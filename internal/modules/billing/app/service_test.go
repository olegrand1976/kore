package app

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewServiceUsesMockWhenNoStripeKey(t *testing.T) {
	svc := NewService(nil, "", "", 60)
	require.NotNil(t, svc)
}

func TestNewServiceWithGateway(t *testing.T) {
	svc := NewServiceWithGateway(nil, nil, 60)
	require.NotNil(t, svc)
}
