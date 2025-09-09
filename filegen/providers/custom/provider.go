package custom

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/anasamu/microservices-library-go/filegen/types"
)

// Config contains configuration for the custom provider
type Config struct {
	TemplatePath  string
	OutputPath    string
	DefaultFormat string
}

// Provider implements the custom file generation provider
type Provider struct {
	config    *Config
	templates map[string]*template.Template
}

// NewProvider creates a new custom provider
func NewProvider(config *Config) (*Provider, error) {
	if config == nil {
		config = &Config{
			TemplatePath:  "./templates",
			OutputPath:    "./output",
			DefaultFormat: "text",
		}
	}

	provider := &Provider{
		config:    config,
		templates: make(map[string]*template.Template),
	}

	// Load templates
	if err := provider.loadTemplates(); err != nil {
		return nil, fmt.Errorf("failed to load templates: %w", err)
	}

	return provider, nil
}

// GenerateFile generates a custom file based on the request
func (p *Provider) GenerateFile(ctx context.Context, req *types.FileRequest) (*types.FileResponse, error) {
	content, err := p.generateContent(req)
	if err != nil {
		return &types.FileResponse{
			Success: false,
			Error:   err.Error(),
		}, err
	}

	// Determine MIME type based on format
	mimeType := p.getMimeType(req.Options.Format)

	return &types.FileResponse{
		Success:  true,
		Content:  content,
		MimeType: mimeType,
	}, nil
}

// GenerateFileToWriter generates a custom file and writes it to the provided writer
func (p *Provider) GenerateFileToWriter(ctx context.Context, req *types.FileRequest, w io.Writer) error {
	content, err := p.generateContent(req)
	if err != nil {
		return err
	}

	_, err = w.Write(content)
	return err
}

// GetSupportedTypes returns the file types supported by this provider
func (p *Provider) GetSupportedTypes() []types.FileType {
	return []types.FileType{types.FileTypeCustom}
}

// ValidateRequest validates the file generation request
func (p *Provider) ValidateRequest(req *types.FileRequest) error {
	if req.Type != types.FileTypeCustom {
		return fmt.Errorf("unsupported file type: %s", req.Type)
	}

	if req.Data == nil {
		return fmt.Errorf("data is required")
	}

	return nil
}

// GetTemplateList returns available templates for this provider
func (p *Provider) GetTemplateList() ([]string, error) {
	templates := make([]string, 0, len(p.templates))
	for name := range p.templates {
		templates = append(templates, name)
	}
	return templates, nil
}

// generateContent generates the custom file content as bytes
func (p *Provider) generateContent(req *types.FileRequest) ([]byte, error) {
	// Use template if specified
	if req.Template != "" {
		return p.generateFromTemplate(req)
	}

	// Generate from data directly
	return p.generateFromData(req)
}

// generateFromTemplate generates content using a template
func (p *Provider) generateFromTemplate(req *types.FileRequest) ([]byte, error) {
	tmpl, exists := p.templates[req.Template]
	if !exists {
		return nil, fmt.Errorf("template not found: %s", req.Template)
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, req.Data); err != nil {
		return nil, fmt.Errorf("failed to execute template: %w", err)
	}

	return []byte(buf.String()), nil
}

// generateFromData generates content directly from data
func (p *Provider) generateFromData(req *types.FileRequest) ([]byte, error) {
	format := req.Options.Format
	if format == "" {
		format = p.config.DefaultFormat
	}

	switch format {
	case "json":
		return p.generateJSON(req.Data)
	case "xml":
		return p.generateXML(req.Data)
	case "yaml":
		return p.generateYAML(req.Data)
	case "html":
		return p.generateHTML(req.Data)
	case "text":
		return p.generateText(req.Data)
	case "markdown":
		return p.generateMarkdown(req.Data)
	default:
		return p.generateText(req.Data)
	}
}

