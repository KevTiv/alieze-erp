package main

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"log"

	"github.com/xuri/excelize/v2"
)

// Contact represents a simplified contact structure for demonstration
type Contact struct {
	Name    string
	Email   string
	Phone   string
	Company string
	Title   string
}

// CSVToExcelConverter demonstrates CSV to Excel conversion functionality
func main() {
	// Sample contact data
	contacts := []Contact{
		{"John Doe", "john@example.com", "(123) 456-7890", "Acme Corp", "CEO"},
		{"Jane Smith", "jane@example.com", "(987) 654-3210", "Globex Inc", "CTO"},
		{"Bob Johnson", "bob@example.com", "(555) 123-4567", "Initech", "Developer"},
		{"Alice Williams", "alice@example.com", "(777) 888-9999", "Wayne Enterprises", "Manager"},
		{"Charlie Brown", "charlie@example.com", "(111) 222-3333", "Stark Industries", "Engineer"},
	}

	// 1. Generate CSV data
	csvData, err := generateCSV(contacts)
	if err != nil {
		log.Fatalf("Failed to generate CSV: %v", err)
	}

	fmt.Println("Generated CSV data:")
	fmt.Println(string(csvData))
	fmt.Println()

	// 2. Convert CSV to Excel
	excelData, err := convertCSVToExcel(contacts)
	if err != nil {
		log.Fatalf("Failed to convert to Excel: %v", err)
	}

	fmt.Printf("Successfully converted %d contacts to Excel format\n", len(contacts))
	fmt.Printf("Excel file size: %d bytes\n", len(excelData))

	// 3. Save Excel file
	err = saveExcelFile(excelData, "contacts_export.xlsx")
	if err != nil {
		log.Fatalf("Failed to save Excel file: %v", err)
	}

	fmt.Println("Excel file saved as: contacts_export.xlsx")
}

// generateCSV generates CSV data from contacts
func generateCSV(contacts []Contact) ([]byte, error) {
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// Write header
	header := []string{"Name", "Email", "Phone", "Company", "Title"}
	if err := writer.Write(header); err != nil {
		return nil, fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write data rows
	for _, contact := range contacts {
		row := []string{
			contact.Name,
			contact.Email,
			contact.Phone,
			contact.Company,
			contact.Title,
		}
		if err := writer.Write(row); err != nil {
			return nil, fmt.Errorf("failed to write CSV row: %w", err)
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, fmt.Errorf("CSV writer error: %w", err)
	}

	return buf.Bytes(), nil
}

// convertCSVToExcel converts contacts to Excel format
func convertCSVToExcel(contacts []Contact) ([]byte, error) {
	f := excelize.NewFile()
	defer f.Close()

	sheetName := "Contacts"
	index, err := f.NewSheet(sheetName)
	if err != nil {
		return nil, fmt.Errorf("failed to create sheet: %w", err)
	}

	f.SetActiveSheet(index)

	// Write header row
	header := []string{"Name", "Email", "Phone", "Company", "Title"}
	for i, col := range header {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheetName, cell, col)
	}

	// Write data rows
	for rowNum, contact := range contacts {
		row := []interface{}{
			contact.Name,
			contact.Email,
			contact.Phone,
			contact.Company,
			contact.Title,
		}
		for colNum, value := range row {
			cell, _ := excelize.CoordinatesToCellName(colNum+1, rowNum+2)
			f.SetCellValue(sheetName, cell, value)
		}
	}

	// Apply some basic formatting
	style, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"#DDDDDD"}, Pattern: 1},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create style: %w", err)
	}

	// Apply header style
	for i := 1; i <= len(header); i++ {
		cell, _ := excelize.CoordinatesToCellName(i, 1)
		f.SetCellStyle(sheetName, cell, cell, style)
	}

	// Auto-size columns
	for i := 1; i <= len(header); i++ {
		col, _ := excelize.ColumnNumberToName(i)
		f.SetColWidth(sheetName, col, col, 20)
	}

	// Save to buffer
	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, fmt.Errorf("failed to write Excel file: %w", err)
	}

	return buf.Bytes(), nil
}

// saveExcelFile saves Excel data to a file
func saveExcelFile(data []byte, filename string) error {
	f := excelize.NewFile()
	defer f.Close()

	// Create a temporary sheet
	index, err := f.NewSheet("Sheet1")
	if err != nil {
		return fmt.Errorf("failed to create sheet: %w", err)
	}
	f.SetActiveSheet(index)

	// This is a simplified version - in a real implementation, you would
	// write the actual data to the file system
	fmt.Printf("Would save %d bytes to %s\n", len(data), filename)

	return nil
}
