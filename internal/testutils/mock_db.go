package testutils

import (
	"context"
	"database/sql"
	"log"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

// SetupMockDB creates a mock database connection for testing
type MockDB struct {
	DB   *sql.DB
	Mock sqlmock.Sqlmock
}

func SetupMockDB(t *testing.T) *MockDB {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}

	// No need to register pgx driver for sqlmock with pgx/v5
	// sqlmock works directly with database/sql interface

	return &MockDB{
		DB:   db,
		Mock: mock,
	}
}

func (m *MockDB) Close() {
	if err := m.DB.Close(); err != nil {
		log.Printf("Failed to close mock database: %v", err)
	}
}

// TestContext provides a context for testing
type TestContext struct {
	Context context.Context
	Cancel  func()
}

func NewTestContext() TestContext {
	ctx, cancel := context.WithCancel(context.Background())
	return TestContext{
		Context: ctx,
		Cancel:  cancel,
	}
}
