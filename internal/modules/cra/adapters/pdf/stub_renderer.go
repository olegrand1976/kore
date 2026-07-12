package pdf

import (
	"context"
	"fmt"

	"github.com/kore/kore/internal/modules/cra/domain"
	"github.com/kore/kore/internal/modules/cra/ports"
)

// StubRenderer is a placeholder PDF gateway until a real renderer is wired.
type StubRenderer struct{}

func NewStubRenderer() ports.PDFRenderer {
	return StubRenderer{}
}

func (StubRenderer) Render(_ context.Context, ts domain.Timesheet) (domain.Document, error) {
	body := fmt.Sprintf("CRA stub PDF — user %s — month %s — status %s",
		ts.UserID, ts.Month, ts.Status)
	return domain.Document{
		Filename: fmt.Sprintf("cra-%s-%s.pdf", ts.UserID, ts.Month),
		Content:  []byte(body),
		MimeType: "application/pdf",
	}, nil
}
