package gateway

import (
	"context"
	"io"
)

// FileType represents the type of file to generate
type FileType string

const (
	FileTypeDOCX   FileType = "docx"
	FileTypeExcel  FileType = "excel"
	FileTypeCSV    FileType = "csv"
	FileTypePDF    FileType = "pdf"
	FileTypeCustom FileType = "custom"
)

// FileRequest represents a request to generate a file
type FileRequest struct {
	Type       FileType               `json:"type"`
	Template   string                 `json:"template,omitempty"`
	Data       map[string]interface{} `json:"data"`
	Options    FileOptions            `json:"options,omitempty"`
	OutputPath string                 `json:"output_path,omitempty"`
}

// FileOptions contains additional options for file generation
type FileOptions struct {
	Format    string                 `json:"format,omitempty"`
	Encoding  string                 `json:"encoding,omitempty"`
	Delimiter string                 `json:"delimiter,omitempty"`
	Headers   []string               `json:"headers,omitempty"`
	Metadata  map[string]string      `json:"metadata,omitempty"`
	Custom    map[string]interface{} `json:"custom,omitempty"`
}

// FileResponse represents the response from file generation
type FileResponse struct {
	Success  bool   `json:"success"`
	FilePath string `json:"file_path,omitempty"`
	FileSize int64  `json:"file_size,omitempty"`
	Error    string `json:"error,omitempty"`
	Content  []byte `json:"content,omitempty"`
	MimeType string `json:"mime_type,omitempty"`
}

// FileProvider defines the interface for file generation providers
type FileProvider interface {
	// GenerateFile generates a file based on the request
	GenerateFile(ctx context.Context, req *FileRequest) (*FileResponse, error)

	// GenerateFileToWriter generates a file and writes it to the provided writer
	GenerateFileToWriter(ctx context.Context, req *FileRequest, w io.Writer) error

	// GetSupportedTypes returns the file types supported by this provider
	GetSupportedTypes() []FileType

	// ValidateRequest validates the file generation request
	ValidateRequest(req *FileRequest) error

	// GetTemplateList returns available templates for this provider
	GetTemplateList() ([]string, error)
}

// TemplateInfo represents information about a template
type TemplateInfo struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Type        FileType        `json:"type"`
	Parameters  []TemplateParam `json:"parameters"`
	Preview     string          `json:"preview,omitempty"`
}

// TemplateParam represents a template parameter
type TemplateParam struct {
	Name        string      `json:"name"`
	Type        string      `json:"type"`
	Required    bool        `json:"required"`
	Description string      `json:"description"`
	Default     interface{} `json:"default,omitempty"`
	Options     []string    `json:"options,omitempty"`
}

// FileGenerationError represents a file generation error
type FileGenerationError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
	Code    int    `json:"code"`
}

func (e *FileGenerationError) Error() string {
	return e.Message
}

// Common error types
const (
	ErrorTypeValidation = "validation_error"
	ErrorTypeTemplate   = "template_error"
	ErrorTypeGeneration = "generation_error"
	ErrorTypeIO         = "io_error"
	ErrorTypeProvider   = "provider_error"
)
