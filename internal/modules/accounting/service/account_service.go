package service

import (
	"context"
	"fmt"

	"github.com/KevTiv/alieze-erp/internal/modules/accounting/repository"
	"github.com/KevTiv/alieze-erp/internal/modules/accounting/types"

	"github.com/google/uuid"
)

type AccountService struct {
	repo repository.AccountRepository
}

func NewAccountService(repo repository.AccountRepository) *AccountService {
	return &AccountService{
		repo: repo,
	}
}

func (s *AccountService) CreateAccount(ctx context.Context, account types.Account) (*types.Account, error) {
	// Validate the account
	if err := s.validateAccount(account); err != nil {
		return nil, fmt.Errorf("invalid account: %w", err)
	}

	// Check for duplicate code within organization
	existing, err := s.repo.FindByCode(ctx, account.OrganizationID, account.Code)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing account: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("account with code '%s' already exists", account.Code)
	}

	// Create the account
	createdAccount, err := s.repo.Create(ctx, account)
	if err != nil {
		return nil, fmt.Errorf("failed to create account: %w", err)
	}

	return createdAccount, nil
}

func (s *AccountService) GetAccount(ctx context.Context, id uuid.UUID) (*types.Account, error) {
	account, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get account: %w", err)
	}
	if account == nil {
		return nil, fmt.Errorf("account not found")
	}

	return account, nil
}

func (s *AccountService) ListAccounts(ctx context.Context, filters repository.AccountFilter) ([]types.Account, error) {
	accounts, err := s.repo.FindAll(ctx, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to list accounts: %w", err)
	}

	return accounts, nil
}

func (s *AccountService) UpdateAccount(ctx context.Context, account types.Account) (*types.Account, error) {
	// Validate the account
	if err := s.validateAccount(account); err != nil {
		return nil, fmt.Errorf("invalid account: %w", err)
	}

	// Check if account exists
	existing, err := s.repo.FindByID(ctx, account.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing account: %w", err)
	}
	if existing == nil {
		return nil, fmt.Errorf("account not found")
	}

	// Check for duplicate code if code is being changed
	if existing.Code != account.Code {
		duplicate, err := s.repo.FindByCode(ctx, account.OrganizationID, account.Code)
		if err != nil {
			return nil, fmt.Errorf("failed to check duplicate code: %w", err)
		}
		if duplicate != nil {
			return nil, fmt.Errorf("account with code '%s' already exists", account.Code)
		}
	}

	// Update the account
	updatedAccount, err := s.repo.Update(ctx, account)
	if err != nil {
		return nil, fmt.Errorf("failed to update account: %w", err)
	}

	return updatedAccount, nil
}

func (s *AccountService) DeleteAccount(ctx context.Context, id uuid.UUID) error {
	// Check if account exists
	existing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to check existing account: %w", err)
	}
	if existing == nil {
		return fmt.Errorf("account not found")
	}

	// Delete the account (soft delete)
	err = s.repo.Delete(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete account: %w", err)
	}

	return nil
}

func (s *AccountService) GetAccountsByType(ctx context.Context, organizationID uuid.UUID, accountType string) ([]types.Account, error) {
	accounts, err := s.repo.FindByType(ctx, organizationID, accountType)
	if err != nil {
		return nil, fmt.Errorf("failed to get accounts by type: %w", err)
	}

	return accounts, nil
}

func (s *AccountService) validateAccount(account types.Account) error {
	if account.OrganizationID == uuid.Nil {
		return fmt.Errorf("organization ID is required")
	}

	if account.Name == "" {
		return fmt.Errorf("name is required")
	}

	if account.Code == "" {
		return fmt.Errorf("code is required")
	}

	if account.AccountType == "" {
		return fmt.Errorf("account type is required")
	}

	// Validate account type is one of the valid types
	validTypes := map[string]bool{
		"receivable":      true,
		"payable":         true,
		"liquidity":       true,
		"other":           true,
		"regular":         true,
		"income":          true,
		"expense":         true,
		"depreciation":    true,
		"prepayments":     true,
		"equity":          true,
		"current_assets":  true,
		"non_current_assets": true,
		"current_liabilities": true,
		"non_current_liabilities": true,
	}

	if !validTypes[account.AccountType] {
		return fmt.Errorf("invalid account type: %s", account.AccountType)
	}

	return nil
}
