package excel

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/anasamu/microservices-library-go/filegen/types"
	"github.com/xuri/excelize/v2"
)

// Config contains configuration for the Excel provider
type Config struct {
	TemplatePath string
	DefaultSheet string
}

// Provider implements the Excel file generation provider
type Provider struct {
	config    *Config
	templates map[string]*excelize.File
}

// NewProvider creates a new Excel provider
func NewProvider(config *Config) (*Provider, error) {
	if config == nil {
		config = &Config{
			TemplatePath: "./templates",
			DefaultSheet: "Sheet1",
		}
	}

	provider := &Provider{
		config:    config,
		templates: make(map[string]*excelize.File),
	}

	// Load templates
	if err := provider.loadTemplates(); err != nil {
		return nil, fmt.Errorf("failed to load templates: %w", err)
	}

	return provider, nil
}

// GenerateFile generates an Excel file based on the request
func (p *Provider) GenerateFile(ctx context.Context, req *types.FileRequest) (*types.FileResponse, error) {
	file, err := p.createWorkbook(req)
	if err != nil {
		return &types.FileResponse{
			Success: false,
			Error:   err.Error(),
		}, err
	}

	// Generate file content
	content, err := p.generateContent(file)
	if err != nil {
		return &types.FileResponse{
			Success: false,
			Error:   err.Error(),
		}, err
	}

	return &types.FileResponse{
		Success:  true,
		Content:  content,
		MimeType: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
	}, nil
}

// GenerateFileToWriter generates an Excel file and writes it to the provided writer
func (p *Provider) GenerateFileToWriter(ctx context.Context, req *types.FileRequest, w io.Writer) error {
	file, err := p.createWorkbook(req)
	if err != nil {
		return err
	}

	// Write to buffer first, then to writer
	buf, err := p.generateContent(file)
	if err != nil {
		return err
	}

	_, err = w.Write(buf)
	return err
}

// GetSupportedTypes returns the file types supported by this provider
func (p *Provider) GetSupportedTypes() []types.FileType {
	return []types.FileType{types.FileTypeExcel}
}

