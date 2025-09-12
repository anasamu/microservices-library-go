package unit

import (
	"context"
	"testing"

	"github.com/anasamu/microservices-library-go/filegen"
)

func TestManager(t *testing.T) {
	config := &gateway.ManagerConfig{
		TemplatePath: "./templates",
		OutputPath:   "./output",
		MaxFileSize:  100 * 1024 * 1024, // 100MB
		AllowedTypes: []gateway.FileType{gateway.FileTypeDOCX, gateway.FileTypeExcel, gateway.FileTypeCSV, gateway.FileTypePDF},
	}

	manager, err := gateway.NewManager(config)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}
	defer manager.Close()

	ctx := context.Background()

	t.Run("GetSupportedTypes", func(t *testing.T) {
		types := manager.GetSupportedTypes()
		expectedTypes := []gateway.FileType{
			gateway.FileTypeDOCX,
			gateway.FileTypeExcel,
			gateway.FileTypeCSV,
			gateway.FileTypePDF,
			gateway.FileTypeCustom,
		}

		if len(types) != len(expectedTypes) {
			t.Errorf("Expected %d supported types, got %d", len(expectedTypes), len(types))
		}

		for _, expectedType := range expectedTypes {
			found := false
			for _, actualType := range types {
				if actualType == expectedType {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected type %s not found in supported types", expectedType)
			}
		}
	})

	t.Run("GetProvider", func(t *testing.T) {
		// Test valid provider
		provider, err := manager.GetProvider(gateway.FileTypeCSV)
		if err != nil {
			t.Errorf("Failed to get CSV provider: %v", err)
		}
		if provider == nil {
			t.Error("Provider should not be nil")
		}

		// Test invalid provider
		_, err = manager.GetProvider("invalid_type")
		if err == nil {
			t.Error("Expected error for invalid provider type")
		}
	})

	t.Run("GetTemplateList", func(t *testing.T) {
		templates, err := manager.GetTemplateList(gateway.FileTypeCSV)
		if err != nil {
			t.Errorf("Failed to get template list: %v", err)
		}
		if templates == nil {
			t.Error("Template list should not be nil")
		}
	})

	t.Run("GetTemplateInfo", func(t *testing.T) {
		info, err := manager.GetTemplateInfo(gateway.FileTypeCSV, "test_template")
		if err != nil {
			t.Errorf("Failed to get template info: %v", err)
		}
		if info == nil {
			t.Error("Template info should not be nil")
		}
		if info.Name != "test_template" {
			t.Errorf("Expected template name 'test_template', got '%s'", info.Name)
		}
	})

	t.Run("GetMimeType", func(t *testing.T) {
		testCases := []struct {
			fileType     gateway.FileType
			expectedMime string
		}{
			{gateway.FileTypeDOCX, "application/vnd.openxmlformats-officedocument.wordprocessingml.document"},
			{gateway.FileTypeExcel, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"},
			{gateway.FileTypeCSV, "text/csv"},
			{gateway.FileTypePDF, "application/pdf"},
			{gateway.FileTypeCustom, "application/octet-stream"},
		}

		for _, tc := range testCases {
			mimeType := manager.GetMimeType(tc.fileType)
			if mimeType != tc.expectedMime {
				t.Errorf("Expected MIME type '%s' for %s, got '%s'", tc.expectedMime, tc.fileType, mimeType)
			}
		}
	})

	t.Run("GetFileExtension", func(t *testing.T) {
		testCases := []struct {
			fileType    gateway.FileType
			expectedExt string
		}{
			{gateway.FileTypeDOCX, ".docx"},
			{gateway.FileTypeExcel, ".xlsx"},
			{gateway.FileTypeCSV, ".csv"},
			{gateway.FileTypePDF, ".pdf"},
			{gateway.FileTypeCustom, ""},
		}

		for _, tc := range testCases {
			ext := manager.GetFileExtension(tc.fileType)
			if ext != tc.expectedExt {
				t.Errorf("Expected extension '%s' for %s, got '%s'", tc.expectedExt, tc.fileType, ext)
			}
		}
	})

	t.Run("ValidateRequest", func(t *testing.T) {
		// Test valid request
		req := &gateway.FileRequest{
			Type: gateway.FileTypeCSV,
			Data: map[string]interface{}{"test": "data"},
		}

		// This should not cause a panic or error
		_, err := manager.GenerateFile(ctx, req)
		// We expect an error because we don't have actual providers set up in unit tests
		if err == nil {
			t.Error("Expected error for unit test without proper provider setup")
		}

		// Test invalid request
		invalidReq := &gateway.FileRequest{
			Type: "invalid_type",
			Data: map[string]interface{}{"test": "data"},
		}

		_, err = manager.GenerateFile(ctx, invalidReq)
		if err == nil {
			t.Error("Expected error for invalid file type")
		}
	})
}

func TestManagerConfig(t *testing.T) {
	t.Run("DefaultConfig", func(t *testing.T) {
		manager, err := gateway.NewManager(nil)
		if err != nil {
			t.Fatalf("Failed to create manager with default config: %v", err)
		}
		defer manager.Close()

		// Should not panic and should have default values
		types := manager.GetSupportedTypes()
		if len(types) == 0 {
			t.Error("Manager should have supported types")
		}
	})

	t.Run("CustomConfig", func(t *testing.T) {
		config := &gateway.ManagerConfig{
			TemplatePath: "/custom/templates",
			OutputPath:   "/custom/output",
			MaxFileSize:  50 * 1024 * 1024, // 50MB
			AllowedTypes: []gateway.FileType{gateway.FileTypeCSV},
		}

		manager, err := gateway.NewManager(config)
		if err != nil {
			t.Fatalf("Failed to create manager with custom config: %v", err)
		}
		defer manager.Close()

		// Should have the custom configuration
		types := manager.GetSupportedTypes()
		if len(types) == 0 {
			t.Error("Manager should have supported types")
		}
	})
}

func TestFileGenerationError(t *testing.T) {
	t.Run("ErrorString", func(t *testing.T) {
		err := &gateway.FileGenerationError{
			Type:    gateway.ErrorTypeValidation,
			Message: "Test error message",
			Code:    400,
		}

		expected := "Test error message"
		if err.Error() != expected {
			t.Errorf("Expected error message '%s', got '%s'", expected, err.Error())
		}
	})

	t.Run("ErrorTypes", func(t *testing.T) {
		errorTypes := []string{
			gateway.ErrorTypeValidation,
			gateway.ErrorTypeTemplate,
			gateway.ErrorTypeGeneration,
			gateway.ErrorTypeIO,
			gateway.ErrorTypeProvider,
		}

		for _, errorType := range errorTypes {
			err := &gateway.FileGenerationError{
				Type:    errorType,
				Message: "Test message",
				Code:    500,
			}

			if err.Type != errorType {
				t.Errorf("Expected error type '%s', got '%s'", errorType, err.Type)
			}
		}
	})
}
