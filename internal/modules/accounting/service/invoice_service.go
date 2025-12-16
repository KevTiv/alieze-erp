package service

import (
	"context"
	"fmt"
	"time"

	"alieze-erp/internal/modules/accounting/types"
	"alieze-erp/internal/modules/accounting/repository"
	"alieze-erp/pkg/tax"

	"github.com/google/uuid"
)

type InvoiceService struct {
	repo         repository.InvoiceRepository
	paymentRepo  repository.PaymentRepository
	stateMachine interface{} // State machine for invoice workflow
	eventBus     interface{} // Event bus for publishing domain events
	taxCalc      *tax.Calculator
}

func NewInvoiceService(repo repository.InvoiceRepository, paymentRepo repository.PaymentRepository, taxCalc *tax.Calculator) *InvoiceService {
	return &InvoiceService{
		repo:        repo,
		paymentRepo: paymentRepo,
		taxCalc:     taxCalc,
	}
}

// NewInvoiceServiceWithStateMachine creates an invoice service with state machine support
func NewInvoiceServiceWithStateMachine(repo repository.InvoiceRepository, paymentRepo repository.PaymentRepository, taxCalc *tax.Calculator, stateMachine interface{}) *InvoiceService {
	service := NewInvoiceService(repo, paymentRepo, taxCalc)
	service.stateMachine = stateMachine
	return service
}

// NewInvoiceServiceWithDependencies creates an invoice service with all dependencies
func NewInvoiceServiceWithDependencies(repo repository.InvoiceRepository, paymentRepo repository.PaymentRepository, taxCalc *tax.Calculator, stateMachine interface{}, eventBus interface{}) *InvoiceService {
	service := NewInvoiceService(repo, paymentRepo, taxCalc)
	service.stateMachine = stateMachine
	service.eventBus = eventBus
	return service
}

func (s *InvoiceService) CreateInvoice(ctx context.Context, invoice domain.Invoice) (*domain.Invoice, error) {
	// Validate the invoice
	if err := s.validateInvoice(invoice); err != nil {
		return nil, fmt.Errorf("invalid invoice: %w", err)
	}

	// Set default values
	if invoice.Status == "" {
		invoice.Status = domain.InvoiceStatusDraft
	}
	if invoice.InvoiceDate.IsZero() {
		invoice.InvoiceDate = time.Now()
	}
	if invoice.DueDate.IsZero() {
		invoice.DueDate = invoice.InvoiceDate.AddDate(0, 0, 30) // Default 30 days
	}
	if invoice.ID == uuid.Nil {
		invoice.ID = uuid.New()
	}

	// Calculate amounts
	if err := s.calculateInvoiceAmounts(ctx, &invoice); err != nil {
		return nil, fmt.Errorf("failed to calculate invoice amounts: %w", err)
	}

	// Set residual amount equal to total amount for new invoices
	invoice.AmountResidual = invoice.AmountTotal

	// Create the invoice
	createdInvoice, err := s.repo.Create(ctx, invoice)
	if err != nil {
		return nil, fmt.Errorf("failed to create invoice: %w", err)
	}

	// Publish invoice.created event
	s.publishEvent(ctx, "invoice.created", createdInvoice)

	return createdInvoice, nil
}

func (s *InvoiceService) GetInvoice(ctx context.Context, id uuid.UUID) (*domain.Invoice, error) {
	invoice, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get invoice: %w", err)
	}
	if invoice == nil {
		return nil, nil
	}

	return invoice, nil
}

func (s *InvoiceService) ListInvoices(ctx context.Context, filters repository.InvoiceFilter) ([]domain.Invoice, error) {
	invoices, err := s.repo.FindAll(ctx, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to list invoices: %w", err)
	}

	return invoices, nil
}

func (s *InvoiceService) UpdateInvoice(ctx context.Context, invoice domain.Invoice) (*domain.Invoice, error) {
	// Validate the invoice
	if err := s.validateInvoice(invoice); err != nil {
		return nil, fmt.Errorf("invalid invoice: %w", err)
	}

	// Calculate amounts
	if err := s.calculateInvoiceAmounts(ctx, &invoice); err != nil {
		return nil, fmt.Errorf("failed to calculate invoice amounts: %w", err)
	}

	// Update the invoice
	updatedInvoice, err := s.repo.Update(ctx, invoice)
	if err != nil {
		return nil, fmt.Errorf("failed to update invoice: %w", err)
	}

	// Publish invoice.updated event
	s.publishEvent(ctx, "invoice.updated", updatedInvoice)

	return updatedInvoice, nil
}

