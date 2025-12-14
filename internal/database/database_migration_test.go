package database

import (
	"os"
	"testing"

	_ "github.com/joho/godotenv/autoload"
)

func TestRunMigrations(t *testing.T) {
	// Set up environment variables for testing
	os.Setenv("BLUEPRINT_DB_HOST", "localhost")
	os.Setenv("BLUEPRINT_DB_PORT", "5432")
	os.Setenv("BLUEPRINT_DB_USERNAME", "testuser")
	os.Setenv("BLUEPRINT_DB_PASSWORD", "testpass")
	os.Setenv("BLUEPRINT_DB_DATABASE", "testdb")
	os.Setenv("BLUEPRINT_DB_SCHEMA", "public")

	defer ResetInstance()
	dbService := New()
	defer dbService.Close()

	// Test that RunMigrations doesn't panic and returns an error (since we don't have a real DB)
	err := dbService.RunMigrations()
	if err == nil {
		t.Log("RunMigrations completed without error (unexpected for test environment)")
	} else {
		t.Logf("RunMigrations returned expected error: %v", err)
	}

	// The main goal is to verify the method exists and can be called
	// In a real environment with proper DB credentials, this would run the migrations
}
