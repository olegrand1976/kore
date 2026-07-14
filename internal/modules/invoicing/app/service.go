package app

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/invoicing/adapters/en16931"
	"github.com/kore/kore/internal/modules/invoicing/domain"
	"github.com/kore/kore/internal/modules/invoicing/ports"
	ssiiports "github.com/kore/kore/internal/modules/ssii/ports"
	"github.com/kore/kore/pkg/kernel"
)

type service struct {
	repo          ports.InvoicingRepository
	pdp           ports.PDPGateway
	missionReader ssiiports.MissionReader
}

type Option func(*service)

func WithPDPGateway(gw ports.PDPGateway) Option {
	return func(s *service) {
		s.pdp = gw
	}
}

func WithMissionReader(reader ssiiports.MissionReader) Option {
	return func(s *service) {
		s.missionReader = reader
	}
}

func NewService(repo ports.InvoicingRepository, opts ...Option) ports.InvoicingService {
	s := &service{repo: repo}
	for _, opt := range opts {
		opt(s)
	}
	return s
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
	clientID := cmd.ClientID
	currency := "EUR"
	lines := cmd.Lines
	if cmd.MissionID != nil && s.missionReader != nil && len(lines) == 0 {
		billing, err := s.missionReader.ActiveMissionDays(ctx, cmd.TenantID, *cmd.MissionID, cmd.Period)
		if err != nil {
			return domain.Invoice{}, err
		}
		clientID = billing.ClientID
		currency = billing.Currency
		lines = []ports.InvoiceLineInput{{
			Description: fmt.Sprintf("Mission SSII — %.0f j × TJM", billing.Days),
			Quantity:    billing.Days,
			UnitPrice:   billing.TJMAmount,
			TaxRate:     20,
		}}
	}
	inv := domain.NewInvoice(cmd.TenantID, clientID, domain.InvoiceTypeStandard, currency)
	if len(lines) == 0 {
		lines = []ports.InvoiceLineInput{{
			Description: fmt.Sprintf("Prestation virtuelle %s — %s", cmd.Period.Start.Format("2006-01-02"), cmd.Period.End.Format("2006-01-02")),
			Quantity:    1,
			UnitPrice:   0,
			TaxRate:     20,
		}}
	}
	var total, tax int64
	for _, lineIn := range lines {
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
	return inv, s.repo.SaveInvoice(ctx, inv)
}

func (s *service) Transmit(ctx context.Context, tenant kernel.TenantID, id uuid.UUID) (domain.Invoice, error) {
	inv, err := s.Get(ctx, tenant, id)
	if err != nil {
		return domain.Invoice{}, err
	}
	if err := inv.Transmit(); err != nil {
		return domain.Invoice{}, err
	}
	doc := ports.En16931Document(en16931.MapInvoice(inv))
	if s.pdp != nil {
		receipt, err := s.pdp.Transmit(ctx, tenant, doc)
		if err != nil {
			item := domain.PDPQueueItem{
				ID:        uuid.New(),
				TenantID:  tenant,
				InvoiceID: id,
				Payload:   map[string]any{"invoiceId": id.String(), "document": doc},
				Status:    "pending",
				LastError: err.Error(),
				CreatedAt: time.Now().UTC(),
			}
			_ = s.repo.SavePDPQueueItem(ctx, item)
			return domain.Invoice{}, fmt.Errorf("pdp unavailable: %w", err)
		}
		inv.PDPReceiptID = receipt.ID
	}
	if err := s.repo.SaveInvoice(ctx, inv); err != nil {
		return domain.Invoice{}, err
	}
	item := domain.PDPQueueItem{
		ID:        uuid.New(),
		TenantID:  tenant,
		InvoiceID: id,
		Payload:   map[string]any{"invoiceId": id.String(), "receiptId": inv.PDPReceiptID},
		Status:    "sent",
		CreatedAt: inv.TransmittedAt.UTC(),
	}
	return inv, s.repo.SavePDPQueueItem(ctx, item)
}

func (s *service) SyncPDPStatus(ctx context.Context, evt ports.PDPStatusEvent) error {
	inv, err := s.repo.GetInvoice(ctx, evt.TenantID, evt.InvoiceID)
	if err != nil {
		return err
	}
	status := evt.Status
	if status == "" && s.pdp != nil && evt.ReceiptID != "" {
		status, err = s.pdp.SyncStatus(ctx, evt.ReceiptID)
		if err != nil {
			return err
		}
	}
	if status == "" {
		return domain.ErrInvalidInvoiceState
	}
	inv.Status = status
	if evt.ReceiptID != "" {
		inv.PDPReceiptID = evt.ReceiptID
	}
	return s.repo.SaveInvoice(ctx, inv)
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
