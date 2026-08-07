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
	enabledReader kernel.InvoicingEnabledReader
	clientReader  ports.ClientContactReader
	mailer        ports.MailSender
	workflow      ports.WorkflowService
	clock         func() time.Time
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

func WithEnabledReader(reader kernel.InvoicingEnabledReader) Option {
	return func(s *service) {
		s.enabledReader = reader
	}
}

func WithClientContactReader(reader ports.ClientContactReader) Option {
	return func(s *service) {
		s.clientReader = reader
	}
}

func WithMailSender(mailer ports.MailSender) Option {
	return func(s *service) {
		s.mailer = mailer
	}
}

func WithWorkflow(wf ports.WorkflowService) Option {
	return func(s *service) {
		s.workflow = wf
	}
}

func WithClock(clock func() time.Time) Option {
	return func(s *service) {
		s.clock = clock
	}
}

func NewService(repo ports.InvoicingRepository, opts ...Option) ports.InvoicingService {
	s := &service{repo: repo, clock: time.Now}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

func (s *service) now() time.Time {
	if s.clock == nil {
		return time.Now().UTC()
	}
	return s.clock().UTC()
}

func (s *service) requireEnabled(ctx context.Context, tenant kernel.TenantID) error {
	if s.enabledReader == nil {
		return nil
	}
	ok, err := s.enabledReader.IsInvoicingEnabled(ctx, tenant)
	if err != nil {
		return err
	}
	if !ok {
		return domain.ErrInvoicingDisabled
	}
	return nil
}

func (s *service) List(ctx context.Context, tenant kernel.TenantID) ([]domain.Invoice, error) {
	if err := s.requireEnabled(ctx, tenant); err != nil {
		return nil, err
	}
	return s.repo.ListInvoices(ctx, tenant)
}

func (s *service) Get(ctx context.Context, tenant kernel.TenantID, id uuid.UUID) (domain.Invoice, error) {
	if err := s.requireEnabled(ctx, tenant); err != nil {
		return domain.Invoice{}, err
	}
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
	if err := s.requireEnabled(ctx, cmd.TenantID); err != nil {
		return domain.Invoice{}, err
	}
	inv, lines, err := s.buildPrepareeInvoice(cmd)
	if err != nil {
		return domain.Invoice{}, err
	}
	if err := s.repo.SaveInvoiceWithLines(ctx, inv, lines); err != nil {
		return domain.Invoice{}, err
	}
	return inv, nil
}

func (s *service) buildPrepareeInvoice(cmd ports.CreateInvoiceCommand) (domain.Invoice, []domain.InvoiceLine, error) {
	if len(cmd.Lines) == 0 {
		return domain.Invoice{}, nil, domain.ErrInvalidInvoiceLine
	}
	inv := domain.NewInvoice(cmd.TenantID, cmd.ClientID, cmd.Type, cmd.Currency)
	inv.SourceTimesheetID = cmd.SourceTimesheetID
	var total, tax int64
	lines := make([]domain.InvoiceLine, 0, len(cmd.Lines))
	for _, lineIn := range cmd.Lines {
		lineTotal, err := domain.LineNetCents(lineIn.UnitPrice, lineIn.Quantity)
		if err != nil {
			return domain.Invoice{}, nil, err
		}
		lineTax, err := domain.LineTaxCents(lineTotal, lineIn.TaxRate)
		if err != nil {
			return domain.Invoice{}, nil, err
		}
		line := domain.InvoiceLine{
			ID:          uuid.New(),
			TenantID:    cmd.TenantID,
			InvoiceID:   inv.ID,
			Description: lineIn.Description,
			Quantity:    lineIn.Quantity,
			UnitPrice:   lineIn.UnitPrice,
			TaxRate:     lineIn.TaxRate,
		}
		total += lineTotal
		tax += lineTax
		lines = append(lines, line)
	}
	inv.TotalAmount = total
	inv.TaxAmount = tax
	inv.Status = domain.InvoiceStatusPreparee
	inv.Lines = lines
	return inv, lines, nil
}

func (s *service) ComputeVirtual(ctx context.Context, cmd ports.ComputeVirtualCommand) (domain.Invoice, error) {
	if err := s.requireEnabled(ctx, cmd.TenantID); err != nil {
		return domain.Invoice{}, err
	}
	clientID := cmd.ClientID
	currency := "EUR"
	linesIn := cmd.Lines
	if cmd.MissionID != nil && s.missionReader != nil && len(linesIn) == 0 {
		billing, err := s.missionReader.ActiveMissionDays(ctx, cmd.TenantID, *cmd.MissionID, cmd.Period)
		if err != nil {
			return domain.Invoice{}, err
		}
		clientID = billing.ClientID
		currency = billing.Currency
		label := "Mission SSII"
		if billing.Title != "" {
			label = billing.Title
		}
		var desc string
		switch billing.RateUnit {
		case "hourly":
			desc = fmt.Sprintf("%s — %.2f h × taux horaire", label, billing.Quantity)
		default:
			desc = fmt.Sprintf("%s — %.0f j × TJM", label, billing.Quantity)
		}
		linesIn = []ports.InvoiceLineInput{{
			Description: desc,
			Quantity:    billing.Quantity,
			UnitPrice:   billing.UnitPrice,
			TaxRate:     20,
		}}
	}
	inv := domain.NewInvoice(cmd.TenantID, clientID, domain.InvoiceTypeStandard, currency)
	if len(linesIn) == 0 {
		return domain.Invoice{}, domain.ErrNoBillableContent
	}
	var total, tax int64
	lines := make([]domain.InvoiceLine, 0, len(linesIn))
	for _, lineIn := range linesIn {
		lineTotal, err := domain.LineNetCents(lineIn.UnitPrice, lineIn.Quantity)
		if err != nil {
			return domain.Invoice{}, err
		}
		lineTax, err := domain.LineTaxCents(lineTotal, lineIn.TaxRate)
		if err != nil {
			return domain.Invoice{}, err
		}
		line := domain.InvoiceLine{
			ID:          uuid.New(),
			TenantID:    cmd.TenantID,
			InvoiceID:   inv.ID,
			Description: lineIn.Description,
			Quantity:    lineIn.Quantity,
			UnitPrice:   lineIn.UnitPrice,
			TaxRate:     lineIn.TaxRate,
		}
		total += lineTotal
		tax += lineTax
		lines = append(lines, line)
	}
	inv.TotalAmount = total
	inv.TaxAmount = tax
	inv.Lines = lines
	if err := s.repo.SaveInvoiceWithLines(ctx, inv, lines); err != nil {
		return domain.Invoice{}, err
	}
	return inv, nil
}

func (s *service) Transmit(ctx context.Context, tenant kernel.TenantID, id uuid.UUID) (domain.Invoice, error) {
	if err := s.requireEnabled(ctx, tenant); err != nil {
		return domain.Invoice{}, err
	}
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
	if err := s.requireEnabled(ctx, tenant); err != nil {
		return domain.Invoice{}, err
	}
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
