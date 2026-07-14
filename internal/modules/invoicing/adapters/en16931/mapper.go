package en16931

import (
	"time"

	"github.com/kore/kore/internal/modules/invoicing/domain"
)

// Document is a simplified EN 16931 / UBL 2.1 JSON representation.
type Document map[string]any

func MapInvoice(inv domain.Invoice) Document {
	lines := make([]map[string]any, 0, len(inv.Lines))
	for i, line := range inv.Lines {
		lineTotal := int64(float64(line.UnitPrice) * line.Quantity)
		lines = append(lines, map[string]any{
			"id":                  line.ID.String(),
			"lineNumber":          i + 1,
			"description":         line.Description,
			"quantity":            line.Quantity,
			"unitPrice":           line.UnitPrice,
			"lineExtensionAmount": lineTotal,
			"taxRate":             line.TaxRate,
		})
	}
	doc := Document{
		"specificationIdentifier": "urn:cen.eu:en16931:2017",
		"invoiceNumber":           inv.ID.String(),
		"issueDate":               inv.CreatedAt.UTC().Format(time.RFC3339),
		"invoiceTypeCode":         mapInvoiceType(inv.Type),
		"documentCurrencyCode":    inv.Currency,
		"buyerReference":          inv.ClientID.String(),
		"legalMonetaryTotal": map[string]any{
			"taxExclusiveAmount": inv.TotalAmount,
			"taxInclusiveAmount": inv.TotalAmount + inv.TaxAmount,
			"payableAmount":      inv.TotalAmount + inv.TaxAmount,
		},
		"taxTotal": map[string]any{
			"taxAmount": inv.TaxAmount,
		},
		"invoiceLines": lines,
	}
	if inv.PDPReceiptID != "" {
		doc["pdpReceiptId"] = inv.PDPReceiptID
	}
	return doc
}

func mapInvoiceType(t domain.InvoiceType) string {
	switch t {
	case domain.InvoiceTypeStandard:
		return "380"
	case domain.InvoiceTypeCreditNote:
		return "381"
	default:
		return "380"
	}
}
