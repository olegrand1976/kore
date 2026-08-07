package org

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/invoicing/ports"
	orgports "github.com/kore/kore/internal/modules/org/ports"
	"github.com/kore/kore/pkg/kernel"
)

type ClientContactReader struct {
	clients orgports.ClientService
}

func NewClientContactReader(clients orgports.ClientService) *ClientContactReader {
	return &ClientContactReader{clients: clients}
}

func (r *ClientContactReader) PrimaryBillingContact(ctx context.Context, tenant kernel.TenantID, clientID uuid.UUID) (string, string, error) {
	if r == nil || r.clients == nil {
		return "", "", nil
	}
	client, err := r.clients.GetClient(ctx, tenant, clientID)
	if err != nil {
		return "", "", err
	}
	for _, c := range client.Contacts {
		email := strings.TrimSpace(strings.ToLower(c.Email))
		if email != "" && strings.Contains(email, "@") {
			return email, client.RaisonSociale, nil
		}
	}
	return "", client.RaisonSociale, nil
}

var _ ports.ClientContactReader = (*ClientContactReader)(nil)
