package docx

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/anasamu/microservices-library-go/filegen/types"
	"github.com/unidoc/unioffice/color"
	"github.com/unidoc/unioffice/document"
	"github.com/unidoc/unioffice/measurement"
	"github.com/unidoc/unioffice/schema/soo/wml"
)

// Config contains configuration for the DOCX provider
type Config struct {
	TemplatePath string
	DefaultFont  string
	DefaultSize  measurement.Distance
}

// Provider implements the DOCX file generation provider
type Provider struct {
	config    *Config
	templates map[string]*document.Document
}

// NewProvider creates a new DOCX provider
func NewProvider(config *Config) (*Provider, error) {
	if config == nil {
		config = &Config{
			TemplatePath: "./templates",
			DefaultFont:  "Calibri",
			DefaultSize:  measurement.Point * 11,
		}
	}

	provider := &Provider{
		config:    config,
		templates: make(map[string]*document.Document),
	}

	// Load templates
	if err := provider.loadTemplates(); err != nil {
		return nil, fmt.Errorf("failed to load templates: %w", err)
	}

	return provider, nil
}

// GenerateFile generates a DOCX file based on the request
func (p *Provider) GenerateFile(ctx context.Context, req *types.FileRequest) (*types.FileResponse, error) {
	doc, err := p.createDocument(req)
	if err != nil {
		return &types.FileResponse{
			Success: false,
			Error:   err.Error(),
		}, err
	}

	// Generate file content
	content, err := p.generateContent(doc)
	if err != nil {
		return &types.FileResponse{
			Success: false,
			Error:   err.Error(),
		}, err
	}

	return &types.FileResponse{
		Success:  true,
		Content:  content,
		MimeType: "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
	}, nil
}

// GenerateFileToWriter generates a DOCX file and writes it to the provided writer
func (p *Provider) GenerateFileToWriter(ctx context.Context, req *types.FileRequest, w io.Writer) error {
	doc, err := p.createDocument(req)
	if err != nil {
		return err
	}

	// Write to buffer first, then to writer
	buf, err := p.generateContent(doc)
	if err != nil {
		return err
	}

	_, err = w.Write(buf)
	return err
}

// GetSupportedTypes returns the file types supported by this provider
func (p *Provider) GetSupportedTypes() []types.FileType {
	return []types.FileType{types.FileTypeDOCX}
}

