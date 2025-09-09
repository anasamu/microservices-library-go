package gateway

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	csvProvider "github.com/anasamu/microservices-library-go/filegen/providers/csv"
	customProvider "github.com/anasamu/microservices-library-go/filegen/providers/custom"
	docxProvider "github.com/anasamu/microservices-library-go/filegen/providers/docx"
	excelProvider "github.com/anasamu/microservices-library-go/filegen/providers/excel"
	pdfProvider "github.com/anasamu/microservices-library-go/filegen/providers/pdf"
)

// Manager manages file generation providers
type Manager struct {
	providers map[FileType]FileProvider
	mu        sync.RWMutex
	config    *ManagerConfig
}

// ManagerConfig contains configuration for the manager
type ManagerConfig struct {
	TemplatePath string
	OutputPath   string
	MaxFileSize  int64
	AllowedTypes []FileType
}

// NewManager creates a new file generation manager
func NewManager(config *ManagerConfig) (*Manager, error) {
	if config == nil {
		config = &ManagerConfig{
			TemplatePath: "./templates",
			OutputPath:   "./output",
			MaxFileSize:  100 * 1024 * 1024, // 100MB
			AllowedTypes: []FileType{FileTypeDOCX, FileTypeExcel, FileTypeCSV, FileTypePDF, FileTypeCustom},
		}
	}

	manager := &Manager{
		providers: make(map[FileType]FileProvider),
		config:    config,
	}

	// Initialize providers
	if err := manager.initializeProviders(); err != nil {
		return nil, fmt.Errorf("failed to initialize providers: %w", err)
	}

	return manager, nil
}

// initializeProviders initializes all available providers
func (m *Manager) initializeProviders() error {
	// Initialize DOCX provider
	docx, err := docxProvider.NewProvider(&docxProvider.Config{
		TemplatePath: filepath.Join(m.config.TemplatePath, "docx"),
	})
	if err != nil {
		return fmt.Errorf("failed to initialize DOCX provider: %w", err)
	}
	m.providers[FileTypeDOCX] = docx

	// Initialize Excel provider
	excel, err := excelProvider.NewProvider(&excelProvider.Config{
		TemplatePath: filepath.Join(m.config.TemplatePath, "excel"),
	})
	if err != nil {
		return fmt.Errorf("failed to initialize Excel provider: %w", err)
	}
	m.providers[FileTypeExcel] = excel

	// Initialize CSV provider
	csv, err := csvProvider.NewProvider(&csvProvider.Config{
		TemplatePath: filepath.Join(m.config.TemplatePath, "csv"),
	})
	if err != nil {
		return fmt.Errorf("failed to initialize CSV provider: %w", err)
	}
	m.providers[FileTypeCSV] = csv

	// Initialize PDF provider
	pdf, err := pdfProvider.NewProvider(&pdfProvider.Config{
		TemplatePath: filepath.Join(m.config.TemplatePath, "pdf"),
	})
	if err != nil {
		return fmt.Errorf("failed to initialize PDF provider: %w", err)
	}
	m.providers[FileTypePDF] = pdf

	// Initialize Custom provider
	custom, err := customProvider.NewProvider(&customProvider.Config{
		TemplatePath: filepath.Join(m.config.TemplatePath, "custom"),
	})
	if err != nil {
		return fmt.Errorf("failed to initialize Custom provider: %w", err)
	}
	m.providers[FileTypeCustom] = custom

	return nil
}

// GenerateFile generates a file using the appropriate provider
func (m *Manager) GenerateFile(ctx context.Context, req *FileRequest) (*FileResponse, error) {
	if err := m.validateRequest(req); err != nil {
		return &FileResponse{
			Success: false,
			Error:   err.Error(),
		}, err
	}

	m.mu.RLock()
	provider, exists := m.providers[req.Type]
	m.mu.RUnlock()

	if !exists {
		return &FileResponse{
			Success: false,
			Error:   fmt.Sprintf("unsupported file type: %s", req.Type),
		}, fmt.Errorf("unsupported file type: %s", req.Type)
	}

	// Validate request with provider
	if err := provider.ValidateRequest(req); err != nil {
		return &FileResponse{
			Success: false,
			Error:   err.Error(),
		}, err
	}

	// Generate file
	response, err := provider.GenerateFile(ctx, req)
	if err != nil {
		return &FileResponse{
			Success: false,
			Error:   err.Error(),
		}, err
	}

	// Save file if output path is specified
	if req.OutputPath != "" {
		if err := m.saveFile(response, req.OutputPath); err != nil {
			return &FileResponse{
				Success: false,
				Error:   fmt.Sprintf("failed to save file: %v", err),
			}, err
		}
		response.FilePath = req.OutputPath
	}

	return response, nil
}