func (s *InvoiceService) DeleteInvoice(ctx context.Context, id uuid.UUID) error {
	// Check if invoice exists
	invoice, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to check invoice existence: %w", err)
	}
	if invoice == nil {
		return nil
	}

	// Prevent deletion of confirmed invoices
	if invoice.Status == domain.InvoiceStatusPaid {
		return fmt.Errorf("cannot delete paid invoices")
	}

	// Delete the invoice
	err = s.repo.Delete(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete invoice: %w", err)
	}

	// Publish invoice.deleted event
	s.publishEvent(ctx, "invoice.deleted", map[string]interface{}{
		"id":              id,
		"organization_id": invoice.OrganizationID,
	})

	return nil
}

func (s *InvoiceService) ConfirmInvoice(ctx context.Context, id uuid.UUID) (*domain.Invoice, error) {
	invoice, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get invoice: %w", err)
	}
	if invoice == nil {
		return nil, fmt.Errorf("invoice not found")
	}

	// Use state machine for validation and transition if available
	if s.stateMachine != nil {
		if sm, ok := s.stateMachine.(interface{
			Transition(ctx context.Context, transitionName string, entity interface{}) error
		}); ok {
			if err := sm.Transition(ctx, "confirm", invoice); err != nil {
				return nil, fmt.Errorf("failed to confirm invoice: %w", err)
			}
			// State machine has updated the status
		} else {
			// Fallback to hardcoded validation
			if invoice.Status != domain.InvoiceStatusDraft {
				return nil, fmt.Errorf("only draft invoices can be confirmed")
			}

			if len(invoice.Lines) == 0 {
				return nil, fmt.Errorf("invoice must have at least one line to be confirmed")
			}

			// Update status to open
			invoice.Status = domain.InvoiceStatusOpen
		}
	} else {
		// Fallback to hardcoded validation
		if invoice.Status != domain.InvoiceStatusDraft {
			return nil, fmt.Errorf("only draft invoices can be confirmed")
		}

		if len(invoice.Lines) == 0 {
			return nil, fmt.Errorf("invoice must have at least one line to be confirmed")
		}

		// Update status to open
		invoice.Status = domain.InvoiceStatusOpen
	}
	invoice.UpdatedAt = time.Now()

	updatedInvoice, err := s.repo.Update(ctx, *invoice)
	if err != nil {
		return nil, fmt.Errorf("failed to update invoice: %w", err)
	}

	// Publish invoice.confirmed event
	s.publishEvent(ctx, "invoice.confirmed", updatedInvoice)

	return updatedInvoice, nil
}

func (s *InvoiceService) CancelInvoice(ctx context.Context, id uuid.UUID) (*domain.Invoice, error) {
	invoice, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get invoice: %w", err)
	}
	if invoice == nil {
		return nil, fmt.Errorf("invoice not found")
	}

	// Validate invoice for cancellation
	if invoice.Status == domain.InvoiceStatusCancelled || invoice.Status == domain.InvoiceStatusPaid {
		return nil, fmt.Errorf("invoice cannot be cancelled in its current state")
	}

	// Update status
	invoice.Status = domain.InvoiceStatusCancelled
	invoice.UpdatedAt = time.Now()

	updatedInvoice, err := s.repo.Update(ctx, *invoice)
	if err != nil {
		return nil, fmt.Errorf("failed to update invoice: %w", err)
	}

	// Publish invoice.cancelled event
	s.publishEvent(ctx, "invoice.cancelled", updatedInvoice)

	return updatedInvoice, nil
}

func (s *InvoiceService) RecordPayment(ctx context.Context, invoiceID uuid.UUID, payment domain.Payment) (*domain.Invoice, error) {
	// Get the invoice
	invoice, err := s.repo.FindByID(ctx, invoiceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get invoice: %w", err)
	}
	if invoice == nil {
		return nil, fmt.Errorf("invoice not found")
	}

	// Validate invoice for payment
	if invoice.Status != domain.InvoiceStatusOpen {
		return nil, fmt.Errorf("only open invoices can receive payments")
	}

	// Validate payment
	if payment.Amount <= 0 {
		return nil, fmt.Errorf("payment amount must be positive")
	}
	if payment.Amount > invoice.AmountResidual {
		return nil, fmt.Errorf("payment amount cannot exceed residual amount")
	}

	// Set payment defaults
	if payment.ID == uuid.Nil {
		payment.ID = uuid.New()
	}
	if payment.PaymentDate.IsZero() {
		payment.PaymentDate = time.Now()
	}
	payment.InvoiceID = invoiceID
	payment.PartnerID = invoice.PartnerID
	payment.OrganizationID = invoice.OrganizationID
	payment.CompanyID = invoice.CompanyID
	payment.CurrencyID = invoice.CurrencyID

	// Create the payment
	_, err = s.paymentRepo.Create(ctx, payment)
	if err != nil {
		return nil, fmt.Errorf("failed to create payment: %w", err)
	}

	// Update invoice residual amount
	invoice.AmountResidual -= payment.Amount

	// Check if invoice is now paid
	if invoice.AmountResidual <= 0 {
		invoice.Status = domain.InvoiceStatusPaid
	}

	invoice.UpdatedAt = time.Now()

	// Update the invoice
	updatedInvoice, err := s.repo.Update(ctx, *invoice)
	if err != nil {
		return nil, fmt.Errorf("failed to update invoice: %w", err)
	}

	// Publish payment.received event
	s.publishEvent(ctx, "payment.received", map[string]interface{}{
		"invoice_id": invoiceID,
		"payment":    payment,
		"invoice":    updatedInvoice,
	})

	// If invoice is now paid, publish invoice.paid event
	if updatedInvoice.Status == domain.InvoiceStatusPaid {
		s.publishEvent(ctx, "invoice.paid", updatedInvoice)
	}

	return updatedInvoice, nil
}

