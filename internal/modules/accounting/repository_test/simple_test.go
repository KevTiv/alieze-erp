package repository_test

import (
	"testing"

	"github.com/KevTiv/alieze-erp/internal/modules/accounting/repository"

	"github.com/stretchr/testify/assert"
)

func TestRepositoryInterfacesExist(t *testing.T) {
	// This is a simple compilation test to ensure the repository interfaces exist
	var _ repository.InvoiceRepository = nil
	var _ repository.PaymentRepository = nil

	assert.True(t, true, "Repository interfaces should exist")
}
