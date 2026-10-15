// internal/app/adapters/certificate/pdf_adapter.go

package certificate

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/johnfercher/maroto/v2"
	"github.com/johnfercher/maroto/v2/pkg/components/col"
	"github.com/johnfercher/maroto/v2/pkg/components/image"
	"github.com/johnfercher/maroto/v2/pkg/components/text"
	"github.com/johnfercher/maroto/v2/pkg/config"
	"github.com/johnfercher/maroto/v2/pkg/consts/align"
	"github.com/johnfercher/maroto/v2/pkg/consts/extension"
	"github.com/johnfercher/maroto/v2/pkg/consts/fontstyle"
	"github.com/johnfercher/maroto/v2/pkg/consts/orientation"
	"github.com/johnfercher/maroto/v2/pkg/props"
	"github.com/skip2/go-qrcode"

	certdomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/certificate/domain"
)

// PDFAdapter implements domain.PDFGenerator using maroto v2 for
// layout and go-qrcode for the verification QR.
type PDFAdapter struct {
	httpClient *http.Client
}

// NewPDFAdapter constructs a PDFAdapter with a default HTTP client for
// fetching remote images referenced by the template design.
func NewPDFAdapter() *PDFAdapter {
	return &PDFAdapter{
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// Render implements domain.PDFGenerator.
//
// Produces a landscape A4 certificate. Design values from the template
// override colors, fonts, images, and text. Text fields support
// {{key}} placeholder substitution from the certificate's metadata.
func (a *PDFAdapter) Render(
	ctx context.Context,
	in certdomain.RenderInput,
) ([]byte, error) {
	if in.Certificate == nil {
		return nil, fmt.Errorf("pdf: certificate is nil")
	}
	if in.Template == nil {
		return nil, fmt.Errorf("pdf: template is nil")
	}

	design := in.Template.Design
	cert := in.Certificate

	// ---- QR code ----
	qrBytes, err := qrcode.Encode(in.VerifyURL, qrcode.Medium, 256)
	if err != nil {
		return nil, fmt.Errorf("pdf: qr generation failed: %w", err)
	}

	// ---- Substituted text ----
	titleText := substitute(design.TitleText, cert)
	if titleText == "" {
		titleText = "Certificate of Completion"
	}
	subtitleText := substitute(design.SubtitleText, cert)
	if subtitleText == "" {
		subtitleText = "This certifies that"
	}
	bodyText := substitute(design.BodyText, cert)
	if bodyText == "" {
		bodyText = "has successfully completed {{context_title}} on {{context_date}}."
		bodyText = substitute(bodyText, cert)
	}
	footerText := substitute(design.FooterText, cert)
	if footerText == "" {
		footerText = in.VerifyURL
	}

	// ---- Colors ----
	textColor := hexToColor(design.TextColor, 0x1A, 0x1A, 0x1A)
	accentColor := hexToColor(design.AccentColor, 0x0A, 0x6E, 0xBD)

	// ---- Build document ----
	cfg := config.NewBuilder().
		WithOrientation(orientation.Horizontal).
		WithLeftMargin(20).
		WithRightMargin(20).
		WithTopMargin(15).
		WithBottomMargin(15).
		Build()

	m := maroto.New(cfg)

	// ---- Row 1: Logo (left) | Seal (right) ----
	logoCol := col.New(6)
	sealCol := col.New(6)

	if design.LogoURL != "" {
		if data, ext, err := a.fetchImage(ctx, design.LogoURL); err == nil {
			logoCol.Add(image.NewFromBytes(data, ext, props.Rect{
				Percent: 50,
				Center:  false,
			}))
		}
	}
	if design.SealURL != "" {
		if data, ext, err := a.fetchImage(ctx, design.SealURL); err == nil {
			sealCol.Add(image.NewFromBytes(data, ext, props.Rect{
				Percent: 50,
				Center:  true,
			}))
		}
	}
	if design.LogoURL != "" || design.SealURL != "" {
		m.AddRow(30, logoCol, sealCol)
	} else {
		m.AddRow(20)
	}

	// ---- Row 2: Title ----
	m.AddRow(20,
		col.New(12).Add(text.New(titleText, props.Text{
			Size:  28,
			Style: fontstyle.Bold,
			Align: align.Center,
			Color: textColor,
		})),
	)

	// ---- Row 3: Subtitle ----
	m.AddRow(10,
		col.New(12).Add(text.New(subtitleText, props.Text{
			Size:  12,
			Align: align.Center,
			Color: textColor,
		})),
	)

	// ---- Row 4: Recipient name ----
	m.AddRow(20,
		col.New(12).Add(text.New(cert.RecipientName, props.Text{
			Size:  32,
			Style: fontstyle.Bold,
			Align: align.Center,
			Color: accentColor,
		})),
	)

	// ---- Row 5: Body ----
	m.AddRow(20,
		col.New(12).Add(text.New(bodyText, props.Text{
			Size:  14,
			Align: align.Center,
			Color: textColor,
		})),
	)

	m.AddRow(5)

	// ---- Row 6: Signature (left) | QR (right) ----
	sigCol := col.New(6)
	if design.SignatureURL != "" {
		if data, ext, err := a.fetchImage(ctx, design.SignatureURL); err == nil {
			sigCol.Add(image.NewFromBytes(data, ext, props.Rect{
				Percent: 40,
				Center:  false,
			}))
		}
	}
	if design.SignatoryName != "" {
		sigCol.Add(text.New(design.SignatoryName, props.Text{
			Size:  11,
			Style: fontstyle.Bold,
			Align: align.Left,
			Color: textColor,
		}))
	}
	if design.SignatoryTitle != "" {
		sigCol.Add(text.New(design.SignatoryTitle, props.Text{
			Size:  9,
			Align: align.Left,
			Color: textColor,
		}))
	}

	qrCol := col.New(6).Add(image.NewFromBytes(qrBytes, extension.Png, props.Rect{
		Percent: 70,
		Center:  true,
	}))

	m.AddRow(45, sigCol, qrCol)

	// ---- Row 7: Footer ----
	m.AddRow(10,
		col.New(12).Add(text.New(footerText, props.Text{
			Size:  8,
			Align: align.Center,
			Color: textColor,
		})),
	)

	doc, err := m.Generate()
	if err != nil {
		return nil, fmt.Errorf("pdf: generation failed: %w", err)
	}

	return doc.GetBytes(), nil
}

// ============================================================
// HELPERS
// ============================================================

// fetchImage downloads a remote image and returns its bytes plus the
// maroto extension type. Falls back to PNG if the content-type is
// not recognised.
func (a *PDFAdapter) fetchImage(ctx context.Context, url string) ([]byte, extension.Type, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, "", fmt.Errorf("pdf: build image request: %w", err)
	}

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("pdf: fetch image: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("pdf: fetch image: status %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("pdf: read image: %w", err)
	}

	ext := extension.Png
	ct := strings.ToLower(resp.Header.Get("Content-Type"))
	switch {
	case strings.Contains(ct, "jpeg"), strings.Contains(ct, "jpg"):
		ext = extension.Jpg
	}
	return data, ext, nil
}

// substitute replaces {{key}} placeholders in text with values from
// the certificate. Unknown placeholders are left in place.
func substitute(s string, cert *certdomain.Certificate) string {
	if s == "" {
		return ""
	}
	replacements := map[string]string{
		"{{recipient_name}}":   cert.RecipientName,
		"{{context_title}}":    cert.Metadata.Get(certdomain.MetadataKeyContextTitle),
		"{{context_date}}":     cert.Metadata.Get(certdomain.MetadataKeyContextDate),
		"{{context_duration}}": cert.Metadata.Get(certdomain.MetadataKeyContextDuration),
		"{{issuer_name}}":      cert.Metadata.Get(certdomain.MetadataKeyIssuerName),
	}
	for k, v := range replacements {
		s = strings.ReplaceAll(s, k, v)
	}
	return s
}

// hexToColor converts "#RRGGBB" into a maroto color. Returns the
// supplied fallback RGB on empty or invalid input.
func hexToColor(hex string, fallbackR, fallbackG, fallbackB int) *props.Color {
	hex = strings.TrimPrefix(strings.TrimSpace(hex), "#")
	if len(hex) != 6 {
		return &props.Color{Red: fallbackR, Green: fallbackG, Blue: fallbackB}
	}
	var r, g, b int
	if _, err := fmt.Sscanf(hex, "%02x%02x%02x", &r, &g, &b); err != nil {
		return &props.Color{Red: fallbackR, Green: fallbackG, Blue: fallbackB}
	}
	return &props.Color{Red: r, Green: g, Blue: b}
}

// compile-time assertion
var _ certdomain.PDFGenerator = (*PDFAdapter)(nil)