package gateway

import (
	"context"
	"fmt"
	"log"
	"os"
)

// Example demonstrates how to use the file generation manager
func Example() {
	// Create manager configuration
	config := &ManagerConfig{
		TemplatePath: "./templates",
		OutputPath:   "./output",
		MaxFileSize:  100 * 1024 * 1024, // 100MB
		AllowedTypes: []FileType{FileTypeDOCX, FileTypeExcel, FileTypeCSV, FileTypePDF},
	}

	// Initialize manager
	manager, err := NewManager(config)
	if err != nil {
		log.Fatalf("Failed to create manager: %v", err)
	}
	defer manager.Close()

	ctx := context.Background()

	// Example 1: Generate a CSV file
	fmt.Println("=== CSV Generation Example ===")
	csvRequest := &FileRequest{
		Type: FileTypeCSV,
		Data: map[string]interface{}{
			"headers": []string{"Name", "Age", "Email"},
			"rows": [][]string{
				{"John Doe", "30", "john@example.com"},
				{"Jane Smith", "25", "jane@example.com"},
				{"Bob Johnson", "35", "bob@example.com"},
			},
		},
		Options: FileOptions{
			Delimiter: ",",
		},
		OutputPath: "./output/example.csv",
	}

	csvResponse, err := manager.GenerateFile(ctx, csvRequest)
	if err != nil {
		log.Printf("CSV generation failed: %v", err)
	} else {
		fmt.Printf("CSV generated successfully: %s (Size: %d bytes)\n",
			csvResponse.FilePath, csvResponse.FileSize)
	}

	// Example 2: Generate an Excel file
	fmt.Println("\n=== Excel Generation Example ===")
	excelRequest := &FileRequest{
		Type: FileTypeExcel,
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
		OutputPath: "./output/example.xlsx",
	}

	excelResponse, err := manager.GenerateFile(ctx, excelRequest)
	if err != nil {
		log.Printf("Excel generation failed: %v", err)
	} else {
		fmt.Printf("Excel generated successfully: %s (Size: %d bytes)\n",
			excelResponse.FilePath, excelResponse.FileSize)
	}

	// Example 3: Generate a DOCX file
	fmt.Println("\n=== DOCX Generation Example ===")
	docxRequest := &FileRequest{
		Type: FileTypeDOCX,
		Data: map[string]interface{}{
			"title": "Monthly Report",
			"content": []map[string]interface{}{
				{
					"type":  "heading",
					"text":  "Executive Summary",
					"level": 1,
				},
				{
					"type": "paragraph",
					"text": "This report covers the monthly performance metrics for our organization.",
				},
				{
					"type":  "heading",
					"text":  "Key Metrics",
					"level": 2,
				},
				{
					"type":    "table",
					"headers": []string{"Metric", "Value", "Change"},
					"rows": [][]string{
						{"Revenue", "$1,000,000", "+5%"},
						{"Customers", "1,250", "+10%"},
						{"Satisfaction", "4.8/5", "+0.2"},
					},
				},
			},
		},
		OutputPath: "./output/example.docx",
	}

	docxResponse, err := manager.GenerateFile(ctx, docxRequest)
	if err != nil {
		log.Printf("DOCX generation failed: %v", err)
	} else {
		fmt.Printf("DOCX generated successfully: %s (Size: %d bytes)\n",
			docxResponse.FilePath, docxResponse.FileSize)
	}

	// Example 4: Generate a PDF file
	fmt.Println("\n=== PDF Generation Example ===")
	pdfRequest := &FileRequest{
		Type: FileTypePDF,
		Data: map[string]interface{}{
			"title": "Invoice #INV-001",
			"content": []map[string]interface{}{
				{
					"type":  "heading",
					"text":  "Invoice Details",
					"level": 1,
				},
				{
					"type": "paragraph",
					"text": "Invoice Date: 2024-01-15",
				},
				{
					"type": "paragraph",
					"text": "Due Date: 2024-02-15",
				},
				{
					"type":    "table",
					"headers": []string{"Description", "Quantity", "Price", "Total"},
					"rows": [][]string{
						{"Web Development", "40 hours", "$100/hour", "$4,000"},
						{"Design Services", "20 hours", "$80/hour", "$1,600"},
						{"Consultation", "10 hours", "$120/hour", "$1,200"},
					},
				},
				{
					"type":  "paragraph",
					"text":  "Total Amount: $6,800",
					"style": "bold",
				},
			},
		},
		OutputPath: "./output/example.pdf",
	}

	pdfResponse, err := manager.GenerateFile(ctx, pdfRequest)
	if err != nil {
		log.Printf("PDF generation failed: %v", err)
	} else {
		fmt.Printf("PDF generated successfully: %s (Size: %d bytes)\n",
			pdfResponse.FilePath, pdfResponse.FileSize)
	}

	// Example 5: Generate file to writer (in-memory)
	fmt.Println("\n=== In-Memory Generation Example ===")

	// Create a buffer to write to
	var buf []byte
	writer := &bufferWriter{&buf}

	// Generate CSV to buffer
	csvRequest.OutputPath = "" // No file output
	err = manager.GenerateFileToWriter(ctx, csvRequest, writer)
	if err != nil {
		log.Printf("In-memory CSV generation failed: %v", err)
	} else {
		fmt.Printf("In-memory CSV generated successfully (Size: %d bytes)\n", len(buf))
	}

	// Example 6: Get supported types and templates
	fmt.Println("\n=== Manager Information ===")
	supportedTypes := manager.GetSupportedTypes()
	fmt.Printf("Supported file types: %v\n", supportedTypes)

	for _, fileType := range supportedTypes {
		templates, err := manager.GetTemplateList(fileType)
		if err != nil {
			log.Printf("Failed to get templates for %s: %v", fileType, err)
			continue
		}
		fmt.Printf("Templates for %s: %v\n", fileType, templates)
	}

	// Example 7: Error handling
	fmt.Println("\n=== Error Handling Example ===")
	invalidRequest := &FileRequest{
		Type: "invalid_type",
		Data: map[string]interface{}{},
	}

	_, err = manager.GenerateFile(ctx, invalidRequest)
	if err != nil {
		fmt.Printf("Expected error for invalid type: %v\n", err)
	}
}

