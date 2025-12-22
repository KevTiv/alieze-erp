package main

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"log"
	"strings"

	"github.com/xuri/excelize/v2"
)

// Contact represents a contact for import
type Contact struct {
	Name    string
	Email   string
	Phone   string
	Company string
	Title   string
}

// ContactImportExample demonstrates CSV and Excel import functionality
func main() {
	// Sample CSV data for import
	csvData := `Name,Email,Phone,Company,Title
John Doe,john@example.com,(123) 456-7890,Acme Corp,CEO
Jane Smith,jane@example.com,(987) 654-3210,Globex Inc,CTO
Bob Johnson,bob@example.com,(555) 123-4567,Initech,Developer
Alice Williams,alice@example.com,(777) 888-9999,Wayne Enterprises,Manager
Charlie Brown,charlie@example.com,(111) 222-3333,Stark Industries,Engineer`

	fmt.Println("=== Contact Import Example ===")
	fmt.Println()

	// 1. Parse CSV data
	fmt.Println("1. Parsing CSV data:")
	contacts, err := parseCSV(csvData)
	if err != nil {
		log.Fatalf("Failed to parse CSV: %v", err)
	}

	for i, contact := range contacts {
		fmt.Printf("   Contact %d: %s (%s)\n", i+1, contact.Name, contact.Email)
	}
	fmt.Println()

	// 2. Validate contacts
	fmt.Println("2. Validating contacts:")
	validationResults := validateContacts(contacts)
	for _, result := range validationResults {
		if result.Valid {
			fmt.Printf("   ✓ %s - Valid\n", result.Contact.Name)
		} else {
			fmt.Printf("   ✗ %s - %s\n", result.Contact.Name, result.Error)
		}
	}
	fmt.Println()

	// 3. Import valid contacts
	fmt.Println("3. Importing valid contacts:")
	importResults := importContacts(validationResults)
	fmt.Printf("   Successfully imported: %d contacts\n", importResults.Successful)
	fmt.Printf("   Failed to import: %d contacts\n", importResults.Failed)
	fmt.Println()

	// 4. Generate import report
	fmt.Println("4. Generating import report:")
	report := generateImportReport(importResults)
	fmt.Println(report)
	fmt.Println()

	// 5. Create Excel import template
	fmt.Println("5. Creating Excel import template:")
	excelTemplate, err := createExcelImportTemplate()
	if err != nil {
		log.Fatalf("Failed to create Excel template: %v", err)
	}
	fmt.Printf("   Excel template created: %d bytes\n", len(excelTemplate))
	fmt.Println()

	fmt.Println("Contact import example completed successfully!")
}

// parseCSV parses CSV data into contacts
type CSVContact struct {
	Name    string
	Email   string
	Phone   string
	Company string
	Title   string
}

func parseCSV(data string) ([]CSVContact, error) {
	reader := csv.NewReader(strings.NewReader(data))

	// Read header row
	headers, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV headers: %w", err)
	}

	// Create field mapping
	fieldMap := make(map[string]int)
	for i, header := range headers {
		fieldMap[strings.ToLower(strings.TrimSpace(header))] = i
	}

	var contacts []CSVContact

	// Read data rows
	for {
		row, err := reader.Read()
		if err != nil {
			break // End of file
		}

		contact := CSVContact{}

		// Map fields based on headers
		if nameIdx, ok := fieldMap["name"]; ok && nameIdx < len(row) {
			contact.Name = strings.TrimSpace(row[nameIdx])
		}
		if emailIdx, ok := fieldMap["email"]; ok && emailIdx < len(row) {
			contact.Email = strings.TrimSpace(row[emailIdx])
		}
		if phoneIdx, ok := fieldMap["phone"]; ok && phoneIdx < len(row) {
			contact.Phone = strings.TrimSpace(row[phoneIdx])
		}
		if companyIdx, ok := fieldMap["company"]; ok && companyIdx < len(row) {
			contact.Company = strings.TrimSpace(row[companyIdx])
		}
		if titleIdx, ok := fieldMap["title"]; ok && titleIdx < len(row) {
			contact.Title = strings.TrimSpace(row[titleIdx])
		}

		contacts = append(contacts, contact)
	}

	return contacts, nil
}

// ValidationResult represents contact validation result
type ValidationResult struct {
	Contact Contact
	Valid   bool
	Error   string
}

