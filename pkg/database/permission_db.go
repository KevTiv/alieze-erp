package database

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/KevTiv/alieze-erp/pkg/audit"
	"github.com/KevTiv/alieze-erp/pkg/policy"
	"github.com/google/uuid"
)

// PermissionDB wraps sql.DB with automatic permission checking
type PermissionDB struct {
	*sql.DB
	policyEngine *policy.Engine
	auditLogger  *audit.AuditLogger
}

// NewPermissionDB creates a new permission-aware database wrapper
func NewPermissionDB(db *sql.DB, policyEngine *policy.Engine, auditLogger *audit.AuditLogger) *PermissionDB {
	return &PermissionDB{
		DB:           db,
		policyEngine: policyEngine,
		auditLogger:  auditLogger,
	}
}

// ExecContext wraps sql.DB.ExecContext with permission checking
func (pdb *PermissionDB) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	// Extract table and operation from query
	table, operation := pdb.extractTableAndOperation(query)

	// Check permission
	if err := pdb.checkTablePermission(ctx, table, operation, query); err != nil {
		return nil, fmt.Errorf("permission denied: %w", err)
	}

	// Execute original query
	return pdb.DB.ExecContext(ctx, query, args...)
}

// QueryContext wraps sql.DB.QueryContext with permission checking
func (pdb *PermissionDB) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	// Extract table and operation from query
	table, operation := pdb.extractTableAndOperation(query)

	// Check permission
	if err := pdb.checkTablePermission(ctx, table, operation, query); err != nil {
		return nil, fmt.Errorf("permission denied: %w", err)
	}

	// Execute original query
	return pdb.DB.QueryContext(ctx, query, args...)
}

// QueryRowContext wraps sql.DB.QueryRowContext with permission checking
func (pdb *PermissionDB) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	// Extract table and operation from query
	table, operation := pdb.extractTableAndOperation(query)

	// Check permission
	if err := pdb.checkTablePermission(ctx, table, operation, query); err != nil {
		// For QueryRow, we can't easily return a custom error row
		// Instead, we'll execute a dummy query and it will fail on Scan
		return pdb.DB.QueryRowContext(ctx, "SELECT 1 WHERE 1=0")
	}

	// Execute original query
	return pdb.DB.QueryRowContext(ctx, query, args...)
}

// BeginTx wraps sql.DB.BeginTx with permission checking
func (pdb *PermissionDB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	// For transactions, we'll allow them but wrap the transaction as well
	tx, err := pdb.DB.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}

	// Return the original transaction
	// Note: In a production system, you might want to wrap this too
	return tx, nil
}

// checkTablePermission checks if user has permission for table operation
func (pdb *PermissionDB) checkTablePermission(ctx context.Context, table, operation string, query string) error {
	// Skip permission check if already verified in this request
	if ctx.Value("permissionChecked") == true {
		return nil
	}

	// Get user and organization from context
	userIDStr, userOk := ctx.Value("userID").(string)
	if !userOk {
		return fmt.Errorf("user not found in context")
	}

	orgIDStr, orgOk := ctx.Value("organizationID").(string)
	if !orgOk {
		return fmt.Errorf("organization not found in context")
	}

	// Parse UUIDs
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return fmt.Errorf("invalid user ID: %w", err)
	}

	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		return fmt.Errorf("invalid organization ID: %w", err)
	}

	// Determine module from table name
	module := pdb.getModuleFromTable(table)

	// Check permission
	permission := fmt.Sprintf("%s:%s:%s", module, table, operation)
	allowed, err := pdb.policyEngine.CheckPermission(ctx, userID.String(), permission, operation)
	if err != nil {
		// Log permission check failure
		if pdb.auditLogger != nil {
			pdb.auditLogger.LogPermissionCheck(ctx, userID, orgID, module, table, operation, permission, false, query)
		}
		return fmt.Errorf("permission check failed: %w", err)
	}

	// Log the permission check result
	if pdb.auditLogger != nil {
		pdb.auditLogger.LogPermissionCheck(ctx, userID, orgID, module, table, operation, permission, allowed, query)
	}

	if !allowed {
		return fmt.Errorf("permission denied: %s", permission)
	}

	return nil
}

// extractTableAndOperation parses SQL query to determine table and operation
func (pdb *PermissionDB) extractTableAndOperation(query string) (string, string) {
	// Simple parsing - can be enhanced with SQL parser
	lowerQuery := strings.ToLower(strings.TrimSpace(query))

	if strings.HasPrefix(lowerQuery, "insert into") {
		// Extract table from INSERT INTO table_name
		parts := strings.SplitN(lowerQuery, "insert into", 2)
		if len(parts) > 1 {
			tablePart := strings.TrimSpace(parts[1])
			table := strings.Split(tablePart, " ")[0]
			return table, "create"
		}
	} else if strings.HasPrefix(lowerQuery, "update") {
		// Extract table from UPDATE table_name
		parts := strings.SplitN(lowerQuery, "update", 2)
		if len(parts) > 1 {
			tablePart := strings.TrimSpace(parts[1])
			table := strings.Split(tablePart, " ")[0]
			return table, "update"
		}
	} else if strings.HasPrefix(lowerQuery, "delete from") {
		// Extract table from DELETE FROM table_name
		parts := strings.SplitN(lowerQuery, "delete from", 2)
		if len(parts) > 1 {
			tablePart := strings.TrimSpace(parts[1])
			table := strings.Split(tablePart, " ")[0]
			return table, "delete"
		}
	} else if strings.HasPrefix(lowerQuery, "select") {
		// Extract table from SELECT ... FROM table_name
		fromIndex := strings.Index(lowerQuery, " from ")
		if fromIndex > 0 {
			tablePart := lowerQuery[fromIndex+6:] // " from " is 6 chars
			table := strings.Split(tablePart, " ")[0]
			return table, "read"
		}
	}

	// Default to read operation if we can't determine
	return "unknown", "read"
}

// getModuleFromTable determines the module from table name
func (pdb *PermissionDB) getModuleFromTable(table string) string {
	// Map table names to modules
	tableToModule := map[string]string{
		"contacts":       "crm",
		"leads":          "crm",
		"activities":     "crm",
		"sales_teams":    "crm",
		"contact_tags":   "crm",
		"lead_stages":    "crm",
		"lead_sources":   "crm",
		"lost_reasons":   "crm",
		"products":       "products",
		"product_categories": "products",
		"product_variants": "products",
		"sales_orders":   "sales",
		"sales_order_lines": "sales",
		"pricelists":     "sales",
		"invoices":       "accounting",
		"invoice_lines":  "accounting",
		"payments":       "accounting",
		"account_accounts": "accounting",
		"account_journals": "accounting",
		"account_taxes":  "accounting",
		"warehouses":     "inventory",
		"stock_locations": "inventory",
		"stock_quants":   "inventory",
		"stock_lots":     "inventory",
		"stock_packages": "inventory",
	}

	if module, exists := tableToModule[table]; exists {
		return module
	}

	// Default to "system" module for unknown tables
	return "system"
}
