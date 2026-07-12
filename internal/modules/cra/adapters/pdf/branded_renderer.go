package pdf

import (
	"context"

	"github.com/kore/kore/internal/modules/cra/domain"
	"github.com/kore/kore/internal/modules/cra/ports"
)

// BrandedRenderer renders charted HTML documents for CRA exports.
type BrandedRenderer struct {
	html  *HTMLRenderer
	brand CRABrandData
}

func NewBrandedRenderer(brand CRABrandData) ports.PDFRenderer {
	if brand.ShowKoreFooter == false && brand.CompanyName == "" {
		brand.ShowKoreFooter = true
	}
	return &BrandedRenderer{
		html:  NewHTMLRenderer(),
		brand: brand,
	}
}

func (r *BrandedRenderer) Render(ctx context.Context, ts domain.Timesheet) (domain.Document, error) {
	content, err := r.html.Render(ctx, ts, r.brand)
	if err != nil {
		return domain.Document{}, err
	}
	return domain.Document{
		Filename: r.html.Filename(ts),
		Content:  content,
		MimeType: "text/html",
	}, nil
}

var _ ports.PDFRenderer = (*BrandedRenderer)(nil)
