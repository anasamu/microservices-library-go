# File Generation Library

A comprehensive Go library for generating various file formats including DOCX, Excel, CSV, PDF, and custom formats. This library provides a unified interface for file generation with support for templates, custom formatting, and multiple output options.

## Features

- **Multiple File Formats**: Support for DOCX, Excel, CSV, PDF, and custom formats
- **Template Support**: Use pre-built templates for consistent document generation
- **Flexible Data Input**: Support for various data structures and formats
- **Custom Formatting**: Extensive formatting options for each file type
- **Streaming Support**: Generate files directly to writers for memory efficiency
- **Validation**: Built-in request validation and error handling
- **Extensible**: Easy to add new file format providers

## Installation

```bash
go get github.com/anasamu/microservices-library-go/filegen
```

## Quick Start

```go
package main

import (
    "context"
    "log"
    
    "github.com/anasamu/microservices-library-go/filegen"
)

func main() {
    // Create manager
    manager, err := gateway.NewManager(nil)
    if err != nil {
        log.Fatal(err)
    }
    defer manager.Close()

    ctx := context.Background()

    // Generate CSV file
    req := &gateway.FileRequest{
        Type: gateway.FileTypeCSV,
        Data: map[string]interface{}{
            "headers": []string{"Name", "Age", "City"},
            "rows": [][]string{
                {"John Doe", "30", "New York"},
                {"Jane Smith", "25", "Los Angeles"},
            },
        },
        OutputPath: "./output.csv",
    }

    response, err := manager.GenerateFile(ctx, req)
    if err != nil {
        log.Fatal(err)
    }

    if response.Success {
        log.Printf("File generated successfully: %s", response.FilePath)
    }
}
```

## Supported File Types

### DOCX (Microsoft Word)
Generate professional Word documents with:
- Headings and paragraphs
- Tables and lists
- Images and formatting
- Template support

```go
req := &gateway.FileRequest{
    Type: gateway.FileTypeDOCX,
    Data: map[string]interface{}{
        "title": "Monthly Report",
        "content": []map[string]interface{}{
            {
                "type": "heading",
                "text": "Executive Summary",
                "level": 1,
            },
            {
                "type": "paragraph",
                "text": "This report covers monthly performance metrics.",
            },
            {
                "type": "table",
                "headers": []string{"Metric", "Value", "Change"},
                "rows": [][]string{
                    {"Revenue", "$1,000,000", "+5%"},
                    {"Customers", "1,250", "+10%"},
                },
            },
        },
    },
}
```

### Excel (Microsoft Excel)
Create Excel workbooks with:
- Multiple sheets
- Charts and graphs
- Formulas and formatting
- Data validation

```go
req := &gateway.FileRequest{
    Type: gateway.FileTypeExcel,
    Data: map[string]interface{}{
        "sheets": map[string]interface{}{
            "Employees": map[string]interface{}{
                "headers": []string{"ID", "Name", "Department", "Salary"},
                "rows": [][]interface{}{
                    {1, "John Doe", "Engineering", 75000},
                    {2, "Jane Smith", "Marketing", 65000},
                },
            },
            "Summary": map[string]interface{}{
                "headers": []string{"Department", "Count", "Avg Salary"},
                "rows": [][]interface{}{
                    {"Engineering", 1, 75000},
                    {"Marketing", 1, 65000},
                },
            },
        },
    },
}
```

### CSV (Comma-Separated Values)
Generate CSV files with:
- Custom delimiters
- Headers and data rows
- Multiple data formats
- Template support

```go
req := &gateway.FileRequest{
    Type: gateway.FileTypeCSV,
    Data: map[string]interface{}{
        "headers": []string{"Product", "Price", "Stock"},
        "rows": [][]string{
            {"Laptop", "$999", "50"},
            {"Mouse", "$25", "200"},
        },
    },
    Options: gateway.FileOptions{
        Delimiter: ",",
    },
}
```

### PDF (Portable Document Format)
Create PDF documents with:
- Rich text formatting
- Tables and lists
- Images and metadata
- Page breaks and styling

