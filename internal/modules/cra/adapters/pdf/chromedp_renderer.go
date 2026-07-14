package pdf

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"strings"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/kore/kore/internal/modules/cra/domain"
	"github.com/kore/kore/internal/modules/cra/ports"
)

// ChromedpRenderer wraps an HTML renderer and converts output to PDF via headless Chrome.
type ChromedpRenderer struct {
	inner ports.PDFRenderer
}

func NewChromedpRenderer(inner ports.PDFRenderer) ports.PDFRenderer {
	return &ChromedpRenderer{inner: inner}
}

func (r *ChromedpRenderer) Render(ctx context.Context, ts domain.Timesheet) (domain.Document, error) {
	doc, err := r.inner.Render(ctx, ts)
	if err != nil {
		return domain.Document{}, err
	}
	if doc.MimeType != "text/html" {
		return doc, nil
	}
	pdfBytes, err := htmlToPDF(ctx, string(doc.Content))
	if err != nil {
		return domain.Document{}, fmt.Errorf("pdf render failed: %w", err)
	}
	filename := strings.TrimSuffix(doc.Filename, ".html") + ".pdf"
	return domain.Document{
		Filename: filename,
		Content:  pdfBytes,
		MimeType: "application/pdf",
	}, nil
}

func htmlToPDF(ctx context.Context, html string) ([]byte, error) {
	userDataDir := strings.TrimSpace(os.Getenv("CHROME_USER_DATA_DIR"))
	if userDataDir == "" {
		userDataDir = "/tmp/chrome-user-data"
	}
	if err := os.MkdirAll(userDataDir, 0o700); err != nil {
		return nil, fmt.Errorf("chrome user data dir: %w", err)
	}

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.NoFirstRun,
		chromedp.NoDefaultBrowserCheck,
		chromedp.UserDataDir(userDataDir),
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-setuid-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("disable-software-rasterizer", true),
		chromedp.Flag("disable-extensions", true),
	)
	if chromePath := strings.TrimSpace(os.Getenv("CHROME_PATH")); chromePath != "" {
		opts = append(opts, chromedp.ExecPath(chromePath))
	}
	allocCtx, cancelAlloc := chromedp.NewExecAllocator(ctx, opts...)
	defer cancelAlloc()

	taskCtx, cancelTask := chromedp.NewContext(allocCtx)
	defer cancelTask()

	dataURL := "data:text/html;base64," + base64.StdEncoding.EncodeToString([]byte(html))
	var pdfBuf []byte
	err := chromedp.Run(taskCtx,
		chromedp.Navigate(dataURL),
		chromedp.ActionFunc(func(ctx context.Context) error {
			var err error
			pdfBuf, _, err = page.PrintToPDF().WithPrintBackground(true).Do(ctx)
			return err
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("chromedp pdf: %w", err)
	}
	return pdfBuf, nil
}

var _ ports.PDFRenderer = (*ChromedpRenderer)(nil)
