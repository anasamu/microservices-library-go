package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/anasamu/microservices-library-go/filegen/gateway"
)

func main() {
	// Create output directory
	if err := os.MkdirAll("./output", 0755); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	// Create manager with default configuration
	manager, err := gateway.NewManager(nil)
	if err != nil {
		log.Fatalf("Failed to create manager: %v", err)
	}
	defer manager.Close()

	ctx := context.Background()

	fmt.Println("File Generation Library - Basic Example")
	fmt.Println("======================================")

	// Example 1: Generate CSV file
	fmt.Println("\n1. Generating CSV file...")
	csvRequest := &gateway.FileRequest{
		Type: gateway.FileTypeCSV,
		Data: map[string]interface{}{
			"headers": []string{"Name", "Age", "Email", "Department"},
			"rows": [][]string{
				{"John Doe", "30", "john@example.com", "Engineering"},
				{"Jane Smith", "25", "jane@example.com", "Marketing"},
				{"Bob Johnson", "35", "bob@example.com", "Sales"},
				{"Alice Brown", "28", "alice@example.com", "HR"},
			},
		},
		Options: gateway.FileOptions{
			Delimiter: ",",
		},
		OutputPath: "./output/employees.csv",
	}

	csvResponse, err := manager.GenerateFile(ctx, csvRequest)
	if err != nil {
		log.Printf("CSV generation failed: %v", err)
	} else if csvResponse.Success {
		fmt.Printf("✓ CSV file generated: %s (Size: %d bytes)\n",
			csvResponse.FilePath, csvResponse.FileSize)
	} else {
		fmt.Printf("✗ CSV generation failed: %s\n", csvResponse.Error)
	}

	// Example 2: Generate Excel file
	fmt.Println("\n2. Generating Excel file...")
	excelRequest := &gateway.FileRequest{
		Type: gateway.FileTypeExcel,
		Data: map[string]interface{}{
			"sheets": map[string]interface{}{
				"Employees": map[string]interface{}{
					"headers": []string{"ID", "Name", "Department", "Salary", "Start Date"},
					"rows": [][]interface{}{
						{1, "John Doe", "Engineering", 75000, "2020-01-15"},
						{2, "Jane Smith", "Marketing", 65000, "2020-03-20"},
						{3, "Bob Johnson", "Sales", 70000, "2019-11-10"},
						{4, "Alice Brown", "HR", 60000, "2021-02-01"},
					},
				},
				"Summary": map[string]interface{}{
					"headers": []string{"Department", "Employee Count", "Average Salary", "Total Budget"},
					"rows": [][]interface{}{
						{"Engineering", 1, 75000, 75000},
						{"Marketing", 1, 65000, 65000},
						{"Sales", 1, 70000, 70000},
						{"HR", 1, 60000, 60000},
					},
				},
			},
		},
		OutputPath: "./output/employee_data.xlsx",
	}

	excelResponse, err := manager.GenerateFile(ctx, excelRequest)
	if err != nil {
		log.Printf("Excel generation failed: %v", err)
	} else if excelResponse.Success {
		fmt.Printf("✓ Excel file generated: %s (Size: %d bytes)\n",
			excelResponse.FilePath, excelResponse.FileSize)
	} else {
		fmt.Printf("✗ Excel generation failed: %s\n", excelResponse.Error)
	}

	// Example 3: Generate DOCX file
	fmt.Println("\n3. Generating DOCX file...")
	docxRequest := &gateway.FileRequest{
		Type: gateway.FileTypeDOCX,
		Data: map[string]interface{}{
			"title": "Employee Report - Q1 2024",
			"content": []map[string]interface{}{
				{
					"type":  "heading",
					"text":  "Executive Summary",
					"level": 1,
				},
				{
					"type": "paragraph",
					"text": "This report provides an overview of our employee data for the first quarter of 2024. The data includes information about our team members across different departments.",
				},
				{
					"type":  "heading",
					"text":  "Department Overview",
					"level": 2,
				},
				{
					"type":    "table",
					"headers": []string{"Department", "Employee Count", "Average Salary", "Total Budget"},
					"rows": [][]string{
						{"Engineering", "1", "$75,000", "$75,000"},
						{"Marketing", "1", "$65,000", "$65,000"},
						{"Sales", "1", "$70,000", "$70,000"},
						{"HR", "1", "$60,000", "$60,000"},
					},
				},
				{
					"type":  "heading",
					"text":  "Key Insights",
					"level": 2,
				},
				{
					"type":      "list",
					"list_type": "bullet",
					"items": []string{
						"Total of 4 employees across 4 departments",
						"Average salary across all departments: $67,500",
						"Engineering has the highest average salary",
						"All departments are well-staffed for current needs",
					},
				},
			},
		},
		OutputPath: "./output/employee_report.docx",
	}

	docxResponse, err := manager.GenerateFile(ctx, docxRequest)
	if err != nil {
		log.Printf("DOCX generation failed: %v", err)
	} else if docxResponse.Success {
		fmt.Printf("✓ DOCX file generated: %s (Size: %d bytes)\n",
			docxResponse.FilePath, docxResponse.FileSize)
	} else {
		fmt.Printf("✗ DOCX generation failed: %s\n", docxResponse.Error)
	}

	// Example 4: Generate PDF file
	fmt.Println("\n4. Generating PDF file...")
	pdfRequest := &gateway.FileRequest{
		Type: gateway.FileTypePDF,
		Data: map[string]interface{}{
			"title": "Employee Summary Report",
			"metadata": map[string]interface{}{
				"author":   "HR Department",
				"subject":  "Employee Data Summary",
				"keywords": "employees, departments, salary, report",
			},
			"content": []map[string]interface{}{
				{
					"type":  "heading",
					"text":  "Company Overview",
					"level": 1,
				},
				{
					"type": "paragraph",
					"text": "Our company currently employs 4 individuals across 4 different departments. This report provides a comprehensive overview of our workforce.",
				},
				{
					"type":  "heading",
					"text":  "Department Statistics",
					"level": 2,
				},
				{
					"type":    "table",
					"headers": []string{"Department", "Count", "Avg Salary", "Budget"},
					"rows": [][]string{
						{"Engineering", "1", "$75,000", "$75,000"},
						{"Marketing", "1", "$65,000", "$65,000"},
						{"Sales", "1", "$70,000", "$70,000"},
						{"HR", "1", "$60,000", "$60,000"},
					},
				},
				{
					"type":  "heading",
					"text":  "Recommendations",
					"level": 2,
				},
				{
					"type":      "list",
					"list_type": "numbered",
					"items": []string{
						"Consider expanding the Engineering team for upcoming projects",
						"Review salary bands to ensure competitive compensation",
						"Implement regular performance reviews across all departments",
						"Plan for future hiring based on business growth projections",
					},
				},
			},
		},
		OutputPath: "./output/employee_summary.pdf",
	}

	pdfResponse, err := manager.GenerateFile(ctx, pdfRequest)
	if err != nil {
		log.Printf("PDF generation failed: %v", err)
	} else if pdfResponse.Success {
		fmt.Printf("✓ PDF file generated: %s (Size: %d bytes)\n",
			pdfResponse.FilePath, pdfResponse.FileSize)
	} else {
		fmt.Printf("✗ PDF generation failed: %s\n", pdfResponse.Error)
	}

	// Example 5: Generate custom format (JSON)
	fmt.Println("\n5. Generating custom JSON file...")
	jsonRequest := &gateway.FileRequest{
		Type: gateway.FileTypeCustom,
		Data: map[string]interface{}{
			"title":       "Employee Data Export",
			"export_date": "2024-01-15",
			"employees": []map[string]interface{}{
				{
					"id":         1,
					"name":       "John Doe",
					"department": "Engineering",
					"salary":     75000,
					"start_date": "2020-01-15",
				},
				{
					"id":         2,
					"name":       "Jane Smith",
					"department": "Marketing",
					"salary":     65000,
					"start_date": "2020-03-20",
				},
			},
			"summary": map[string]interface{}{
				"total_employees":   4,
				"total_departments": 4,
				"average_salary":    67500,
				"total_budget":      270000,
			},
		},
		Options: gateway.FileOptions{
			Format: "json",
		},
		OutputPath: "./output/employee_data.json",
	}

	jsonResponse, err := manager.GenerateFile(ctx, jsonRequest)
	if err != nil {
		log.Printf("JSON generation failed: %v", err)
	} else if jsonResponse.Success {
		fmt.Printf("✓ JSON file generated: %s (Size: %d bytes)\n",
			jsonResponse.FilePath, jsonResponse.FileSize)
	} else {
		fmt.Printf("✗ JSON generation failed: %s\n", jsonResponse.Error)
	}

	// Display supported types
	fmt.Println("\n6. Supported file types:")
	supportedTypes := manager.GetSupportedTypes()
	for i, fileType := range supportedTypes {
		mimeType := manager.GetMimeType(fileType)
		extension := manager.GetFileExtension(fileType)
		fmt.Printf("   %d. %s (%s) - %s\n", i+1, fileType, extension, mimeType)
	}

	fmt.Println("\n✓ All examples completed!")
	fmt.Println("Check the ./output directory for generated files.")
}