// bufferWriter is a simple writer that writes to a byte slice
type bufferWriter struct {
	buf *[]byte
}

func (w *bufferWriter) Write(p []byte) (n int, err error) {
	*w.buf = append(*w.buf, p...)
	return len(p), nil
}

// ExampleWithTemplates demonstrates template-based file generation
func ExampleWithTemplates() {
	config := &ManagerConfig{
		TemplatePath: "./templates",
		OutputPath:   "./output",
	}

	manager, err := NewManager(config)
	if err != nil {
		log.Fatalf("Failed to create manager: %v", err)
	}
	defer manager.Close()

	ctx := context.Background()

	// Example using a template
	templateRequest := &FileRequest{
		Type:     FileTypeDOCX,
		Template: "report",
		Data: map[string]interface{}{
			"company_name": "Acme Corp",
			"report_date":  "2024-01-15",
			"metrics": map[string]interface{}{
				"revenue":   1000000,
				"customers": 1250,
				"growth":    "15%",
			},
			"departments": []map[string]interface{}{
				{"name": "Engineering", "budget": 500000, "employees": 25},
				{"name": "Marketing", "budget": 200000, "employees": 10},
				{"name": "Sales", "budget": 300000, "employees": 15},
			},
		},
		OutputPath: "./output/template_report.docx",
	}

	response, err := manager.GenerateFile(ctx, templateRequest)
	if err != nil {
		log.Printf("Template generation failed: %v", err)
	} else {
		fmt.Printf("Template-based file generated: %s\n", response.FilePath)
	}
}

// ExampleBatchGeneration demonstrates batch file generation
func ExampleBatchGeneration() {
	config := &ManagerConfig{
		TemplatePath: "./templates",
		OutputPath:   "./output",
	}

	manager, err := NewManager(config)
	if err != nil {
		log.Fatalf("Failed to create manager: %v", err)
	}
	defer manager.Close()

	ctx := context.Background()

	// Batch generation requests
	requests := []*FileRequest{
		{
			Type: FileTypeCSV,
			Data: map[string]interface{}{
				"headers": []string{"Product", "Price", "Stock"},
				"rows": [][]string{
					{"Laptop", "$999", "50"},
					{"Mouse", "$25", "200"},
					{"Keyboard", "$75", "150"},
				},
			},
			OutputPath: "./output/products.csv",
		},
		{
			Type: FileTypeExcel,
			Data: map[string]interface{}{
				"sheets": map[string]interface{}{
					"Products": map[string]interface{}{
						"headers": []string{"Product", "Price", "Stock"},
						"rows": [][]interface{}{
							{"Laptop", 999, 50},
							{"Mouse", 25, 200},
							{"Keyboard", 75, 150},
						},
					},
				},
			},
			OutputPath: "./output/products.xlsx",
		},
	}

	// Process batch requests
	for i, req := range requests {
		response, err := manager.GenerateFile(ctx, req)
		if err != nil {
			log.Printf("Batch request %d failed: %v", i+1, err)
		} else {
			fmt.Printf("Batch request %d completed: %s\n", i+1, response.FilePath)
		}
	}
}

func main() {
	// Create output directory
	if err := os.MkdirAll("./output", 0755); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	fmt.Println("File Generation Library Examples")
	fmt.Println("================================")

	// Run examples
	Example()
	ExampleWithTemplates()
	ExampleBatchGeneration()

	fmt.Println("\nAll examples completed!")
}