```go
req := &gateway.FileRequest{
    Type: gateway.FileTypePDF,
    Data: map[string]interface{}{
        "title": "Invoice #INV-001",
        "metadata": map[string]interface{}{
            "author": "Your Company",
            "subject": "Invoice",
        },
        "content": []map[string]interface{}{
            {
                "type": "heading",
                "text": "Invoice Details",
                "level": 1,
            },
            {
                "type": "table",
                "headers": []string{"Description", "Quantity", "Price", "Total"},
                "rows": [][]string{
                    {"Web Development", "40 hours", "$100/hour", "$4,000"},
                },
            },
        },
    },
}
```

### Custom Formats
Generate custom file formats including:
- JSON
- XML
- YAML
- HTML
- Markdown
- Plain text

```go
req := &gateway.FileRequest{
    Type: gateway.FileTypeCustom,
    Data: map[string]interface{}{
        "title": "Custom Document",
        "content": "This is a custom document",
        "metadata": map[string]interface{}{
            "version": "1.0",
            "author": "System",
        },
    },
    Options: gateway.FileOptions{
        Format: "json",
    },
}
```

## Template Support

### Using Templates

```go
req := &gateway.FileRequest{
    Type:     gateway.FileTypeDOCX,
    Template: "report", // Template name
    Data: map[string]interface{}{
        "company_name": "Acme Corp",
        "report_date":  "2024-01-15",
        "metrics": map[string]interface{}{
            "revenue":    1000000,
            "customers":  1250,
            "growth":     "15%",
        },
    },
}
```

### Template Structure

Templates are stored in the `templates/` directory:

```
templates/
├── docx/
│   ├── report.docx
│   └── invoice.docx
├── excel/
│   ├── report.xlsx
│   └── data.xlsx
├── pdf/
│   └── report.tmpl
└── custom/
    └── invoice.tmpl
```

## Configuration

### Manager Configuration

```go
config := &gateway.ManagerConfig{
    TemplatePath: "./templates",
    OutputPath:   "./output",
    MaxFileSize:  100 * 1024 * 1024, // 100MB
    AllowedTypes: []gateway.FileType{
        gateway.FileTypeDOCX,
        gateway.FileTypeExcel,
        gateway.FileTypeCSV,
        gateway.FileTypePDF,
    },
}

manager, err := gateway.NewManager(config)
```

### Provider Configuration

Each provider can be configured independently:

```go
// DOCX Provider
docxConfig := &docx.Config{
    TemplatePath: "./templates/docx",
    DefaultFont:  "Calibri",
    DefaultSize:  11,
}

// Excel Provider
excelConfig := &excel.Config{
    TemplatePath: "./templates/excel",
    DefaultSheet: "Sheet1",
}

// CSV Provider
csvConfig := &csv.Config{
    TemplatePath:    "./templates/csv",
    DefaultDelimiter: ',',
    DefaultEncoding:  "utf-8",
}
```

## Advanced Usage

### Streaming to Writer

For large files or memory efficiency:

```go
var buf bytes.Buffer
err := manager.GenerateFileToWriter(ctx, req, &buf)
if err != nil {
    log.Fatal(err)
}

// Use buf.Bytes() or buf.String()
```

### Batch Generation

```go
requests := []*gateway.FileRequest{
    {
        Type: gateway.FileTypeCSV,
        Data: map[string]interface{}{
            "headers": []string{"Product", "Price"},
            "rows": [][]string{{"Laptop", "$999"}},
        },
        OutputPath: "./products.csv",
    },
    {
        Type: gateway.FileTypeExcel,
        Data: map[string]interface{}{
            "headers": []string{"Product", "Price"},
            "rows": [][]interface{}{{"Laptop", 999}},
        },
        OutputPath: "./products.xlsx",
    },
}

for i, req := range requests {
    response, err := manager.GenerateFile(ctx, req)
    if err != nil {
        log.Printf("Request %d failed: %v", i+1, err)
    } else {
        log.Printf("Request %d completed: %s", i+1, response.FilePath)
    }
}
```

### Error Handling

