package mocks

import (
	"context"
	"fmt"
	"io"
)

// MockFileProvider is a mock implementation of the FileProvider interface
type MockFileProvider struct {
	GenerateFileFunc         func(ctx context.Context, req *FileRequest) (*FileResponse, error)
	GenerateFileToWriterFunc func(ctx context.Context, req *FileRequest, w io.Writer) error
	GetSupportedTypesFunc    func() []FileType
	ValidateRequestFunc      func(req *FileRequest) error
	GetTemplateListFunc      func() ([]string, error)
	CloseFunc                func() error
}

// GenerateFile calls the mock function
func (m *MockFileProvider) GenerateFile(ctx context.Context, req *FileRequest) (*FileResponse, error) {
	if m.GenerateFileFunc != nil {
		return m.GenerateFileFunc(ctx, req)
	}
	return &FileResponse{Success: true, Content: []byte("mock content")}, nil
}

// GenerateFileToWriter calls the mock function
func (m *MockFileProvider) GenerateFileToWriter(ctx context.Context, req *FileRequest, w io.Writer) error {
	if m.GenerateFileToWriterFunc != nil {
		return m.GenerateFileToWriterFunc(ctx, req, w)
	}
	_, err := w.Write([]byte("mock content"))
	return err
}

// GetSupportedTypes calls the mock function
func (m *MockFileProvider) GetSupportedTypes() []FileType {
	if m.GetSupportedTypesFunc != nil {
		return m.GetSupportedTypesFunc()
	}
	return []FileType{"mock"}
}

// ValidateRequest calls the mock function
func (m *MockFileProvider) ValidateRequest(req *FileRequest) error {
	if m.ValidateRequestFunc != nil {
		return m.ValidateRequestFunc(req)
	}
	return nil
}

// GetTemplateList calls the mock function
func (m *MockFileProvider) GetTemplateList() ([]string, error) {
	if m.GetTemplateListFunc != nil {
		return m.GetTemplateListFunc()
	}
	return []string{"template1", "template2"}, nil
}

// Close calls the mock function
func (m *MockFileProvider) Close() error {
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	return nil
}

