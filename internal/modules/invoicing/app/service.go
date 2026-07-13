package app

import (
	"context"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/invoicing/domain"
	"github.com/kore/kore/internal/modules/invoicing/ports"
	"github.com/kore/kore/pkg/kernel"
)

type service struct {
	repo ports.InvoicingRepository
}

func NewService(repo ports.InvoicingRepository) ports.InvoicingService {
	return &service{repo: repo}
}

func (s *service) List(ctx context.Context, tenant kernel.TenantID) ([]domain.Invoice, error) {
	return s.repo.ListInvoices(ctx, tenant)
}

func (s *service) Get(ctx context.Context, tenant kernel.TenantID, id uuid.UUID) (domain.Invoice, error) {
	inv, err := s.repo.GetInvoice(ctx, tenant, id)
	if err != nil {
		return domain.Invoice{}, err
	}
	lines, err := s.repo.ListInvoiceLines(ctx, tenant, id)
	if err != nil {
		return domain.Invoice{}, err
	}
	inv.Lines = lines
	return inv, nil
}

func (s *service) Create(ctx context.Context, cmd ports.CreateInvoiceCommand) (domain.Invoice, error) {
	inv := domain.NewInvoice(cmd.TenantID, cmd.ClientID, cmd.Type, cmd.Currency)
	var total, tax int64
	for _, lineIn := range cmd.Lines {
		line := domain.InvoiceLine{
			ID:          uuid.New(),
			TenantID:    cmd.TenantID,
			InvoiceID:   inv.ID,
			Description: lineIn.Description,
			Quantity:    lineIn.Quantity,
			UnitPrice:   lineIn.UnitPrice,
			TaxRate:     lineIn.TaxRate,
		}
		lineTotal := int64(float64(line.UnitPrice) * line.Quantity)
		total += lineTotal
		tax += int64(float64(lineTotal) * line.TaxRate / 100)
		if err := s.repo.SaveInvoiceLine(ctx, line); err != nil {
			return domain.Invoice{}, err
		}
		inv.Lines = append(inv.Lines, line)
	}
	inv.TotalAmount = total
	inv.TaxAmount = tax
	inv.Status = domain.InvoiceStatusPreparee
	return inv, s.repo.SaveInvoice(ctx, inv)
}

func (s *service) ComputeVirtual(ctx context.Context, cmd ports.ComputeVirtualCommand) (domain.Invoice, error) {
	inv := domain.NewInvoice(cmd.TenantID, cmd.ClientID, domain.InvoiceTypeStandard, "EUR")
	return inv, s.repo.SaveInvoice(ctx, inv)
}

func (s *service) Transmit(ctx context.Context, tenant kernel.TenantID, id uuid.UUID) (domain.Invoice, error) {
	inv, err := s.repo.GetInvoice(ctx, tenant, id)
	if err != nil {
		return domain.Invoice{}, err
	}
	if err := inv.Transmit(); err != nil {
		return domain.Invoice{}, err
	}
	if err := s.repo.SaveInvoice(ctx, inv); err != nil {
		return domain.Invoice{}, err
	}
	item := domain.PDPQueueItem{
		ID:        uuid.New(),
		TenantID:  tenant,
		InvoiceID: id,
		Payload:   map[string]any{"invoiceId": id.String()},
		Status:    "pending",
		CreatedAt: inv.TransmittedAt.UTC(),
	}
	return inv, s.repo.SavePDPQueueItem(ctx, item)
}

func (s *service) CreateCreditNote(ctx context.Context, tenant kernel.TenantID, invoiceID uuid.UUID) (domain.Invoice, error) {
	orig, err := s.repo.GetInvoice(ctx, tenant, invoiceID)
	if err != nil {
		return domain.Invoice{}, err
	}
	cn := domain.NewInvoice(tenant, orig.ClientID, domain.InvoiceTypeCreditNote, orig.Currency)
	cn.TotalAmount = -orig.TotalAmount
	cn.TaxAmount = -orig.TaxAmount
	cn.Status = domain.InvoiceStatusPreparee
	return cn, s.repo.SaveInvoice(ctx, cn)
}

var _ ports.InvoicingService = (*service)(nil)
