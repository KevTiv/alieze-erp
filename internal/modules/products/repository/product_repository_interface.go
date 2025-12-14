package repository

import (
	"context"

	"alieze-erp/internal/modules/products/domain"

	"github.com/google/uuid"
)

// ProductRepo defines the interface for product repository operations
type ProductRepo interface {
	Create(ctx context.Context, product domain.Product) (*domain.Product, error)
	FindByID(ctx context.Context, id uuid.UUID) (*domain.Product, error)
	FindAll(ctx context.Context, filter domain.ProductFilter) ([]domain.Product, error)
	Update(ctx context.Context, product domain.Product) (*domain.Product, error)
	Delete(ctx context.Context, id uuid.UUID) error
	Count(ctx context.Context, filter domain.ProductFilter) (int, error)
}