// generateJSON generates JSON content
func (p *Provider) generateJSON(data map[string]interface{}) ([]byte, error) {
	// Simple JSON generation (in production, use encoding/json)
	var result strings.Builder
	result.WriteString("{\n")

	first := true
	for key, value := range data {
		if !first {
			result.WriteString(",\n")
		}
		first = false

		result.WriteString(fmt.Sprintf("  \"%s\": ", key))
		switch v := value.(type) {
		case string:
			result.WriteString(fmt.Sprintf("\"%s\"", v))
		case int, int64, float64:
			result.WriteString(fmt.Sprintf("%v", v))
		case bool:
			result.WriteString(fmt.Sprintf("%t", v))
		default:
			result.WriteString(fmt.Sprintf("\"%v\"", v))
		}
	}

	result.WriteString("\n}")
	return []byte(result.String()), nil
}

// generateXML generates XML content
func (p *Provider) generateXML(data map[string]interface{}) ([]byte, error) {
	var result strings.Builder
	result.WriteString("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n")
	result.WriteString("<root>\n")

	for key, value := range data {
		result.WriteString(fmt.Sprintf("  <%s>%v</%s>\n", key, value, key))
	}

	result.WriteString("</root>")
	return []byte(result.String()), nil
}

// generateYAML generates YAML content
func (p *Provider) generateYAML(data map[string]interface{}) ([]byte, error) {
	var result strings.Builder

	for key, value := range data {
		result.WriteString(fmt.Sprintf("%s: %v\n", key, value))
	}

	return []byte(result.String()), nil
}

// generateHTML generates HTML content
func (p *Provider) generateHTML(data map[string]interface{}) ([]byte, error) {
	var result strings.Builder
	result.WriteString("<!DOCTYPE html>\n<html>\n<head>\n")
	result.WriteString("  <title>Generated Document</title>\n")
	result.WriteString("</head>\n<body>\n")

	// Add title if present
	if title, ok := data["title"].(string); ok {
		result.WriteString(fmt.Sprintf("  <h1>%s</h1>\n", title))
	}

	// Add content
	if content, ok := data["content"].(string); ok {
		result.WriteString(fmt.Sprintf("  <p>%s</p>\n", content))
	}

	// Add table if present
	if table, ok := data["table"].(map[string]interface{}); ok {
		result.WriteString("  <table border=\"1\">\n")

		// Headers
		if headers, ok := table["headers"].([]interface{}); ok {
			result.WriteString("    <tr>\n")
			for _, header := range headers {
				result.WriteString(fmt.Sprintf("      <th>%v</th>\n", header))
			}
			result.WriteString("    </tr>\n")
		}

		// Rows
		if rows, ok := table["rows"].([]interface{}); ok {
			for _, row := range rows {
				if rowData, ok := row.([]interface{}); ok {
					result.WriteString("    <tr>\n")
					for _, cell := range rowData {
						result.WriteString(fmt.Sprintf("      <td>%v</td>\n", cell))
					}
					result.WriteString("    </tr>\n")
				}
			}
		}

		result.WriteString("  </table>\n")
	}

	result.WriteString("</body>\n</html>")
	return []byte(result.String()), nil
}

// generateText generates plain text content
func (p *Provider) generateText(data map[string]interface{}) ([]byte, error) {
	var result strings.Builder

	// Add title if present
	if title, ok := data["title"].(string); ok {
		result.WriteString(fmt.Sprintf("%s\n", title))
		result.WriteString(strings.Repeat("=", len(title)) + "\n\n")
	}

	// Add content
	for key, value := range data {
		if key == "title" {
			continue
		}

		result.WriteString(fmt.Sprintf("%s: %v\n", strings.Title(key), value))
	}

	return []byte(result.String()), nil
}

