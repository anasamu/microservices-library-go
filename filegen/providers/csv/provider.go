package csv

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/anasamu/microservices-library-go/filegen/types"
)

// Config contains configuration for the CSV provider
type Config struct {
	TemplatePath     string
	DefaultDelimiter rune
	DefaultEncoding  string
}

// Provider implements the CSV file generation provider
type Provider struct {
	config    *Config
	templates map[string][]string
}

// NewProvider creates a new CSV provider
func NewProvider(config *Config) (*Provider, error) {
	if config == nil {
		config = &Config{
			TemplatePath:     "./templates",
			DefaultDelimiter: ',',
			DefaultEncoding:  "utf-8",
		}
	}

	provider := &Provider{
		config:    config,
		templates: make(map[string][]string),
	}

	// Load templates
	if err := provider.loadTemplates(); err != nil {
		return nil, fmt.Errorf("failed to load templates: %w", err)
	}

	return provider, nil
}

// GenerateFile generates a CSV file based on the request
func (p *Provider) GenerateFile(ctx context.Context, req *types.FileRequest) (*types.FileResponse, error) {
	content, err := p.generateContent(req)
	if err != nil {
		return &types.FileResponse{
			Success: false,
			Error:   err.Error(),
		}, err
	}

	return &types.FileResponse{
		Success:  true,
		Content:  content,
		MimeType: "text/csv",
	}, nil
}

// GenerateFileToWriter generates a CSV file and writes it to the provided writer
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
	return []types.FileType{types.FileTypeCSV}
}

// ValidateRequest validates the file generation request
func (p *Provider) ValidateRequest(req *types.FileRequest) error {
	if req.Type != types.FileTypeCSV {
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

// generateContent generates the CSV content as bytes
func (p *Provider) generateContent(req *types.FileRequest) ([]byte, error) {
	// Create a buffer to write CSV data
	var buf strings.Builder
	writer := csv.NewWriter(&buf)

	// Get delimiter from options
	delimiter := p.config.DefaultDelimiter
	if req.Options.Delimiter != "" {
		delimiter = rune(req.Options.Delimiter[0])
	}
	writer.Comma = delimiter

	// Process data
	if err := p.processData(writer, req.Data, req.Template); err != nil {
		return nil, fmt.Errorf("failed to process data: %w", err)
	}

	// Flush the writer
	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, fmt.Errorf("failed to write CSV: %w", err)
	}

	return []byte(buf.String()), nil
}

// processData processes the data and writes it to the CSV writer
func (p *Provider) processData(writer *csv.Writer, data map[string]interface{}, templateName string) error {
	// Process headers
	if headers, ok := data["headers"].([]interface{}); ok {
		headerStrings := make([]string, len(headers))
		for i, header := range headers {
			headerStrings[i] = fmt.Sprintf("%v", header)
		}
		if err := writer.Write(headerStrings); err != nil {
			return fmt.Errorf("failed to write headers: %w", err)
		}
	}

	// Process rows
	if rows, ok := data["rows"].([]interface{}); ok {
		for _, row := range rows {
			rowData, ok := row.([]interface{})
			if !ok {
				continue
			}

			rowStrings := make([]string, len(rowData))
			for i, cell := range rowData {
				rowStrings[i] = fmt.Sprintf("%v", cell)
			}

			if err := writer.Write(rowStrings); err != nil {
				return fmt.Errorf("failed to write row: %w", err)
			}
		}
	}

	// Process simple data array (without headers)
	if dataArray, ok := data["data"].([]interface{}); ok {
		for _, row := range dataArray {
			rowData, ok := row.([]interface{})
			if !ok {
				continue
			}

			rowStrings := make([]string, len(rowData))
			for i, cell := range rowData {
				rowStrings[i] = fmt.Sprintf("%v", cell)
			}

			if err := writer.Write(rowStrings); err != nil {
				return fmt.Errorf("failed to write data row: %w", err)
			}
		}
	}

	// Process records (array of maps)
	if records, ok := data["records"].([]interface{}); ok {
		// First, write headers from the first record
		if len(records) > 0 {
			firstRecord, ok := records[0].(map[string]interface{})
			if ok {
				headers := make([]string, 0, len(firstRecord))
				for key := range firstRecord {
					headers = append(headers, key)
				}
				if err := writer.Write(headers); err != nil {
					return fmt.Errorf("failed to write record headers: %w", err)
				}
			}
		}

		// Then write data rows
		for _, record := range records {
			recordMap, ok := record.(map[string]interface{})
			if !ok {
				continue
			}

			// Get values in the same order as headers
			values := make([]string, 0, len(recordMap))
			for key := range recordMap {
				values = append(values, fmt.Sprintf("%v", recordMap[key]))
			}

			if err := writer.Write(values); err != nil {
				return fmt.Errorf("failed to write record: %w", err)
			}
		}
	}

	// Process template data
	if templateName != "" {
		if template, exists := p.templates[templateName]; exists {
			for _, row := range template {
				// Split template row by delimiter
				fields := strings.Split(row, string(p.config.DefaultDelimiter))
				if err := writer.Write(fields); err != nil {
					return fmt.Errorf("failed to write template row: %w", err)
				}
			}
		}
	}

	return nil
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
	files, err := filepath.Glob(filepath.Join(p.config.TemplatePath, "*.csv"))
	if err != nil {
		return fmt.Errorf("failed to read template directory: %w", err)
	}

	for _, file := range files {
		templateName := strings.TrimSuffix(filepath.Base(file), ".csv")

		// Load template
		template, err := p.loadTemplateFile(file)
		if err != nil {
			continue // Skip invalid templates
		}

		p.templates[templateName] = template
	}

	return nil
}

// loadTemplateFile loads a single template file
func (p *Provider) loadTemplateFile(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Comma = p.config.DefaultDelimiter

	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	// Convert records to strings
	template := make([]string, len(records))
	for i, record := range records {
		template[i] = strings.Join(record, string(p.config.DefaultDelimiter))
	}

	return template, nil
}

// Close closes the provider and cleans up resources
func (p *Provider) Close() error {
	p.templates = make(map[string][]string)
	return nil
}

// FileRequest, FileResponse, FileOptions, and FileType are now imported from the types package
