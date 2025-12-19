package repository

import (
	"context"

	"github.com/KevTiv/alieze-erp/internal/modules/products/types"

	"github.com/google/uuid"
)

// ProductRepo defines the interface for product repository operations
type ProductRepo interface {
	Create(ctx context.Context, product types.Product) (*types.Product, error)
	FindByID(ctx context.Context, id uuid.UUID) (*types.Product, error)
	FindAll(ctx context.Context, filter types.ProductFilter) ([]types.Product, error)
	Update(ctx context.Context, product types.Product) (*types.Product, error)
	Delete(ctx context.Context, id uuid.UUID) error
	Count(ctx context.Context, filter types.ProductFilter) (int, error)
}
