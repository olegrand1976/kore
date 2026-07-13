package app

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewTMAService(t *testing.T) {
	svc := NewService(nil, nil, nil, nil)
	require.NotNil(t, svc)
}
