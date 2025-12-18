package templates

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"os"
	"path/filepath"
)

// Engine handles template rendering
type Engine struct {
	baseDir   string
	templates map[string]*template.Template
	funcMap   template.FuncMap
}

// NewEngine creates a new template engine
func NewEngine(baseDir string) *Engine {
	return &Engine{
		baseDir:   baseDir,
		templates: make(map[string]*template.Template),
		funcMap:   DefaultFuncMap(),
	}
}

// DefaultFuncMap returns default template functions
func DefaultFuncMap() template.FuncMap {
	return template.FuncMap{
		"formatCurrency": func(amount float64) string {
			return fmt.Sprintf("$%.2f", amount)
		},
		"formatDate": func(date string) string {
			// Add date formatting logic
			return date
		},
		"add": func(a, b int) int {
			return a + b
		},
		"multiply": func(a, b float64) float64 {
			return a * b
		},
	}
}

// LoadTemplate loads a template from file
func (e *Engine) LoadTemplate(name string, filePath string) error {
	fullPath := filepath.Join(e.baseDir, filePath)

	// Check if file exists
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return fmt.Errorf("template file not found: %s", fullPath)
	}

	// Parse template
	tmpl, err := template.New(name).Funcs(e.funcMap).ParseFiles(fullPath)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	e.templates[name] = tmpl
	return nil
}

// LoadTemplates loads multiple templates from a directory
func (e *Engine) LoadTemplates(dir string, pattern string) error {
	searchPath := filepath.Join(e.baseDir, dir, pattern)

	files, err := filepath.Glob(searchPath)
	if err != nil {
		return fmt.Errorf("failed to glob templates: %w", err)
	}

	for _, file := range files {
		name := filepath.Base(file)
		name = name[:len(name)-len(filepath.Ext(name))] // Remove extension

		relPath, err := filepath.Rel(e.baseDir, file)
		if err != nil {
			return fmt.Errorf("failed to get relative path: %w", err)
		}

		if err := e.LoadTemplate(name, relPath); err != nil {
			return err
		}
	}

	return nil
}

// RenderHTML renders a template to HTML string
func (e *Engine) RenderHTML(templateName string, data interface{}) (string, error) {
	tmpl, ok := e.templates[templateName]
	if !ok {
		return "", fmt.Errorf("template not found: %s", templateName)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

// RenderHTMLToWriter renders a template to a writer
func (e *Engine) RenderHTMLToWriter(templateName string, data interface{}, writer io.Writer) error {
	tmpl, ok := e.templates[templateName]
	if !ok {
		return fmt.Errorf("template not found: %s", templateName)
	}

	if err := tmpl.Execute(writer, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	return nil
}

// AddFunc adds a custom template function
func (e *Engine) AddFunc(name string, fn interface{}) {
	e.funcMap[name] = fn
}

// HasTemplate checks if a template exists
func (e *Engine) HasTemplate(name string) bool {
	_, ok := e.templates[name]
	return ok
}
