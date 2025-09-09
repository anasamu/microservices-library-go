package integration

import (
	"context"
	"os"
	"testing"

	excelProvider "github.com/anasamu/microservices-library-go/filegen/providers/excel"
)

func TestExcelProvider(t *testing.T) {
	// Create temporary directory for templates
	tmpDir := t.TempDir()

	config := &excelProvider.Config{
		TemplatePath: tmpDir,
		DefaultSheet: "Sheet1",
	}

	provider, err := excelProvider.NewProvider(config)
	if err != nil {
		t.Fatalf("Failed to create Excel provider: %v", err)
	}
	defer provider.Close()

	ctx := context.Background()

	t.Run("GenerateSimpleWorkbook", func(t *testing.T) {
		req := &excelProvider.FileRequest{
			Type: "excel",
			Data: map[string]interface{}{
				"headers": []string{"Name", "Age", "City"},
				"rows": [][]interface{}{
					{"John Doe", 30, "New York"},
					{"Jane Smith", 25, "Los Angeles"},
					{"Bob Johnson", 35, "Chicago"},
				},
			},
		}

		response, err := provider.GenerateFile(ctx, req)
		if err != nil {
			t.Fatalf("Failed to generate Excel file: %v", err)
		}

		if !response.Success {
			t.Fatalf("File generation failed: %s", response.Error)
		}

		if len(response.Content) == 0 {
			t.Fatal("Generated content is empty")
		}

		if response.MimeType != "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet" {
			t.Errorf("Expected MIME type 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet', got '%s'", response.MimeType)
		}
	})

	t.Run("GenerateMultiSheetWorkbook", func(t *testing.T) {
		req := &excelProvider.FileRequest{
			Type: "excel",
			Data: map[string]interface{}{
				"sheets": map[string]interface{}{
					"Employees": map[string]interface{}{
						"headers": []string{"ID", "Name", "Department", "Salary"},
						"rows": [][]interface{}{
							{1, "John Doe", "Engineering", 75000},
							{2, "Jane Smith", "Marketing", 65000},
							{3, "Bob Johnson", "Sales", 70000},
						},
					},
					"Summary": map[string]interface{}{
						"headers": []string{"Department", "Count", "Avg Salary"},
						"rows": [][]interface{}{
							{"Engineering", 1, 75000},
							{"Marketing", 1, 65000},
							{"Sales", 1, 70000},
						},
					},
				},
			},
		}

		response, err := provider.GenerateFile(ctx, req)
		if err != nil {
			t.Fatalf("Failed to generate multi-sheet Excel file: %v", err)
		}

		if !response.Success {
			t.Fatalf("File generation failed: %s", response.Error)
		}

		if len(response.Content) == 0 {
			t.Fatal("Generated content is empty")
		}
	})

	t.Run("ValidateRequest", func(t *testing.T) {
		// Test valid request
		req := &excelProvider.FileRequest{
			Type: "excel",
			Data: map[string]interface{}{"test": "data"},
		}

		if err := provider.ValidateRequest(req); err != nil {
			t.Errorf("Valid request should not fail validation: %v", err)
		}

		// Test invalid file type
		req.Type = "invalid"
		if err := provider.ValidateRequest(req); err == nil {
			t.Error("Invalid file type should fail validation")
		}

		// Test missing data
		req.Type = "excel"
		req.Data = nil
		if err := provider.ValidateRequest(req); err == nil {
			t.Error("Missing data should fail validation")
		}
	})

	t.Run("GetSupportedTypes", func(t *testing.T) {
		types := provider.GetSupportedTypes()
		if len(types) != 1 || types[0] != "excel" {
			t.Errorf("Expected supported types ['excel'], got %v", types)
		}
	})

	t.Run("GetTemplateList", func(t *testing.T) {
		templates, err := provider.GetTemplateList()
		if err != nil {
			t.Errorf("Failed to get template list: %v", err)
		}

		// Should be empty since no templates were loaded
		if len(templates) != 0 {
			t.Errorf("Expected empty template list, got %v", templates)
		}
	})
}

func TestExcelProviderWithWriter(t *testing.T) {
	config := &excelProvider.Config{
		TemplatePath: t.TempDir(),
	}

	provider, err := excelProvider.NewProvider(config)
	if err != nil {
		t.Fatalf("Failed to create Excel provider: %v", err)
	}
	defer provider.Close()

	ctx := context.Background()

	req := &excelProvider.FileRequest{
		Type: "excel",
		Data: map[string]interface{}{
			"headers": []string{"Product", "Price", "Stock"},
			"rows": [][]interface{}{
				{"Laptop", 999, 50},
				{"Mouse", 25, 200},
				{"Keyboard", 75, 150},
			},
		},
	}

	// Create temporary file
	tmpFile, err := os.CreateTemp("", "test_*.xlsx")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Generate file to writer
	err = provider.GenerateFileToWriter(ctx, req, tmpFile)
	if err != nil {
		t.Fatalf("Failed to generate file to writer: %v", err)
	}

	// Check if file was written
	info, err := tmpFile.Stat()
	if err != nil {
		t.Fatalf("Failed to get file info: %v", err)
	}

	if info.Size() == 0 {
		t.Fatal("Generated file is empty")
	}
}