func (s *InvoiceService) GetInvoicesByPartner(ctx context.Context, partnerID uuid.UUID) ([]domain.Invoice, error) {
	invoices, err := s.repo.FindByPartnerID(ctx, partnerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get invoices by partner: %w", err)
	}

	return invoices, nil
}

func (s *InvoiceService) GetInvoicesByStatus(ctx context.Context, status domain.InvoiceStatus) ([]domain.Invoice, error) {
	invoices, err := s.repo.FindByStatus(ctx, status)
	if err != nil {
		return nil, fmt.Errorf("failed to get invoices by status: %w", err)
	}

	return invoices, nil
}

func (s *InvoiceService) GetInvoicesByType(ctx context.Context, invoiceType domain.InvoiceType) ([]domain.Invoice, error) {
	invoices, err := s.repo.FindByType(ctx, invoiceType)
	if err != nil {
		return nil, fmt.Errorf("failed to get invoices by type: %w", err)
	}

	return invoices, nil
}

func (s *InvoiceService) validateInvoice(invoice domain.Invoice) error {
	if invoice.OrganizationID == uuid.Nil {
		return fmt.Errorf("organization ID is required")
	}

	if invoice.CompanyID == uuid.Nil {
		return fmt.Errorf("company ID is required")
	}

	if invoice.PartnerID == uuid.Nil {
		return fmt.Errorf("partner ID is required")
	}

	if invoice.CurrencyID == uuid.Nil {
		return fmt.Errorf("currency ID is required")
	}

	if invoice.JournalID == uuid.Nil {
		return fmt.Errorf("journal ID is required")
	}

	if invoice.Type == "" {
		return fmt.Errorf("invoice type is required")
	}

	// Validate lines
	if len(invoice.Lines) == 0 {
		return fmt.Errorf("invoice must have at least one line")
	}

	for _, line := range invoice.Lines {
		if line.AccountID == uuid.Nil {
			return fmt.Errorf("account ID is required for all lines")
		}
		if line.Quantity <= 0 {
			return fmt.Errorf("quantity must be positive for all lines")
		}
		if line.UnitPrice < 0 {
			return fmt.Errorf("unit price cannot be negative")
		}
	}

	return nil
}

func (s *InvoiceService) calculateInvoiceAmounts(ctx context.Context, invoice *domain.Invoice) error {
	var amountUntaxed, amountTax, amountTotal float64

	for i, line := range invoice.Lines {
		// Calculate line subtotal
		line.PriceSubtotal = line.Quantity * line.UnitPrice
		if line.Discount > 0 {
			line.PriceSubtotal = line.PriceSubtotal * (1 - line.Discount/100)
		}

		// Calculate line tax using tax calculator
		line.PriceTax = 0
		if line.TaxID != nil && s.taxCalc != nil {
			taxAmount, err := s.taxCalc.CalculateLineTax(ctx, *line.TaxID, line.PriceSubtotal)
			if err != nil {
				// Log error but don't fail the calculation
				fmt.Printf("Failed to calculate tax for line: %v\n", err)
				line.PriceTax = 0
			} else {
				line.PriceTax = taxAmount
			}
		}

		// Calculate line total
		line.PriceTotal = line.PriceSubtotal + line.PriceTax

		// Update the line in the slice
		invoice.Lines[i] = line

		// Accumulate invoice totals
		amountUntaxed += line.PriceSubtotal
		amountTax += line.PriceTax
		amountTotal += line.PriceTotal
	}

	invoice.AmountUntaxed = amountUntaxed
	invoice.AmountTax = amountTax
	invoice.AmountTotal = amountTotal

	return nil
}

// publishEvent publishes an event to the event bus if available
func (s *InvoiceService) publishEvent(ctx context.Context, eventType string, payload interface{}) {
	if s.eventBus != nil {
		if bus, ok := s.eventBus.(interface {
			Publish(ctx context.Context, eventType string, payload interface{}) error
		}); ok {
			if err := bus.Publish(ctx, eventType, payload); err != nil {
				// Log error but don't fail the operation
				fmt.Printf("Failed to publish event %s: %v\n", eventType, err)
			}
		}
	}
}
