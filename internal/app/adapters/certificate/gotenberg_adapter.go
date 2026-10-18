// internal/app/adapters/certificate/gotenberg_adapter.go

package certificate

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/base64"
	"fmt"
	"html/template"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"time"

	"github.com/skip2/go-qrcode"

	certdomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/certificate/domain"
)

//go:embed templates/certificate.html
var certTemplate string

type GotenbergAdapter struct {
	baseURL    string
	httpClient *http.Client
	tmpl       *template.Template
}

func NewGotenbergAdapter(baseURL string) *GotenbergAdapter {
	tmpl := template.Must(template.New("cert").Parse(certTemplate))
	return &GotenbergAdapter{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		tmpl:       tmpl,
	}
}

func (a *GotenbergAdapter) Render(
	ctx context.Context,
	in certdomain.RenderInput,
) ([]byte, error) {
	if in.Certificate == nil || in.Template == nil {
		return nil, fmt.Errorf("pdf: missing certificate or template")
	}

	// 1. Generate QR Code as Base64 Data URL
	qrBytes, err := qrcode.Encode(in.VerifyURL, qrcode.Medium, 256)
	if err != nil {
		return nil, fmt.Errorf("pdf: qr generation failed: %w", err)
	}
	base64Str := base64.StdEncoding.EncodeToString(qrBytes)
	base64Str = strings.ReplaceAll(base64Str, "\n", "")
	qrDataURL := "data:image/png;base64," + base64Str

	// 2. Resolve text (with placeholders)
	design := in.Template.Design
	cert := in.Certificate

	title := substitute(design.TitleText, cert)
	if title == "" {
		title = "Certificate of Completion"
	}

	subtitle := substitute(design.SubtitleText, cert)
	if subtitle == "" {
		subtitle = "This certifies that"
	}

	body := substitute(design.BodyText, cert)
	if body == "" {
		body = substitute("has successfully completed {{context_title}} on {{context_date}}.", cert)
	}

	footer := substitute(design.FooterText, cert)

	// 3. Build Data Model for HTML
	data := struct {
		BackgroundColor    string
		TextColor          string
		AccentColor        string
		BorderColor        string
		BorderStyle        string
		LogoURL            string
		BackgroundURL      string
		BackgroundOpacity  int
		BackgroundSize     string
		BackgroundPosition string
		SignatureURL       string
		TitleText          string
		SubtitleText       string
		BodyText           string
		RecipientName      string
		SignatoryName      string
		SignatoryTitle     string
		SignatureLine      bool
		QRCodeDataURL      template.URL // Use template.URL to prevent HTML escaping of the Base64 string
		VerifyBadge        bool
		WatermarkText      string
		FooterText         string
		VerificationCode   string
	}{
		BackgroundColor:    fallback(design.BackgroundColor, "#FFFFFF"),
		TextColor:          fallback(design.TextColor, "#1A1A1A"),
		AccentColor:        fallback(design.AccentColor, "#0A6EBD"),
		BorderColor:        fallback(design.BorderColor, fallback(design.AccentColor, "#0A6EBD")),
		BorderStyle:        fallback(design.BorderStyle, "none"),
		LogoURL:            design.LogoURL,
		BackgroundURL:      design.BackgroundURL,
		BackgroundOpacity:  clampOpacity(design.BackgroundOpacity),
		BackgroundSize:     fallback(design.BackgroundSize, "cover"),
		BackgroundPosition: fallback(design.BackgroundPosition, "center"),
		SignatureURL:       design.SignatureURL,
		TitleText:          strings.ToUpper(title),
		SubtitleText:       subtitle,
		BodyText:           body,
		RecipientName:      strings.TrimSpace(cert.RecipientName),
		SignatoryName:      substitute(design.SignatoryName, cert),
		SignatoryTitle:     substitute(design.SignatoryTitle, cert),
		SignatureLine:      design.SignatureLine,
		QRCodeDataURL:      template.URL(qrDataURL),
		VerifyBadge:        design.VerifyBadge,
		WatermarkText:      design.WatermarkText,
		FooterText:         fallback(footer, in.VerifyURL),
		VerificationCode:   string(cert.VerificationCode),
	}

	// 4. Render HTML to string
	var htmlBuf bytes.Buffer
	if err := a.tmpl.Execute(&htmlBuf, data); err != nil {
		return nil, fmt.Errorf("pdf: template render failed: %w", err)
	}

	// 5. Send to Gotenberg
	bodyBuf := &bytes.Buffer{}
	writer := multipart.NewWriter(bodyBuf)

	part, _ := writer.CreateFormFile("files", "index.html")
	io.WriteString(part, htmlBuf.String())

	writer.WriteField("printBackground", "true")
	writer.WriteField("paperWidth", "11.69") // A4 Landscape
	writer.WriteField("paperHeight", "8.27")
	writer.WriteField("marginTop", "0")
	writer.WriteField("marginBottom", "0")
	writer.WriteField("marginLeft", "0")
	writer.WriteField("marginRight", "0")
	writer.Close()

	url := fmt.Sprintf("%s/forms/chromium/convert/html", a.baseURL)
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, url, bodyBuf)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("pdf: gotenberg request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("pdf: gotenberg returned %d: %s", resp.StatusCode, string(errBody))
	}

	return io.ReadAll(resp.Body)
}

func substitute(s string, cert *certdomain.Certificate) string {
	if s == "" {
		return ""
	}
	repl := map[string]string{
		"{{recipient_name}}":   cert.RecipientName,
		"{{context_title}}":    cert.Metadata.Get(certdomain.MetadataKeyContextTitle),
		"{{context_date}}":     cert.Metadata.Get(certdomain.MetadataKeyContextDate),
		"{{context_duration}}": cert.Metadata.Get(certdomain.MetadataKeyContextDuration),
		"{{issuer_name}}":      cert.Metadata.Get(certdomain.MetadataKeyIssuerName),
	}
	for k, v := range repl {
		s = strings.ReplaceAll(s, k, v)
	}
	return s
}

func fallback(s, def string) string {
	if strings.TrimSpace(s) == "" {
		return def
	}
	return s
}

// clampOpacity defends against legacy rows or un-normalized designs.
// The service layer already calls NormalizeBackground on save, but old
// rows might still carry 0 / negative / >100 values.
func clampOpacity(n int) int {
	if n <= 0 {
		return 8
	}
	if n > 100 {
		return 100
	}
	return n
}

var _ certdomain.PDFGenerator = (*GotenbergAdapter)(nil)