// ValidateRequest validates the file generation request
func (p *Provider) ValidateRequest(req *types.FileRequest) error {
	if req.Type != types.FileTypeDOCX {
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

// createDocument creates a new document based on the request
func (p *Provider) createDocument(req *types.FileRequest) (*document.Document, error) {
	var doc *document.Document
	var err error

	// Use template if specified
	if req.Template != "" {
		templatePath := filepath.Join(p.config.TemplatePath, req.Template+".docx")
		doc, err = document.Open(templatePath)
		if err != nil {
			return nil, fmt.Errorf("failed to open template %s: %w", req.Template, err)
		}
	} else {
		// Create new document
		doc = document.New()
	}

	// Process content
	if err := p.processContent(doc, req.Data); err != nil {
		return nil, fmt.Errorf("failed to process content: %w", err)
	}

	return doc, nil
}

// processContent processes the data and adds it to the document
func (p *Provider) processContent(doc *document.Document, data map[string]interface{}) error {
	// Add title if present
	if title, ok := data["title"].(string); ok {
		para := doc.AddParagraph()
		run := para.AddRun()
		run.AddText(title)
		run.Properties().SetSize(p.config.DefaultSize * 2)
		run.Properties().SetBold(true)
	}

	// Process content array
	if content, ok := data["content"].([]interface{}); ok {
		for _, item := range content {
			if err := p.processContentItem(doc, item); err != nil {
				return fmt.Errorf("failed to process content item: %w", err)
			}
		}
	}

	// Process simple text content
	if text, ok := data["text"].(string); ok {
		para := doc.AddParagraph()
		run := para.AddRun()
		run.AddText(text)
	}

	return nil
}

// processContentItem processes a single content item
func (p *Provider) processContentItem(doc *document.Document, item interface{}) error {
	itemMap, ok := item.(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid content item format")
	}

	itemType, ok := itemMap["type"].(string)
	if !ok {
		return fmt.Errorf("content item type is required")
	}

	switch itemType {
	case "heading":
		return p.addHeading(doc, itemMap)
	case "paragraph":
		return p.addParagraph(doc, itemMap)
	case "table":
		return p.addTable(doc, itemMap)
	case "list":
		return p.addList(doc, itemMap)
	default:
		return fmt.Errorf("unsupported content type: %s", itemType)
	}
}

// addHeading adds a heading to the document
func (p *Provider) addHeading(doc *document.Document, item map[string]interface{}) error {
	text, ok := item["text"].(string)
	if !ok {
		return fmt.Errorf("heading text is required")
	}

	level, _ := item["level"].(float64)
	if level == 0 {
		level = 1
	}

	para := doc.AddParagraph()
	run := para.AddRun()
	run.AddText(text)

	// Set heading style based on level
	size := p.config.DefaultSize * measurement.Distance(2-level*0.2)
	run.Properties().SetSize(size)
	run.Properties().SetBold(true)

	return nil
}

// addParagraph adds a paragraph to the document
func (p *Provider) addParagraph(doc *document.Document, item map[string]interface{}) error {
	text, ok := item["text"].(string)
	if !ok {
		return fmt.Errorf("paragraph text is required")
	}

	para := doc.AddParagraph()
	run := para.AddRun()
	run.AddText(text)

	// Apply style if specified
	if style, ok := item["style"].(string); ok {
		switch style {
		case "bold":
			run.Properties().SetBold(true)
		case "italic":
			run.Properties().SetItalic(true)
		case "underline":
			run.Properties().SetUnderline(wml.ST_UnderlineSingle, color.Auto)
		}
	}

	return nil
}

// addTable adds a table to the document
func (p *Provider) addTable(doc *document.Document, item map[string]interface{}) error {
	headers, ok := item["headers"].([]interface{})
	if !ok {
		return fmt.Errorf("table headers are required")
	}

	rows, ok := item["rows"].([]interface{})
	if !ok {
		return fmt.Errorf("table rows are required")
	}

	// Create table
	table := doc.AddTable()
	table.Properties().SetWidth(measurement.Inch * 6)

	// Add header row
	headerRow := table.AddRow()
	for _, header := range headers {
		cell := headerRow.AddCell()
		para := cell.AddParagraph()
		run := para.AddRun()
		run.AddText(fmt.Sprintf("%v", header))
		run.Properties().SetBold(true)
	}

	// Add data rows
	for _, row := range rows {
		rowData, ok := row.([]interface{})
		if !ok {
			continue
		}

		tableRow := table.AddRow()
		for _, cellData := range rowData {
			cell := tableRow.AddCell()
			para := cell.AddParagraph()
			run := para.AddRun()
			run.AddText(fmt.Sprintf("%v", cellData))
		}
	}

	return nil
}

// addList adds a list to the document
func (p *Provider) addList(doc *document.Document, item map[string]interface{}) error {
	items, ok := item["items"].([]interface{})
	if !ok {
		return fmt.Errorf("list items are required")
	}

	listType, _ := item["list_type"].(string)
	if listType == "" {
		listType = "bullet"
	}

	for _, listItem := range items {
		para := doc.AddParagraph()
		run := para.AddRun()

		if listType == "numbered" {
			run.AddText("• ")
		} else {
			run.AddText("• ")
		}

		run.AddText(fmt.Sprintf("%v", listItem))
	}

	return nil
}

// generateContent generates the document content as bytes
func (p *Provider) generateContent(doc *document.Document) ([]byte, error) {
	// Save to temporary file
	tmpFile, err := os.CreateTemp("", "docx_*.docx")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Save document
	if err := doc.SaveToFile(tmpFile.Name()); err != nil {
		return nil, fmt.Errorf("failed to save document: %w", err)
	}

	// Read file content
	content, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return content, nil
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
	files, err := filepath.Glob(filepath.Join(p.config.TemplatePath, "*.docx"))
	if err != nil {
		return fmt.Errorf("failed to read template directory: %w", err)
	}

	for _, file := range files {
		templateName := strings.TrimSuffix(filepath.Base(file), ".docx")

		// Load template
		doc, err := document.Open(file)
		if err != nil {
			continue // Skip invalid templates
		}

		p.templates[templateName] = doc
	}

	return nil
}

// Close closes the provider and cleans up resources
func (p *Provider) Close() error {
	// Close all template documents
	for _, doc := range p.templates {
		doc.Close()
	}
	p.templates = make(map[string]*document.Document)
	return nil
}

// FileRequest, FileResponse, FileOptions, and FileType are now imported from the types package