// GenerateFileToWriter generates a file and writes it to the provided writer
func (m *Manager) GenerateFileToWriter(ctx context.Context, req *FileRequest, w io.Writer) error {
	if err := m.validateRequest(req); err != nil {
		return err
	}

	m.mu.RLock()
	provider, exists := m.providers[req.Type]
	m.mu.RUnlock()

	if !exists {
		return fmt.Errorf("unsupported file type: %s", req.Type)
	}

	if err := provider.ValidateRequest(req); err != nil {
		return err
	}

	return provider.GenerateFileToWriter(ctx, req, w)
}

// GetSupportedTypes returns all supported file types
func (m *Manager) GetSupportedTypes() []FileType {
	m.mu.RLock()
	defer m.mu.RUnlock()

	types := make([]FileType, 0, len(m.providers))
	for fileType := range m.providers {
		types = append(types, fileType)
	}
	return types
}

// GetProvider returns a specific provider by type
func (m *Manager) GetProvider(fileType FileType) (FileProvider, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	provider, exists := m.providers[fileType]
	if !exists {
		return nil, fmt.Errorf("provider not found for type: %s", fileType)
	}
	return provider, nil
}

// GetTemplateList returns available templates for a specific file type
func (m *Manager) GetTemplateList(fileType FileType) ([]string, error) {
	provider, err := m.GetProvider(fileType)
	if err != nil {
		return nil, err
	}
	return provider.GetTemplateList()
}

// GetTemplateInfo returns detailed information about a template
func (m *Manager) GetTemplateInfo(fileType FileType, templateName string) (*TemplateInfo, error) {
	provider, err := m.GetProvider(fileType)
	if err != nil {
		return nil, err
	}

	// This would need to be implemented in each provider
	// For now, return basic info
	return &TemplateInfo{
		Name:        templateName,
		Description: fmt.Sprintf("Template for %s files", fileType),
		Type:        fileType,
		Parameters:  []TemplateParam{},
	}, nil
}

// validateRequest validates the file generation request
func (m *Manager) validateRequest(req *FileRequest) error {
	if req == nil {
		return &FileGenerationError{
			Type:    ErrorTypeValidation,
			Message: "request cannot be nil",
			Code:    400,
		}
	}

	if req.Type == "" {
		return &FileGenerationError{
			Type:    ErrorTypeValidation,
			Message: "file type is required",
			Code:    400,
		}
	}

	// Check if file type is allowed
	allowed := false
	for _, allowedType := range m.config.AllowedTypes {
		if req.Type == allowedType {
			allowed = true
			break
		}
	}

	if !allowed {
		return &FileGenerationError{
			Type:    ErrorTypeValidation,
			Message: fmt.Sprintf("file type %s is not allowed", req.Type),
			Code:    403,
		}
	}

	if req.Data == nil {
		return &FileGenerationError{
			Type:    ErrorTypeValidation,
			Message: "data is required",
			Code:    400,
		}
	}

	return nil
}

// saveFile saves the generated file content to the specified path
func (m *Manager) saveFile(response *FileResponse, outputPath string) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Write file
	if err := os.WriteFile(outputPath, response.Content, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	// Get file size
	info, err := os.Stat(outputPath)
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}

	response.FileSize = info.Size()
	return nil
}

// GetMimeType returns the MIME type for a given file type
func (m *Manager) GetMimeType(fileType FileType) string {
	switch fileType {
	case FileTypeDOCX:
		return "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	case FileTypeExcel:
		return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	case FileTypeCSV:
		return "text/csv"
	case FileTypePDF:
		return "application/pdf"
	default:
		return "application/octet-stream"
	}
}

// GetFileExtension returns the file extension for a given file type
func (m *Manager) GetFileExtension(fileType FileType) string {
	switch fileType {
	case FileTypeDOCX:
		return ".docx"
	case FileTypeExcel:
		return ".xlsx"
	case FileTypeCSV:
		return ".csv"
	case FileTypePDF:
		return ".pdf"
	default:
		return ""
	}
}

// Close closes all providers and cleans up resources
func (m *Manager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var errors []string
	for fileType, provider := range m.providers {
		if closer, ok := provider.(interface{ Close() error }); ok {
			if err := closer.Close(); err != nil {
				errors = append(errors, fmt.Sprintf("%s: %v", fileType, err))
			}
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("errors closing providers: %s", strings.Join(errors, "; "))
	}

	return nil
}