// validateContacts validates imported contacts
func validateContacts(csvContacts []CSVContact) []ValidationResult {
	var results []ValidationResult

	for _, csvContact := range csvContacts {
		contact := Contact{
			Name:    csvContact.Name,
			Email:   csvContact.Email,
			Phone:   csvContact.Phone,
			Company: csvContact.Company,
			Title:   csvContact.Title,
		}

		var errors []string

		// Validate required fields
		if contact.Name == "" {
			errors = append(errors, "name is required")
		}

		if contact.Email == "" {
			errors = append(errors, "email is required")
		} else if !isValidEmail(contact.Email) {
			errors = append(errors, "invalid email format")
		}

		// Validate email uniqueness (simulated)
		if contact.Email != "" && strings.Contains(contact.Email, "example.com") {
			// Simulate duplicate check
			errors = append(errors, "email already exists")
		}

		valid := len(errors) == 0
		errorMsg := ""
		if !valid {
			errorMsg = strings.Join(errors, "; ")
		}

		results = append(results, ValidationResult{
			Contact: contact,
			Valid:   valid,
			Error:   errorMsg,
		})
	}

	return results
}

// ImportResults represents import operation results
type ImportResults struct {
	Successful int
	Failed     int
	Errors     []string
}

// importContacts imports valid contacts
func importContacts(validationResults []ValidationResult) ImportResults {
	var results ImportResults

	for _, result := range validationResults {
		if result.Valid {
			// Simulate contact creation
			fmt.Printf("     Importing: %s (%s)\n", result.Contact.Name, result.Contact.Email)
			results.Successful++
		} else {
			results.Failed++
			results.Errors = append(results.Errors, fmt.Sprintf("%s: %s", result.Contact.Name, result.Error))
		}
	}

	return results
}

// generateImportReport generates an import summary report
func generateImportReport(results ImportResults) string {
	var report strings.Builder

	report.WriteString("Import Summary Report\n")
	report.WriteString("=====================\n")
	report.WriteString(fmt.Sprintf("Total Records: %d\n", results.Successful+results.Failed))
	report.WriteString(fmt.Sprintf("Successfully Imported: %d\n", results.Successful))
	report.WriteString(fmt.Sprintf("Failed to Import: %d\n", results.Failed))

	if len(results.Errors) > 0 {
		report.WriteString("\nErrors:\n")
		for _, error := range results.Errors {
			report.WriteString(fmt.Sprintf("  - %s\n", error))
		}
	}

	return report.String()
}

// createExcelImportTemplate creates an Excel import template
func createExcelImportTemplate() ([]byte, error) {
	f := excelize.NewFile()
	defer f.Close()

	// Create a sheet
	sheetName := "Import Template"
	index, err := f.NewSheet(sheetName)
	if err != nil {
		return nil, fmt.Errorf("failed to create sheet: %w", err)
	}
	f.SetActiveSheet(index)

	// Add headers
	headers := []string{"Name", "Email", "Phone", "Company", "Title"}
	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheetName, cell, header)
	}

	// Add example data
	exampleData := [][]interface{}{
		{"John Doe", "john@example.com", "(123) 456-7890", "Acme Corp", "CEO"},
		{"Jane Smith", "jane@example.com", "(987) 654-3210", "Globex Inc", "CTO"},
	}

	for rowIdx, row := range exampleData {
		for colIdx, value := range row {
			cell, _ := excelize.CoordinatesToCellName(colIdx+1, rowIdx+2)
			f.SetCellValue(sheetName, cell, value)
		}
	}

	// Add instructions
	instructions := []string{
		"Import Instructions:",
		"1. Fill in your contact data starting from row 2",
		"2. Required fields: Name, Email",
		"3. Email must be unique and valid format",
		"4. Save as CSV or Excel file",
		"5. Upload using the import function",
	}

	for i, instruction := range instructions {
		cell, _ := excelize.CoordinatesToCellName(1, i+5)
		f.SetCellValue(sheetName, cell, instruction)
	}

	// Apply formatting
	style, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"#DDDDDD"}, Pattern: 1},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create style: %w", err)
	}

	// Apply header style
	for i := 1; i <= len(headers); i++ {
		cell, _ := excelize.CoordinatesToCellName(i, 1)
		f.SetCellStyle(sheetName, cell, cell, style)
	}

	// Auto-size columns
	for i := 1; i <= len(headers); i++ {
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

// Helper functions

func isValidEmail(email string) bool {
	// Simple email validation
	return strings.Contains(email, "@") && len(email) > 5
}

func getStringPtrValue(ptr *string) string {
	if ptr == nil {
		return ""
	}
	return *ptr
}