// ValidateRequest validates the file generation request
func (p *Provider) ValidateRequest(req *types.FileRequest) error {
	if req.Type != types.FileTypeExcel {
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

// createWorkbook creates a new workbook based on the request
func (p *Provider) createWorkbook(req *types.FileRequest) (*excelize.File, error) {
	var file *excelize.File

	// Use template if specified
	if req.Template != "" {
		template, exists := p.templates[req.Template]
		if !exists {
			return nil, fmt.Errorf("template not found: %s", req.Template)
		}
		// Clone template - excelize doesn't have Clone method, so we'll create a new file
		// For now, we'll just use the template as-is
		file = template
	} else {
		// Create new workbook
		file = excelize.NewFile()
	}

	// Process data
	if err := p.processData(file, req.Data); err != nil {
		return nil, fmt.Errorf("failed to process data: %w", err)
	}

	return file, nil
}

// processData processes the data and adds it to the workbook
func (p *Provider) processData(file *excelize.File, data map[string]interface{}) error {
	// Process sheets data
	if sheets, ok := data["sheets"].(map[string]interface{}); ok {
		for sheetName, sheetData := range sheets {
			if err := p.processSheet(file, sheetName, sheetData); err != nil {
				return fmt.Errorf("failed to process sheet %s: %w", sheetName, err)
			}
		}
	} else {
		// Process single sheet data
		if err := p.processSheet(file, p.config.DefaultSheet, data); err != nil {
			return fmt.Errorf("failed to process default sheet: %w", err)
		}
	}

	return nil
}

// processSheet processes a single sheet
func (p *Provider) processSheet(file *excelize.File, sheetName string, sheetData interface{}) error {
	// Create or get sheet
	index, err := file.NewSheet(sheetName)
	if err != nil {
		// Sheet might already exist, try to get it
		index, err = file.GetSheetIndex(sheetName)
		if err != nil || index == -1 {
			return fmt.Errorf("failed to create or get sheet: %s", sheetName)
		}
	}

	// Set as active sheet if it's the first one
	if index == 1 {
		file.SetActiveSheet(index)
	}

	sheetDataMap, ok := sheetData.(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid sheet data format")
	}

	// Process headers
	if headers, ok := sheetDataMap["headers"].([]interface{}); ok {
		for i, header := range headers {
			cell := fmt.Sprintf("%c1", 'A'+i)
			file.SetCellValue(sheetName, cell, header)
		}
	}

	// Process rows
	if rows, ok := sheetDataMap["rows"].([]interface{}); ok {
		for rowIndex, row := range rows {
			rowData, ok := row.([]interface{})
			if !ok {
				continue
			}

			for colIndex, cellData := range rowData {
				cell := fmt.Sprintf("%c%d", 'A'+colIndex, rowIndex+2) // +2 because headers are in row 1
				file.SetCellValue(sheetName, cell, cellData)
			}
		}
	}

	// Process charts
	if charts, ok := sheetDataMap["charts"].([]interface{}); ok {
		for _, chartData := range charts {
			if err := p.addChart(file, sheetName, chartData); err != nil {
				return fmt.Errorf("failed to add chart: %w", err)
			}
		}
	}

	// Process formulas
	if formulas, ok := sheetDataMap["formulas"].(map[string]interface{}); ok {
		for cell, formula := range formulas {
			file.SetCellFormula(sheetName, cell, fmt.Sprintf("%v", formula))
		}
	}

	// Apply formatting
	if formatting, ok := sheetDataMap["formatting"].(map[string]interface{}); ok {
		if err := p.applyFormatting(file, sheetName, formatting); err != nil {
			return fmt.Errorf("failed to apply formatting: %w", err)
		}
	}

	return nil
}

// addChart adds a chart to the sheet
func (p *Provider) addChart(file *excelize.File, sheetName string, chartData interface{}) error {
	chartMap, ok := chartData.(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid chart data format")
	}

	chartType, _ := chartMap["type"].(string)
	if chartType == "" {
		chartType = "line"
	}

	// Get chart position
	x, _ := chartMap["x"].(float64)
	y, _ := chartMap["y"].(float64)
	width, _ := chartMap["width"].(float64)
	height, _ := chartMap["height"].(float64)

	if width == 0 {
		width = 480
	}
	if height == 0 {
		height = 300
	}

	// Get data range
	dataRange, _ := chartMap["data_range"].(string)
	if dataRange == "" {
		return fmt.Errorf("chart data_range is required")
	}

	// Create chart
	var chart excelize.Chart
	switch chartType {
	case "line":
		chart = excelize.Chart{
			Type: excelize.Line,
			Series: []excelize.ChartSeries{
				{
					Name:       "Series 1",
					Categories: dataRange,
					Values:     dataRange,
				},
			},
			Format: excelize.GraphicOptions{
				OffsetX: int(x),
				OffsetY: int(y),
				ScaleX:  1,
				ScaleY:  1,
			},
			Legend: excelize.ChartLegend{
				Position: "bottom",
			},
		}
	case "bar":
		chart = excelize.Chart{
			Type: excelize.Bar,
			Series: []excelize.ChartSeries{
				{
					Name:       "Series 1",
					Categories: dataRange,
					Values:     dataRange,
				},
			},
			Format: excelize.GraphicOptions{
				OffsetX: int(x),
				OffsetY: int(y),
				ScaleX:  1,
				ScaleY:  1,
			},
			Legend: excelize.ChartLegend{
				Position: "bottom",
			},
		}
	case "pie":
		chart = excelize.Chart{
			Type: excelize.Pie,
			Series: []excelize.ChartSeries{
				{
					Name:       "Series 1",
					Categories: dataRange,
					Values:     dataRange,
				},
			},
			Format: excelize.GraphicOptions{
				OffsetX: int(x),
				OffsetY: int(y),
				ScaleX:  1,
				ScaleY:  1,
			},
			Legend: excelize.ChartLegend{
				Position: "bottom",
			},
		}
	default:
		return fmt.Errorf("unsupported chart type: %s", chartType)
	}

	return file.AddChart(sheetName, "E2", &chart)
}

// applyFormatting applies formatting to the sheet
func (p *Provider) applyFormatting(file *excelize.File, sheetName string, formatting map[string]interface{}) error {
	// Apply column widths
	if columnWidths, ok := formatting["column_widths"].(map[string]interface{}); ok {
		for column, width := range columnWidths {
			if widthFloat, ok := width.(float64); ok {
				file.SetColWidth(sheetName, column, column, widthFloat)
			}
		}
	}

	// Apply row heights
	if rowHeights, ok := formatting["row_heights"].(map[string]interface{}); ok {
		for row, height := range rowHeights {
			if rowInt, err := strconv.Atoi(row); err == nil {
				if heightFloat, ok := height.(float64); ok {
					file.SetRowHeight(sheetName, rowInt, heightFloat)
				}
			}
		}
	}

	// Apply cell styles
	if cellStyles, ok := formatting["cell_styles"].(map[string]interface{}); ok {
		for cell, style := range cellStyles {
			styleMap, ok := style.(map[string]interface{})
			if !ok {
				continue
			}

			styleID, err := file.NewStyle(&excelize.Style{
				Font: &excelize.Font{
					Bold:   getBool(styleMap["bold"]),
					Italic: getBool(styleMap["italic"]),
					Size:   getFloat64(styleMap["size"]),
					Color:  getString(styleMap["color"]),
				},
				Fill: excelize.Fill{
					Type:    "pattern",
					Color:   []string{getString(styleMap["background_color"])},
					Pattern: 1,
				},
				Border: []excelize.Border{
					{Type: "left", Color: "000000", Style: 1},
					{Type: "top", Color: "000000", Style: 1},
					{Type: "bottom", Color: "000000", Style: 1},
					{Type: "right", Color: "000000", Style: 1},
				},
				Alignment: &excelize.Alignment{
					Horizontal: getString(styleMap["horizontal_align"]),
					Vertical:   getString(styleMap["vertical_align"]),
				},
			})
			if err != nil {
				continue
			}

			file.SetCellStyle(sheetName, cell, cell, styleID)
		}
	}

	return nil
}

// generateContent generates the workbook content as bytes
func (p *Provider) generateContent(file *excelize.File) ([]byte, error) {
	// Save to temporary file
	tmpFile, err := os.CreateTemp("", "excel_*.xlsx")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Save workbook
	if err := file.SaveAs(tmpFile.Name()); err != nil {
		return nil, fmt.Errorf("failed to save workbook: %w", err)
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
	files, err := filepath.Glob(filepath.Join(p.config.TemplatePath, "*.xlsx"))
	if err != nil {
		return fmt.Errorf("failed to read template directory: %w", err)
	}

	for _, file := range files {
		templateName := strings.TrimSuffix(filepath.Base(file), ".xlsx")

		// Load template
		template, err := excelize.OpenFile(file)
		if err != nil {
			continue // Skip invalid templates
		}

		p.templates[templateName] = template
	}

	return nil
}

// Close closes the provider and cleans up resources
func (p *Provider) Close() error {
	// Close all template files
	for _, file := range p.templates {
		file.Close()
	}
	p.templates = make(map[string]*excelize.File)
	return nil
}

// Helper functions
func getBool(value interface{}) bool {
	if b, ok := value.(bool); ok {
		return b
	}
	return false
}

func getFloat64(value interface{}) float64 {
	if f, ok := value.(float64); ok {
		return f
	}
	return 0
}

func getString(value interface{}) string {
	if s, ok := value.(string); ok {
		return s
	}
	return ""
}

// FileRequest, FileResponse, FileOptions, and FileType are now imported from the types package
