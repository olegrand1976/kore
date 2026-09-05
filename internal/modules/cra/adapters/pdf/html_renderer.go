package pdf

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"html/template"
	"time"

	"github.com/kore/kore/internal/modules/cra/domain"
)

//go:embed templates/cra.html
var craTemplateHTML string

type CRABrandData struct {
	CompanyName string
	// CompanyLogo is typed template.URL because it carries a data: URI, which
	// html/template's URL filter would otherwise replace with #ZgotmplZ.
	CompanyLogo       template.URL
	CompanyAddr       string
	UserID            string
	Collaborateur     string
	Month             string
	Status            string
	GeneratedAt       string
	ShowKoreFooter    bool
	ProductSite       string
	Client            string
	Mission           string
	Description       string
	Technologies      string
	Lieu              string
	ResponsableClient string
	Lines             []CRALine
}

// CRALine mirrors one row of the CRA grid: who, what it is charged to, the free
// comment and the time in hours and minutes.
type CRALine struct {
	Day        string
	Person     string
	Assignment string
	Comment    string
	Duration   string
}

type HTMLRenderer struct{}

func NewHTMLRenderer() *HTMLRenderer {
	return &HTMLRenderer{}
}

func (HTMLRenderer) Render(_ context.Context, ts domain.Timesheet, brand CRABrandData) ([]byte, error) {
	if brand.GeneratedAt == "" {
		brand.GeneratedAt = time.Now().Format("02/01/2006 15:04")
	}
	brand.UserID = ts.UserID.String()
	brand.Month = string(ts.Month)
	brand.Status = string(ts.Status)
	if len(brand.Lines) == 0 {
		brand.Lines = []CRALine{
			{Day: "—", Person: brand.Collaborateur, Assignment: "Prestation", Duration: "—"},
		}
	}

	tpl, err := template.New("cra").Parse(craTemplateHTML)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err := tpl.Execute(&buf, brand); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (HTMLRenderer) Filename(ts domain.Timesheet) string {
	return fmt.Sprintf("cra-%s-%s.html", ts.UserID, ts.Month)
}
