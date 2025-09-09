package integration

import (
	"context"
	"os"
	"strings"
	"testing"

	csvProvider "github.com/anasamu/microservices-library-go/filegen/providers/csv"
)

func TestCSVProvider(t *testing.T) {
	// Create temporary directory for templates
	tmpDir := t.TempDir()

	config := &csvProvider.Config{
		TemplatePath:     tmpDir,
		DefaultDelimiter: ',',
		DefaultEncoding:  "utf-8",
	}

	provider, err := csvProvider.NewProvider(config)
	if err != nil {
		t.Fatalf("Failed to create CSV provider: %v", err)
	}
	defer provider.Close()

	ctx := context.Background()

	t.Run("GenerateSimpleCSV", func(t *testing.T) {
		req := &csvProvider.FileRequest{
			Type: "csv",
			Data: map[string]interface{}{
				"headers": []string{"Name", "Age", "City"},
				"rows": [][]string{
					{"John Doe", "30", "New York"},
					{"Jane Smith", "25", "Los Angeles"},
					{"Bob Johnson", "35", "Chicago"},
				},
			},
		}

		response, err := provider.GenerateFile(ctx, req)
		if err != nil {
			t.Fatalf("Failed to generate CSV file: %v", err)
		}

		if !response.Success {
			t.Fatalf("File generation failed: %s", response.Error)
		}

		if len(response.Content) == 0 {
			t.Fatal("Generated content is empty")
		}

		if response.MimeType != "text/csv" {
			t.Errorf("Expected MIME type 'text/csv', got '%s'", response.MimeType)
		}

		// Verify CSV content
		content := string(response.Content)
		lines := strings.Split(content, "\n")
		if len(lines) < 4 { // 1 header + 3 data rows
			t.Errorf("Expected at least 4 lines, got %d", len(lines))
		}

		// Check header
		if !strings.Contains(lines[0], "Name,Age,City") {
			t.Errorf("Header not found in content: %s", lines[0])
		}
	})

	t.Run("GenerateCSVWithCustomDelimiter", func(t *testing.T) {
		req := &csvProvider.FileRequest{
			Type: "csv",
			Data: map[string]interface{}{
				"headers": []string{"Name", "Age", "City"},
				"rows": [][]string{
					{"John Doe", "30", "New York"},
					{"Jane Smith", "25", "Los Angeles"},
				},
			},
			Options: csvProvider.FileOptions{
				Delimiter: ";",
			},
		}

		response, err := provider.GenerateFile(ctx, req)
		if err != nil {
			t.Fatalf("Failed to generate CSV file with custom delimiter: %v", err)
		}

		if !response.Success {
			t.Fatalf("File generation failed: %s", response.Error)
		}

		// Verify semicolon delimiter
		content := string(response.Content)
		if !strings.Contains(content, "Name;Age;City") {
			t.Errorf("Custom delimiter not used in content: %s", content)
		}
	})

	t.Run("GenerateCSVFromRecords", func(t *testing.T) {
		req := &csvProvider.FileRequest{
			Type: "csv",
			Data: map[string]interface{}{
				"records": []map[string]interface{}{
					{"name": "John Doe", "age": 30, "city": "New York"},
					{"name": "Jane Smith", "age": 25, "city": "Los Angeles"},
					{"name": "Bob Johnson", "age": 35, "city": "Chicago"},
				},
			},
		}

		response, err := provider.GenerateFile(ctx, req)
		if err != nil {
			t.Fatalf("Failed to generate CSV from records: %v", err)
		}

		if !response.Success {
			t.Fatalf("File generation failed: %s", response.Error)
		}

		// Verify CSV content
		content := string(response.Content)
		lines := strings.Split(content, "\n")
		if len(lines) < 4 { // 1 header + 3 data rows
			t.Errorf("Expected at least 4 lines, got %d", len(lines))
		}
	})

	t.Run("ValidateRequest", func(t *testing.T) {
		// Test valid request
		req := &csvProvider.FileRequest{
			Type: "csv",
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
		req.Type = "csv"
		req.Data = nil
		if err := provider.ValidateRequest(req); err == nil {
			t.Error("Missing data should fail validation")
		}
	})

	t.Run("GetSupportedTypes", func(t *testing.T) {
		types := provider.GetSupportedTypes()
		if len(types) != 1 || types[0] != "csv" {
			t.Errorf("Expected supported types ['csv'], got %v", types)
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

func TestCSVProviderWithWriter(t *testing.T) {
	config := &csvProvider.Config{
		TemplatePath: t.TempDir(),
	}

	provider, err := csvProvider.NewProvider(config)
	if err != nil {
		t.Fatalf("Failed to create CSV provider: %v", err)
	}
	defer provider.Close()

	ctx := context.Background()

	req := &csvProvider.FileRequest{
		Type: "csv",
		Data: map[string]interface{}{
			"headers": []string{"Product", "Price", "Stock"},
			"rows": [][]string{
				{"Laptop", "999", "50"},
				{"Mouse", "25", "200"},
				{"Keyboard", "75", "150"},
			},
		},
	}

	// Create temporary file
	tmpFile, err := os.CreateTemp("", "test_*.csv")
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

	// Verify content
	tmpFile.Seek(0, 0)
	content := make([]byte, info.Size())
	_, err = tmpFile.Read(content)
	if err != nil {
		t.Fatalf("Failed to read file content: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "Product,Price,Stock") {
		t.Errorf("Expected header in content: %s", contentStr)
	}
}
