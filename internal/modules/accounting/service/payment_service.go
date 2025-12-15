package service

import (
	"context"
	"fmt"
	"time"

	"alieze-erp/internal/modules/accounting/types"
	"alieze-erp/internal/modules/accounting/repository"

	"github.com/google/uuid"
)

type PaymentService struct {
	repo repository.PaymentRepository
}

func NewPaymentService(repo repository.PaymentRepository) *PaymentService {
	return &PaymentService{
		repo: repo,
	}
}

func (s *PaymentService) CreatePayment(ctx context.Context, payment domain.Payment) (*domain.Payment, error) {
	// Validate the payment
	if err := s.validatePayment(payment); err != nil {
		return nil, fmt.Errorf("invalid payment: %w", err)
	}

	// Set default values
	if payment.ID == uuid.Nil {
		payment.ID = uuid.New()
	}
	if payment.PaymentDate.IsZero() {
		payment.PaymentDate = time.Now()
	}

	// Create the payment
	createdPayment, err := s.repo.Create(ctx, payment)
	if err != nil {
		return nil, fmt.Errorf("failed to create payment: %w", err)
	}

	return createdPayment, nil
}

func (s *PaymentService) GetPayment(ctx context.Context, id uuid.UUID) (*domain.Payment, error) {
	payment, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment: %w", err)
	}
	if payment == nil {
		return nil, nil
	}

	return payment, nil
}

func (s *PaymentService) ListPayments(ctx context.Context, filters repository.PaymentFilter) ([]domain.Payment, error) {
	payments, err := s.repo.FindAll(ctx, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to list payments: %w", err)
	}

	return payments, nil
}

func (s *PaymentService) UpdatePayment(ctx context.Context, payment domain.Payment) (*domain.Payment, error) {
	// Validate the payment
	if err := s.validatePayment(payment); err != nil {
		return nil, fmt.Errorf("invalid payment: %w", err)
	}

	// Update the payment
	updatedPayment, err := s.repo.Update(ctx, payment)
	if err != nil {
		return nil, fmt.Errorf("failed to update payment: %w", err)
	}

	return updatedPayment, nil
}

func (s *PaymentService) DeletePayment(ctx context.Context, id uuid.UUID) error {
	// Check if payment exists
	payment, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to check payment existence: %w", err)
	}
	if payment == nil {
		return nil
	}

	// Delete the payment
	err = s.repo.Delete(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete payment: %w", err)
	}

	return nil
}

func (s *PaymentService) GetPaymentsByInvoice(ctx context.Context, invoiceID uuid.UUID) ([]domain.Payment, error) {
	payments, err := s.repo.FindByInvoiceID(ctx, invoiceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get payments by invoice: %w", err)
	}

	return payments, nil
}

func (s *PaymentService) GetPaymentsByPartner(ctx context.Context, partnerID uuid.UUID) ([]domain.Payment, error) {
	payments, err := s.repo.FindByPartnerID(ctx, partnerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get payments by partner: %w", err)
	}

	return payments, nil
}

func (s *PaymentService) validatePayment(payment domain.Payment) error {
	if payment.OrganizationID == uuid.Nil {
		return fmt.Errorf("organization ID is required")
	}

	if payment.CompanyID == uuid.Nil {
		return fmt.Errorf("company ID is required")
	}

	if payment.InvoiceID == uuid.Nil {
		return fmt.Errorf("invoice ID is required")
	}

	if payment.PartnerID == uuid.Nil {
		return fmt.Errorf("partner ID is required")
	}

	if payment.CurrencyID == uuid.Nil {
		return fmt.Errorf("currency ID is required")
	}

	if payment.JournalID == uuid.Nil {
		return fmt.Errorf("journal ID is required")
	}

	if payment.Amount <= 0 {
		return fmt.Errorf("payment amount must be positive")
	}

	if payment.PaymentMethod == "" {
		return fmt.Errorf("payment method is required")
	}

	return nil
}