// generateMarkdown generates Markdown content
func (p *Provider) generateMarkdown(data map[string]interface{}) ([]byte, error) {
	var result strings.Builder

	// Add title if present
	if title, ok := data["title"].(string); ok {
		result.WriteString(fmt.Sprintf("# %s\n\n", title))
	}

	// Add content
	if content, ok := data["content"].(string); ok {
		result.WriteString(fmt.Sprintf("%s\n\n", content))
	}

	// Add table if present
	if table, ok := data["table"].(map[string]interface{}); ok {
		// Headers
		if headers, ok := table["headers"].([]interface{}); ok {
			result.WriteString("|")
			for _, header := range headers {
				result.WriteString(fmt.Sprintf(" %v |", header))
			}
			result.WriteString("\n")

			// Separator
			result.WriteString("|")
			for range headers {
				result.WriteString(" --- |")
			}
			result.WriteString("\n")
		}

		// Rows
		if rows, ok := table["rows"].([]interface{}); ok {
			for _, row := range rows {
				if rowData, ok := row.([]interface{}); ok {
					result.WriteString("|")
					for _, cell := range rowData {
						result.WriteString(fmt.Sprintf(" %v |", cell))
					}
					result.WriteString("\n")
				}
			}
		}
		result.WriteString("\n")
	}

	// Add other data as key-value pairs
	for key, value := range data {
		if key == "title" || key == "content" || key == "table" {
			continue
		}

		result.WriteString(fmt.Sprintf("**%s:** %v\n\n", strings.Title(key), value))
	}

	return []byte(result.String()), nil
}

// getMimeType returns the MIME type for a given format
func (p *Provider) getMimeType(format string) string {
	switch format {
	case "json":
		return "application/json"
	case "xml":
		return "application/xml"
	case "yaml":
		return "application/x-yaml"
	case "html":
		return "text/html"
	case "markdown":
		return "text/markdown"
	case "text":
		return "text/plain"
	default:
		return "text/plain"
	}
}

// loadTemplates loads available templates from the template directory
func (p *Provider) loadTemplates() error {
	if p.config.TemplatePath == "" {
		return nil
	}

	// Check if template directory exists
	if _, err := os.Stat(p.config.TemplatePath); os.IsNotExist(err) {
		return nil // No templates to load
	}

	// Read template files
	files, err := filepath.Glob(filepath.Join(p.config.TemplatePath, "*.tmpl"))
	if err != nil {
		return fmt.Errorf("failed to read template directory: %w", err)
	}

	for _, file := range files {
		templateName := strings.TrimSuffix(filepath.Base(file), ".tmpl")

		// Load template
		tmpl, err := template.ParseFiles(file)
		if err != nil {
			continue // Skip invalid templates
		}

		p.templates[templateName] = tmpl
	}

	return nil
}

// Close closes the provider and cleans up resources
func (p *Provider) Close() error {
	p.templates = make(map[string]*template.Template)
	return nil
}

// types.FileRequest and types.FileResponse are imported from the gateway package
// We need to define them here or import them properly
type FileRequest struct {
	Type       string                 `json:"type"`
	Template   string                 `json:"template,omitempty"`
	Data       map[string]interface{} `json:"data"`
	Options    FileOptions            `json:"options,omitempty"`
	OutputPath string                 `json:"output_path,omitempty"`
}

type FileResponse struct {
	Success  bool   `json:"success"`
	FilePath string `json:"file_path,omitempty"`
	FileSize int64  `json:"file_size,omitempty"`
	Error    string `json:"error,omitempty"`
	Content  []byte `json:"content,omitempty"`
	MimeType string `json:"mime_type,omitempty"`
}

type FileOptions struct {
	Format    string                 `json:"format,omitempty"`
	Encoding  string                 `json:"encoding,omitempty"`
	Delimiter string                 `json:"delimiter,omitempty"`
	Headers   []string               `json:"headers,omitempty"`
	Metadata  map[string]string      `json:"metadata,omitempty"`
	Custom    map[string]interface{} `json:"custom,omitempty"`
}

type FileType string

const FileTypeCustom FileType = "custom"
