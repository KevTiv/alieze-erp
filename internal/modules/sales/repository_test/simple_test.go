package repository_test

import (
	"testing"

	"github.com/KevTiv/alieze-erp/internal/modules/sales/repository"

	"github.com/stretchr/testify/assert"
)

func TestRepositoryInterfacesExist(t *testing.T) {
	// This is a simple compilation test to ensure the repository interfaces exist
	var _ repository.SalesOrderRepository = nil
	var _ repository.PricelistRepository = nil

	assert.True(t, true, "Repository interfaces should exist")
}
