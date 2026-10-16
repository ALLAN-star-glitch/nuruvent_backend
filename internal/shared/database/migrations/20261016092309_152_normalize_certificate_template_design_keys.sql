// internal/app/adapters/certificate/pdf_adapter.go

package certificate

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/johnfercher/maroto/v2"
	"github.com/johnfercher/maroto/v2/pkg/components/col"
	"github.com/johnfercher/maroto/v2/pkg/components/image"
	"github.com/johnfercher/maroto/v2/pkg/components/line"
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
//
// Layout (landscape A4, ~297×210mm):
//
//	┌──────────────────────────────────────────────┐
//	│   ┌──────────────────────────────────────┐   │  ← frame
//	│   │           [ISSUER LOGO]              │   │
//	│   │      CERTIFICATE OF COMPLETION       │   │
//	│   │      ────────────────────            │   │
//	│   │           This certifies that        │   │
//	│   │                                      │   │
//	│   │           ALLAN MATHENGE             │   │
//	│   │           ───────────────            │   │
//	│   │                                      │   │
//	│   │    has successfully completed the    │   │
//	│   │    Web Design and Development        │   │
//	│   │    Workshop on 18 October 2026       │   │
//	│   │                                      │   │
//	│   │  [signature img]     [QR]            │   │
//	│   │  ──────────────      [VERIFIED]      │   │
//	│   │  Jane Doe                            │   │
//	│   │  Director of Training                │   │
//	│   │                                      │   │
//	│   │  Verify at nuruvent.com/verify       │   │
//	│   └──────────────────────────────────────┘   │
//	└──────────────────────────────────────────────┘
//
// Every design field is optional. The layout adapts.
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

	// ---- Text substitution ----
	titleText := strings.ToUpper(fallback(substitute(design.TitleText, cert), "Certificate of Completion"))
	subtitleText := fallback(substitute(design.SubtitleText, cert), "This certifies that")
	bodyText := substitute(design.BodyText, cert)
	if bodyText == "" {
		bodyText = substitute("has successfully completed {{context_title}} on {{context_date}}.", cert)
	}
	footerText := fallback(substitute(design.FooterText, cert), in.VerifyURL)
	recipientName := strings.TrimSpace(cert.RecipientName)
	signatoryName := substitute(design.SignatoryName, cert)
	signatoryTitle := substitute(design.SignatoryTitle, cert)

	// ---- Colors ----
	textColor := hexToColor(design.TextColor, 0x1A, 0x1A, 0x1A)
	accentColor := hexToColor(design.AccentColor, 0x0A, 0x6E, 0xBD)
	borderColor := hexToColor(design.BorderColor, accentColor.Red, accentColor.Green, accentColor.Blue)
	mutedColor := &props.Color{Red: 0x80, Green: 0x80, Blue: 0x80}
	watermarkColor := &props.Color{Red: 0xE8, Green: 0xE8, Blue: 0xE8}

	// ---- QR ----
	qrBytes, err := qrcode.Encode(in.VerifyURL, qrcode.Medium, 256)
	if err != nil {
		return nil, fmt.Errorf("pdf: qr generation failed: %w", err)
	}

	// ---- Config ----
	// Tighter margins so the border has room to sit inside them.
	cfg := config.NewBuilder().
		WithOrientation(orientation.Horizontal).
		WithLeftMargin(22).
		WithRightMargin(22).
		WithTopMargin(16).
		WithBottomMargin(14).
		Build()

	m := maroto.New(cfg)

	// ---- Border frame (single or double) ----
	// The frame is drawn as an inner "band" row on each edge. Top and
	// bottom rules use full-width cols; left and right rules use
	// thin-col padding at either side.
	if design.BorderStyle == "single" || design.BorderStyle == "double" {
		// Top rule
		m.AddRow(1,
			col.New(12).Add(line.New(props.Line{
				Color:     borderColor,
				Thickness: 0.7,
			})),
		)
		// Top padding
		m.AddRow(2)

		// Second rule for double border
		if design.BorderStyle == "double" {
			m.AddRow(1,
				col.New(1),
				col.New(10).Add(line.New(props.Line{
					Color:     borderColor,
					Thickness: 0.3,
				})),
				col.New(1),
			)
			m.AddRow(2)
		}
	} else {
		m.AddRow(3)
	}

	// ---- Logo (centered, optional) ----
	if design.LogoURL != "" {
		if data, ext, err := a.fetchImage(ctx, design.LogoURL); err == nil {
			m.AddRow(18,
				col.New(4),
				col.New(4).Add(image.NewFromBytes(data, ext, props.Rect{
					Percent: 90,
					Center:  true,
				})),
				col.New(4),
			)
		} else {
			m.AddRow(4)
		}
	} else {
		m.AddRow(2)
	}

	// ---- Title ----
	m.AddRow(14,
		col.New(12).Add(text.New(titleText, props.Text{
			Size:  22,
			Style: fontstyle.Bold,
			Align: align.Center,
			Color: accentColor,
		})),
	)

	// ---- Accent rule under title ----
	m.AddRow(3,
		col.New(4),
		col.New(4).Add(line.New(props.Line{
			Color:     accentColor,
			Thickness: 0.7,
		})),
		col.New(4),
	)

	// ---- Subtitle ----
	m.AddRow(8,
		col.New(12).Add(text.New(subtitleText, props.Text{
			Size:  11,
			Align: align.Center,
			Color: textColor,
		})),
	)

	// ---- Watermark (optional, faint) ----
	if design.WatermarkText != "" {
		m.AddRow(4,
			col.New(12).Add(text.New(strings.ToUpper(substitute(design.WatermarkText, cert)), props.Text{
				Size:  10,
				Align: align.Center,
				Color: watermarkColor,
			})),
		)
	}

	// ---- Recipient name ----
	m.AddRow(15,
		col.New(12).Add(text.New(recipientName, props.Text{
			Size:  28,
			Style: fontstyle.Bold,
			Align: align.Center,
			Color: accentColor,
		})),
	)

	// ---- Underline under recipient name ----
	m.AddRow(2,
		col.New(4),
		col.New(4).Add(line.New(props.Line{
			Color:     accentColor,
			Thickness: 0.4,
		})),
		col.New(4),
	)

	// ---- Body ----
	m.AddRow(13,
		col.New(2),
		col.New(8).Add(text.New(bodyText, props.Text{
			Size:  12,
			Align: align.Center,
			Color: textColor,
		})),
		col.New(2),
	)

	// ---- Bottom block: signature (left) | QR + badge (right) ----
	a.addBottomBlock(ctx, m, design, qrBytes, textColor, mutedColor, accentColor, signatoryName, signatoryTitle)

	// ---- Spacer ----
	m.AddRow(2)

	// ---- Footer ----
	m.AddRow(6,
		col.New(12).Add(text.New(footerText, props.Text{
			Size:  8,
			Align: align.Center,
			Color: mutedColor,
		})),
	)

	// ---- Bottom border rule + bottom double rule ----
	if design.BorderStyle == "single" || design.BorderStyle == "double" {
		m.AddRow(2)
		if design.BorderStyle == "double" {
			m.AddRow(1,
				col.New(1),
				col.New(10).Add(line.New(props.Line{
					Color:     borderColor,
					Thickness: 0.3,
				})),
				col.New(1),
			)
			m.AddRow(1)
		}
		m.AddRow(1,
			col.New(12).Add(line.New(props.Line{
				Color:     borderColor,
				Thickness: 0.7,
			})),
		)
	}

	doc, err := m.Generate()
	if err != nil {
		return nil, fmt.Errorf("pdf: generation failed: %w", err)
	}

	return doc.GetBytes(), nil
}

