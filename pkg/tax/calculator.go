package tax

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
)

// Tax represents a tax rate from the database
type Tax struct {
	ID                uuid.UUID
	Name              string
	AmountType        string  // 'percent', 'fixed', 'division', 'group'
	Amount            float64
	PriceInclude      bool
	IncludeBaseAmount bool
	IsBaseAffected    bool
}

// Calculator handles tax calculations
type Calculator struct {
	db *sql.DB
}

// NewCalculator creates a new tax calculator
func NewCalculator(db *sql.DB) *Calculator {
	return &Calculator{
		db: db,
	}
}

// CalculateLineTax calculates tax for a single line item
func (c *Calculator) CalculateLineTax(ctx context.Context, taxID uuid.UUID, subtotal float64) (float64, error) {
	// Fetch tax rate from database
	tax, err := c.GetTax(ctx, taxID)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch tax: %w", err)
	}

	return c.ComputeTax(tax, subtotal), nil
}

// GetTax fetches a tax record from the database
func (c *Calculator) GetTax(ctx context.Context, taxID uuid.UUID) (*Tax, error) {
	query := `
		SELECT id, name, amount_type, amount, price_include, include_base_amount, is_base_affected
		FROM account_taxes
		WHERE id = $1 AND active = true
	`

	var tax Tax
	err := c.db.QueryRowContext(ctx, query, taxID).Scan(
		&tax.ID,
		&tax.Name,
		&tax.AmountType,
		&tax.Amount,
		&tax.PriceInclude,
		&tax.IncludeBaseAmount,
		&tax.IsBaseAffected,
	)
	if err != nil {
		return nil, fmt.Errorf("tax not found: %w", err)
	}

	return &tax, nil
}

// ComputeTax computes the tax amount based on tax configuration
func (c *Calculator) ComputeTax(tax *Tax, subtotal float64) float64 {
	if tax == nil {
		return 0
	}

	switch tax.AmountType {
	case "percent":
		// Percentage-based tax (most common)
		return subtotal * (tax.Amount / 100.0)

	case "fixed":
		// Fixed amount tax (flat fee per line)
		return tax.Amount

	case "division":
		// Division-based tax (used for tax-included prices)
		// Formula: subtotal / (1 + rate/100) * (rate/100)
		if tax.Amount > 0 {
			return subtotal / (1 + tax.Amount/100.0) * (tax.Amount / 100.0)
		}
		return 0

	case "group":
		// Group tax would require fetching child taxes
		// For now, return 0 (would need recursive calculation)
		return 0

	default:
		return 0
	}
}

// CalculateMultipleTaxes calculates total tax when multiple taxes apply
func (c *Calculator) CalculateMultipleTaxes(ctx context.Context, taxIDs []uuid.UUID, subtotal float64) (float64, error) {
	if len(taxIDs) == 0 {
		return 0, nil
	}

	totalTax := 0.0
	baseAmount := subtotal

	for _, taxID := range taxIDs {
		tax, err := c.GetTax(ctx, taxID)
		if err != nil {
			// Skip invalid taxes instead of failing
			continue
		}

		// Calculate tax on the current base amount
		taxAmount := c.ComputeTax(tax, baseAmount)
		totalTax += taxAmount

		// If this tax affects the base for subsequent taxes
		if tax.IsBaseAffected {
			baseAmount += taxAmount
		}
	}

	return totalTax, nil
}
