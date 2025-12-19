package testutils

import (
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// MockDB wraps a real database connection with sqlmock for testing
type MockDB struct {
	DB   *sql.DB
	Mock sqlmock.Sqlmock
}

// SetupMockDB creates a new mock database connection with sqlmock
func SetupMockDB(t *testing.T) *MockDB {
	t.Helper()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}

	return &MockDB{
		DB:   db,
		Mock: mock,
	}
}

// NewMockDB creates a new mock database connection with sqlmock
func NewMockDB() *MockDB {
	db, mock, err := sqlmock.New()
	if err != nil {
		panic("Failed to create mock database: " + err.Error())
	}

	return &MockDB{
		DB:   db,
		Mock: mock,
	}
}

// Close closes the mock database connection
func (m *MockDB) Close() {
	if m.DB != nil {
		m.DB.Close()
	}
	if m.Mock != nil {
		m.Mock = nil
	}
}
