package pdf

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/unidoc/unioffice/document"
	"github.com/unidoc/unioffice/measurement"
	"github.com/unidoc/unioffice/schema/soo/wml"
)

// Config contains configuration for the PDF provider
type Config struct {
	TemplatePath string
	DefaultFont  string
	DefaultSize  measurement.Distance
	PageSize     string
	Margins      Margins
}

// Margins represents page margins
type Margins struct {
	Top    measurement.Distance
	Bottom measurement.Distance
	Left   measurement.Distance
	Right  measurement.Distance
}

// Provider implements the PDF file generation provider
type Provider struct {
	config    *Config
	templates map[string]*document.Document
}

// NewProvider creates a new PDF provider
func NewProvider(config *Config) (*Provider, error) {
	if config == nil {
		config = &Config{
			TemplatePath: "./templates",
			DefaultFont:  "Calibri",
			DefaultSize:  measurement.Point * 11,
			PageSize:     "A4",
			Margins: Margins{
				Top:    measurement.Inch,
				Bottom: measurement.Inch,
				Left:   measurement.Inch,
				Right:  measurement.Inch,
			},
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

// GenerateFile generates a PDF file based on the request
func (p *Provider) GenerateFile(ctx context.Context, req *FileRequest) (*FileResponse, error) {
	doc, err := p.createDocument(req)
	if err != nil {
		return &FileResponse{
			Success: false,
			Error:   err.Error(),
		}, err
	}

	// Generate file content
	content, err := p.generateContent(doc)
	if err != nil {
		return &FileResponse{
			Success: false,
			Error:   err.Error(),
		}, err
	}

	return &FileResponse{
		Success:  true,
		Content:  content,
		MimeType: "application/pdf",
	}, nil
}

// GenerateFileToWriter generates a PDF file and writes it to the provided writer
func (p *Provider) GenerateFileToWriter(ctx context.Context, req *FileRequest, w io.Writer) error {
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
func (p *Provider) GetSupportedTypes() []FileType {
	return []FileType{FileTypePDF}
}

// ValidateRequest validates the file generation request
func (p *Provider) ValidateRequest(req *FileRequest) error {
	if req.Type != FileTypePDF {
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
func (p *Provider) createDocument(req *FileRequest) (*document.Document, error) {
	var doc *document.Document

	// Use template if specified
	if req.Template != "" {
		template, exists := p.templates[req.Template]
		if !exists {
			return nil, fmt.Errorf("template not found: %s", req.Template)
		}
		// Clone template
		doc = template.Clone()
	} else {
		// Create new document
		doc = document.New()
	}

	// Set up document properties
	if err := p.setupDocument(doc); err != nil {
		return nil, fmt.Errorf("failed to setup document: %w", err)
	}

	// Process content
	if err := p.processContent(doc, req.Data); err != nil {
		return nil, fmt.Errorf("failed to process content: %w", err)
	}

	return doc, nil
}

// setupDocument sets up document properties
func (p *Provider) setupDocument(doc *document.Document) error {
	// Set page size
	switch p.config.PageSize {
	case "A4":
		doc.Settings.SetPageSize(wml.ST_PageSizeA4)
	case "Letter":
		doc.Settings.SetPageSize(wml.ST_PageSizeLetter)
	case "Legal":
		doc.Settings.SetPageSize(wml.ST_PageSizeLegal)
	default:
		doc.Settings.SetPageSize(wml.ST_PageSizeA4)
	}

	// Set margins
	doc.Settings.SetMargins(
		p.config.Margins.Top,
		p.config.Margins.Bottom,
		p.config.Margins.Left,
		p.config.Margins.Right,
	)

	return nil
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

		// Center align title
		para.Properties().SetAlignment(wml.ST_JcCenter)
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

	// Process metadata
	if metadata, ok := data["metadata"].(map[string]interface{}); ok {
		if err := p.processMetadata(doc, metadata); err != nil {
			return fmt.Errorf("failed to process metadata: %w", err)
		}
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
	case "image":
		return p.addImage(doc, itemMap)
	case "page_break":
		return p.addPageBreak(doc)
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

	// Set alignment
	if align, ok := item["align"].(string); ok {
		switch align {
		case "center":
			para.Properties().SetAlignment(wml.ST_JcCenter)
		case "right":
			para.Properties().SetAlignment(wml.ST_JcRight)
		case "justify":
			para.Properties().SetAlignment(wml.ST_JcBoth)
		}
	}

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
			run.Properties().SetUnderline(wml.ST_UnderlineSingle, wml.ST_UnderlineColorAuto)
		}
	}

	// Set alignment
	if align, ok := item["align"].(string); ok {
		switch align {
		case "center":
			para.Properties().SetAlignment(wml.ST_JcCenter)
		case "right":
			para.Properties().SetAlignment(wml.ST_JcRight)
		case "justify":
			para.Properties().SetAlignment(wml.ST_JcBoth)
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

// addImage adds an image to the document
func (p *Provider) addImage(doc *document.Document, item map[string]interface{}) error {
	imagePath, ok := item["path"].(string)
	if !ok {
		return fmt.Errorf("image path is required")
	}

	// Check if image file exists
	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		return fmt.Errorf("image file not found: %s", imagePath)
	}

	// Add image to document
	img, err := doc.AddImage(imagePath)
	if err != nil {
		return fmt.Errorf("failed to add image: %w", err)
	}

	// Create paragraph with image
	para := doc.AddParagraph()
	run := para.AddRun()
	run.AddDrawingAnchored(img)

	// Set image size if specified
	if width, ok := item["width"].(float64); ok {
		img.SetWidth(measurement.Distance(width))
	}
	if height, ok := item["height"].(float64); ok {
		img.SetHeight(measurement.Distance(height))
	}

	return nil
}

// addPageBreak adds a page break to the document
func (p *Provider) addPageBreak(doc *document.Document) error {
	para := doc.AddParagraph()
	run := para.AddRun()
	run.AddBreak(wml.ST_BrTypePage)
	return nil
}

// processMetadata processes document metadata
func (p *Provider) processMetadata(doc *document.Document, metadata map[string]interface{}) error {
	// Set document properties
	if title, ok := metadata["title"].(string); ok {
		doc.Properties.SetTitle(title)
	}
	if author, ok := metadata["author"].(string); ok {
		doc.Properties.SetAuthor(author)
	}
	if subject, ok := metadata["subject"].(string); ok {
		doc.Properties.SetSubject(subject)
	}
	if keywords, ok := metadata["keywords"].(string); ok {
		doc.Properties.SetKeywords(keywords)
	}
	if description, ok := metadata["description"].(string); ok {
		doc.Properties.SetDescription(description)
	}

	return nil
}

// generateContent generates the document content as bytes
func (p *Provider) generateContent(doc *document.Document) ([]byte, error) {
	// Save to temporary file
	tmpFile, err := os.CreateTemp("", "pdf_*.docx")
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

// FileRequest and FileResponse are imported from the gateway package
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

const FileTypePDF FileType = "pdf"
