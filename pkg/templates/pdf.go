package templates

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

// PDFGenerator handles PDF generation from HTML
type PDFGenerator struct {
	engine      *Engine
	wkhtmltopdf string // Path to wkhtmltopdf binary
}

// NewPDFGenerator creates a new PDF generator
func NewPDFGenerator(engine *Engine) (*PDFGenerator, error) {
	// Find wkhtmltopdf binary
	wkhtmltopdf, err := exec.LookPath("wkhtmltopdf")
	if err != nil {
		return nil, fmt.Errorf("wkhtmltopdf not found in PATH: %w", err)
	}

	return &PDFGenerator{
		engine:      engine,
		wkhtmltopdf: wkhtmltopdf,
	}, nil
}

// PDFOptions contains options for PDF generation
type PDFOptions struct {
	PageSize        string  // A4, Letter, etc.
	Orientation     string  // Portrait, Landscape
	MarginTop       string  // e.g., "10mm"
	MarginRight     string
	MarginBottom    string
	MarginLeft      string
	HeaderHTML      string  // HTML for header
	FooterHTML      string  // HTML for footer
	EnableJavaScript bool
	Quality         int     // 0-100
}

// DefaultPDFOptions returns default PDF options
func DefaultPDFOptions() *PDFOptions {
	return &PDFOptions{
		PageSize:     "A4",
		Orientation:  "Portrait",
		MarginTop:    "10mm",
		MarginRight:  "10mm",
		MarginBottom: "10mm",
		MarginLeft:   "10mm",
		Quality:      94,
	}
}

// RenderPDF renders a template to PDF
func (g *PDFGenerator) RenderPDF(templateName string, data interface{}, opts *PDFOptions) ([]byte, error) {
	if opts == nil {
		opts = DefaultPDFOptions()
	}

	// First render HTML
	html, err := g.engine.RenderHTML(templateName, data)
	if err != nil {
		return nil, fmt.Errorf("failed to render HTML: %w", err)
	}

	// Generate PDF from HTML
	return g.HTMLToPDF(html, opts)
}

// HTMLToPDF converts HTML to PDF
func (g *PDFGenerator) HTMLToPDF(html string, opts *PDFOptions) ([]byte, error) {
	if opts == nil {
		opts = DefaultPDFOptions()
	}

	// Create temp files for input and output
	tmpDir, err := os.MkdirTemp("", "pdf-gen-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	inputFile := filepath.Join(tmpDir, "input.html")
	outputFile := filepath.Join(tmpDir, "output.pdf")

	// Write HTML to temp file
	if err := os.WriteFile(inputFile, []byte(html), 0644); err != nil {
		return nil, fmt.Errorf("failed to write HTML file: %w", err)
	}

	// Build wkhtmltopdf command
	args := g.buildArgs(inputFile, outputFile, opts)

	// Execute wkhtmltopdf
	cmd := exec.Command(g.wkhtmltopdf, args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("wkhtmltopdf failed: %w, stderr: %s", err, stderr.String())
	}

	// Read generated PDF
	pdfData, err := os.ReadFile(outputFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read PDF: %w", err)
	}

	return pdfData, nil
}

// RenderPDFToWriter renders a template to PDF and writes to a writer
func (g *PDFGenerator) RenderPDFToWriter(templateName string, data interface{}, writer io.Writer, opts *PDFOptions) error {
	pdfData, err := g.RenderPDF(templateName, data, opts)
	if err != nil {
		return err
	}

	_, err = writer.Write(pdfData)
	return err
}

func (g *PDFGenerator) buildArgs(inputFile, outputFile string, opts *PDFOptions) []string {
	args := []string{
		"--page-size", opts.PageSize,
		"--orientation", opts.Orientation,
		"--margin-top", opts.MarginTop,
		"--margin-right", opts.MarginRight,
		"--margin-bottom", opts.MarginBottom,
		"--margin-left", opts.MarginLeft,
		"--quiet",
	}

	if opts.EnableJavaScript {
		args = append(args, "--enable-javascript")
	} else {
		args = append(args, "--disable-javascript")
	}

	if opts.Quality > 0 {
		args = append(args, "--image-quality", fmt.Sprintf("%d", opts.Quality))
	}

	if opts.HeaderHTML != "" {
		// Write header to temp file
		headerFile := filepath.Join(filepath.Dir(inputFile), "header.html")
		os.WriteFile(headerFile, []byte(opts.HeaderHTML), 0644)
		args = append(args, "--header-html", headerFile)
	}

	if opts.FooterHTML != "" {
		// Write footer to temp file
		footerFile := filepath.Join(filepath.Dir(inputFile), "footer.html")
		os.WriteFile(footerFile, []byte(opts.FooterHTML), 0644)
		args = append(args, "--footer-html", footerFile)
	}

	args = append(args, inputFile, outputFile)

	return args
}
