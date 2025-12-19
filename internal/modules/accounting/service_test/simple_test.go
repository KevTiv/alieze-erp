package service_test

import (
	"testing"

	"github.com/KevTiv/alieze-erp/internal/modules/accounting/service"

	"github.com/stretchr/testify/assert"
)

func TestServiceStructsExist(t *testing.T) {
	// This is a simple compilation test to ensure the service structs exist
	var _ *service.InvoiceService = nil
	var _ *service.PaymentService = nil

	assert.True(t, true, "Service structs should exist")
}