// addBottomBlock composes the signature (left half) and the QR + badge
// (right half). Both halves are entirely optional.
func (a *PDFAdapter) addBottomBlock(
	ctx context.Context,
	m interface {
		AddRow(height float64, cols ...interface{})
	},
	design certdomain.TemplateDesign,
	qrBytes []byte,
	textColor, mutedColor, accentColor *props.Color,
	signatoryName, signatoryTitle string,
) {
	hasSignature := design.SignatureURL != "" || signatoryName != "" || signatoryTitle != "" || design.SignatureLine
	hasQR := true // QR is always drawn

	// ============================================================
	// Row: signature image | QR
	// ============================================================
	if hasSignature {
		sigImageCol := col.New(6)
		if design.SignatureURL != "" {
			if data, ext, err := a.fetchImage(ctx, design.SignatureURL); err == nil {
				sigImageCol.Add(image.NewFromBytes(data, ext, props.Rect{
					Percent: 70,
					Center:  false,
				}))
			}
		}
		qrCol := col.New(6)
		if hasQR {
			qrCol.Add(image.NewFromBytes(qrBytes, extension.Png, props.Rect{
				Percent: 55,
				Center:  false,
			}))
		}
		m.AddRow(20, sigImageCol, qrCol)
	}

	// ============================================================
	// Row: signature rule | verify badge
	// ============================================================
	if hasSignature || design.VerifyBadge {
		ruleCol := col.New(6)
		if design.SignatureLine {
			ruleCol.Add(line.New(props.Line{
				Color:     textColor,
				Thickness: 0.4,
				OffsetPercent: 5,
			}))
		}
		badgeCol := col.New(6)
		if design.VerifyBadge {
			badgeCol.Add(text.New("✓ VERIFIED", props.Text{
				Size:  8,
				Style: fontstyle.Bold,
				Align: align.Left,
				Color: accentColor,
				Top:   1,
			}))
		}
		m.AddRow(6, ruleCol, badgeCol)
	}

	// ============================================================
	// Row: signatory name + title
	// ============================================================
	if signatoryName != "" {
		m.AddRow(5,
			col.New(6).Add(text.New(signatoryName, props.Text{
				Size:  11,
				Style: fontstyle.Bold,
				Align: align.Left,
				Color: textColor,
			})),
			col.New(6),
		)
	}
	if signatoryTitle != "" {
		m.AddRow(4,
			col.New(6).Add(text.New(signatoryTitle, props.Text{
				Size:  9,
				Align: align.Left,
				Color: mutedColor,
			})),
			col.New(6),
		)
	}
}

// ============================================================
// HELPERS
// ============================================================

// fetchImage downloads a remote image and returns its bytes plus the
// maroto extension type. Falls back to PNG if the content-type is
// not recognised.
func (a *PDFAdapter) fetchImage(ctx context.Context, url string) ([]byte, extension.Type, error) {
	if url == "" {
		return nil, "", errors.New("pdf: empty image url")
	}

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

// fallback returns v when s is empty.
func fallback(s, v string) string {
	if strings.TrimSpace(s) == "" {
		return v
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