```go
response, err := manager.GenerateFile(ctx, req)
if err != nil {
    if fileErr, ok := err.(*gateway.FileGenerationError); ok {
        switch fileErr.Type {
        case gateway.ErrorTypeValidation:
            log.Printf("Validation error: %s", fileErr.Message)
        case gateway.ErrorTypeTemplate:
            log.Printf("Template error: %s", fileErr.Message)
        case gateway.ErrorTypeGeneration:
            log.Printf("Generation error: %s", fileErr.Message)
        default:
            log.Printf("Unknown error: %s", fileErr.Message)
        }
    } else {
        log.Printf("Unexpected error: %v", err)
    }
    return
}

if !response.Success {
    log.Printf("Generation failed: %s", response.Error)
    return
}
```

## API Reference

### Manager

#### `NewManager(config *ManagerConfig) (*Manager, error)`
Creates a new file generation manager.

#### `GenerateFile(ctx context.Context, req *FileRequest) (*FileResponse, error)`
Generates a file based on the request.

#### `GenerateFileToWriter(ctx context.Context, req *FileRequest, w io.Writer) error`
Generates a file and writes it to the provided writer.

#### `GetSupportedTypes() []FileType`
Returns all supported file types.

#### `GetProvider(fileType FileType) (FileProvider, error)`
Returns a specific provider by type.

#### `GetTemplateList(fileType FileType) ([]string, error)`
Returns available templates for a specific file type.

### FileRequest

```go
type FileRequest struct {
    Type       FileType              `json:"type"`
    Template   string                `json:"template,omitempty"`
    Data       map[string]interface{} `json:"data"`
    Options    FileOptions           `json:"options,omitempty"`
    OutputPath string                `json:"output_path,omitempty"`
}
```

### FileResponse

```go
type FileResponse struct {
    Success    bool   `json:"success"`
    FilePath   string `json:"file_path,omitempty"`
    FileSize   int64  `json:"file_size,omitempty"`
    Error      string `json:"error,omitempty"`
    Content    []byte `json:"content,omitempty"`
    MimeType   string `json:"mime_type,omitempty"`
}
```

### FileOptions

```go
type FileOptions struct {
    Format     string                 `json:"format,omitempty"`
    Encoding   string                 `json:"encoding,omitempty"`
    Delimiter  string                 `json:"delimiter,omitempty"`
    Headers    []string               `json:"headers,omitempty"`
    Metadata   map[string]string      `json:"metadata,omitempty"`
    Custom     map[string]interface{} `json:"custom,omitempty"`
}
```

## Testing

### Running Tests

```bash
# Run all tests
go test ./...

# Run integration tests
go test ./test/integration/...

# Run unit tests
go test ./test/unit/...

# Run with coverage
go test -cover ./...
```

### Test Structure

```
test/
├── integration/          # Integration tests
│   ├── docx_test.go
│   ├── excel_test.go
│   ├── csv_test.go
│   └── pdf_test.go
└── unit/                 # Unit tests
    ├── manager_test.go
    └── mocks/
        └── mock_provider.go
```

## Examples

See the `examples/` directory for complete working examples:

- [Basic Usage](examples/basic/main.go)
- [Template Usage](examples/template/main.go)
- [Batch Generation](examples/batch/main.go)
- [Custom Provider](examples/custom/main.go)

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Ensure all tests pass
6. Submit a pull request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Dependencies

- [unidoc/unioffice](https://github.com/unidoc/unioffice) - For DOCX and PDF generation
- [xuri/excelize](https://github.com/xuri/excelize) - For Excel file generation
- Go standard library - For CSV generation

## Performance Considerations

- Use `GenerateFileToWriter` for large files to avoid loading entire content into memory
- Consider using templates for frequently generated documents
- Batch multiple requests when possible
- Monitor memory usage for large document generation

## Troubleshooting

### Common Issues

1. **Template not found**: Ensure template files exist in the correct directory
2. **Memory issues**: Use streaming generation for large files
3. **Permission errors**: Check file system permissions for output directories
4. **Invalid data format**: Validate data structure before generation

### Debug Mode

Enable debug logging by setting the log level:

```go
import "log"

log.SetFlags(log.LstdFlags | log.Lshortfile)
```

## Roadmap

- [ ] Add support for PowerPoint (PPTX) files
- [ ] Implement image generation capabilities
- [ ] Add more chart types for Excel
- [ ] Support for encrypted PDFs
- [ ] Web interface for template management
- [ ] Performance optimizations
- [ ] Additional custom format support
