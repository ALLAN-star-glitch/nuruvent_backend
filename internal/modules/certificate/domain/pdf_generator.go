// internal/modules/certificate/domain/pdf_generator.go

package domain

import "context"

// PDFGenerator renders a certificate to PDF bytes.
//
// The generator receives everything it needs — it must not consult the
// original context (BRD FR-CI-008).
type PDFGenerator interface {
	Render(ctx context.Context, in RenderInput) ([]byte, error)
}

// RenderInput is the material a renderer needs to produce a PDF.
type RenderInput struct {
	Certificate *Certificate
	Template    *Template
	VerifyURL   string
}