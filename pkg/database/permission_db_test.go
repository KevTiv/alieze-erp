package database

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test table and operation extraction
func TestPermissionDB_ExtractTableAndOperation(t *testing.T) {
	// Create a mock PermissionDB for testing
	permDB := &PermissionDB{}

	// Test various SQL queries
	tests := []struct {
		query     string
		expectedTable string
		expectedOp   string
	}{
		{"INSERT INTO contacts (name) VALUES ($1)", "contacts", "create"},
		{"UPDATE contacts SET name = $1 WHERE id = $2", "contacts", "update"},
		{"DELETE FROM contacts WHERE id = $1", "contacts", "delete"},
		{"SELECT * FROM contacts WHERE id = $1", "contacts", "read"},
		{"SELECT id, name FROM users", "users", "read"},
		{"SELECT COUNT(*) FROM invoices", "invoices", "read"},
		{"insert into products (name) values ($1)", "products", "create"},
		{"update sales_orders set status = $1 where id = $2", "sales_orders", "update"},
	}

	for _, test := range tests {
		t := t
		table, op := permDB.extractTableAndOperation(test.query)
		assert.Equal(t, test.expectedTable, table, "Failed for query: %s", test.query)
		assert.Equal(t, test.expectedOp, op, "Failed for query: %s", test.query)
	}
}

// Test module mapping
func TestPermissionDB_GetModuleFromTable(t *testing.T) {
	// Create a mock PermissionDB for testing
	permDB := &PermissionDB{}

	// Test module mapping
	tests := []struct {
		table     string
		expectedModule string
	}{
		{"contacts", "crm"},
		{"leads", "crm"},
		{"activities", "crm"},
		{"sales_teams", "crm"},
		{"contact_tags", "crm"},
		{"lead_stages", "crm"},
		{"lead_sources", "crm"},
		{"lost_reasons", "crm"},
		{"products", "products"},
		{"product_categories", "products"},
		{"product_variants", "products"},
		{"sales_orders", "sales"},
		{"sales_order_lines", "sales"},
		{"pricelists", "sales"},
		{"invoices", "accounting"},
		{"invoice_lines", "accounting"},
		{"payments", "accounting"},
		{"account_accounts", "accounting"},
		{"account_journals", "accounting"},
		{"account_taxes", "accounting"},
		{"warehouses", "inventory"},
		{"stock_locations", "inventory"},
		{"stock_quants", "inventory"},
		{"stock_lots", "inventory"},
		{"stock_packages", "inventory"},
		{"unknown_table", "system"},
	}

	for _, test := range tests {
		module := permDB.getModuleFromTable(test.table)
		assert.Equal(t, test.expectedModule, module)
	}
}

// Test permission string generation
func TestPermissionDB_PermissionStringGeneration(t *testing.T) {
	// Create a mock PermissionDB for testing
	permDB := &PermissionDB{}

	// Test permission string generation
	tests := []struct {
		table     string
		operation string
		expected  string
	}{
		{"contacts", "create", "crm:contacts:create"},
		{"contacts", "read", "crm:contacts:read"},
		{"contacts", "update", "crm:contacts:update"},
		{"contacts", "delete", "crm:contacts:delete"},
		{"products", "create", "products:products:create"},
		{"products", "read", "products:products:read"},
		{"invoices", "read", "accounting:invoices:read"},
		{"invoices", "create", "accounting:invoices:create"},
		{"sales_orders", "update", "sales:sales_orders:update"},
		{"warehouses", "delete", "inventory:warehouses:delete"},
		{"unknown", "read", "system:unknown:read"},
	}

	for _, test := range tests {
		module := permDB.getModuleFromTable(test.table)
		permission := fmt.Sprintf("%s:%s:%s", module, test.table, test.operation)
		assert.Equal(t, test.expected, permission)
	}
}