// FileRequest represents a request to generate a file
type FileRequest struct {
	Type       string                 `json:"type"`
	Template   string                 `json:"template,omitempty"`
	Data       map[string]interface{} `json:"data"`
	Options    FileOptions            `json:"options,omitempty"`
	OutputPath string                 `json:"output_path,omitempty"`
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

// FileOptions contains additional options for file generation
type FileOptions struct {
	Format    string                 `json:"format,omitempty"`
	Encoding  string                 `json:"encoding,omitempty"`
	Delimiter string                 `json:"delimiter,omitempty"`
	Headers   []string               `json:"headers,omitempty"`
	Metadata  map[string]string      `json:"metadata,omitempty"`
	Custom    map[string]interface{} `json:"custom,omitempty"`
}

// FileType represents the type of file to generate
type FileType string

const (
	FileTypeDOCX   FileType = "docx"
	FileTypeExcel  FileType = "excel"
	FileTypeCSV    FileType = "csv"
	FileTypePDF    FileType = "pdf"
	FileTypeCustom FileType = "custom"
)

// MockFileProviderBuilder helps build mock providers with specific behaviors
type MockFileProviderBuilder struct {
	provider *MockFileProvider
}

// NewMockFileProviderBuilder creates a new mock provider builder
func NewMockFileProviderBuilder() *MockFileProviderBuilder {
	return &MockFileProviderBuilder{
		provider: &MockFileProvider{},
	}
}

// WithGenerateFile sets the GenerateFile mock function
func (b *MockFileProviderBuilder) WithGenerateFile(fn func(ctx context.Context, req *FileRequest) (*FileResponse, error)) *MockFileProviderBuilder {
	b.provider.GenerateFileFunc = fn
	return b
}

// WithGenerateFileToWriter sets the GenerateFileToWriter mock function
func (b *MockFileProviderBuilder) WithGenerateFileToWriter(fn func(ctx context.Context, req *FileRequest, w io.Writer) error) *MockFileProviderBuilder {
	b.provider.GenerateFileToWriterFunc = fn
	return b
}

// WithGetSupportedTypes sets the GetSupportedTypes mock function
func (b *MockFileProviderBuilder) WithGetSupportedTypes(fn func() []FileType) *MockFileProviderBuilder {
	b.provider.GetSupportedTypesFunc = fn
	return b
}

// WithValidateRequest sets the ValidateRequest mock function
func (b *MockFileProviderBuilder) WithValidateRequest(fn func(req *FileRequest) error) *MockFileProviderBuilder {
	b.provider.ValidateRequestFunc = fn
	return b
}

// WithGetTemplateList sets the GetTemplateList mock function
func (b *MockFileProviderBuilder) WithGetTemplateList(fn func() ([]string, error)) *MockFileProviderBuilder {
	b.provider.GetTemplateListFunc = fn
	return b
}

// WithClose sets the Close mock function
func (b *MockFileProviderBuilder) WithClose(fn func() error) *MockFileProviderBuilder {
	b.provider.CloseFunc = fn
	return b
}

// Build returns the configured mock provider
func (b *MockFileProviderBuilder) Build() *MockFileProvider {
	return b.provider
}

// Common mock behaviors

// SuccessBehavior returns a mock provider that always succeeds
func SuccessBehavior() *MockFileProvider {
	return NewMockFileProviderBuilder().
		WithGenerateFile(func(ctx context.Context, req *FileRequest) (*FileResponse, error) {
			return &FileResponse{
				Success:  true,
				Content:  []byte("success content"),
				MimeType: "text/plain",
			}, nil
		}).
		WithGenerateFileToWriter(func(ctx context.Context, req *FileRequest, w io.Writer) error {
			_, err := w.Write([]byte("success content"))
			return err
		}).
		WithGetSupportedTypes(func() []FileType {
			return []FileType{FileTypeCSV}
		}).
		WithValidateRequest(func(req *FileRequest) error {
			return nil
		}).
		WithGetTemplateList(func() ([]string, error) {
			return []string{"template1"}, nil
		}).
		Build()
}

// ErrorBehavior returns a mock provider that always returns an error
func ErrorBehavior() *MockFileProvider {
	return NewMockFileProviderBuilder().
		WithGenerateFile(func(ctx context.Context, req *FileRequest) (*FileResponse, error) {
			return &FileResponse{
				Success: false,
				Error:   "mock error",
			}, nil
		}).
		WithGenerateFileToWriter(func(ctx context.Context, req *FileRequest, w io.Writer) error {
			return fmt.Errorf("mock error")
		}).
		WithValidateRequest(func(req *FileRequest) error {
			return fmt.Errorf("validation error")
		}).
		Build()
}

// ValidationErrorBehavior returns a mock provider that fails validation
func ValidationErrorBehavior() *MockFileProvider {
	return NewMockFileProviderBuilder().
		WithValidateRequest(func(req *FileRequest) error {
			if req.Type == "invalid" {
				return fmt.Errorf("invalid file type")
			}
			if req.Data == nil {
				return fmt.Errorf("data is required")
			}
			return nil
		}).
		Build()
}

// TemplateBehavior returns a mock provider that handles templates
func TemplateBehavior() *MockFileProvider {
	return NewMockFileProviderBuilder().
		WithGenerateFile(func(ctx context.Context, req *FileRequest) (*FileResponse, error) {
			content := "default content"
			if req.Template != "" {
				content = "template content: " + req.Template
			}
			return &FileResponse{
				Success:  true,
				Content:  []byte(content),
				MimeType: "text/plain",
			}, nil
		}).
		WithGetTemplateList(func() ([]string, error) {
			return []string{"template1", "template2", "template3"}, nil
		}).
		Build()
}
