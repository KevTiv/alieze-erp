package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	_ "github.com/lib/pq"
)

func main() {
	// Get database connection string from environment variables
	dbHost := os.Getenv("BLUEPRINT_DB_HOST")
	dbPort := os.Getenv("BLUEPRINT_DB_PORT")
	dbUser := os.Getenv("BLUEPRINT_DB_USERNAME")
	dbPass := os.Getenv("BLUEPRINT_DB_PASSWORD")
	dbName := os.Getenv("BLUEPRINT_DB_DATABASE")

	if dbHost == "" || dbPort == "" || dbUser == "" || dbPass == "" || dbName == "" {
		log.Fatal("Missing database environment variables")
	}

	connString := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=disable",
		dbUser, dbPass, dbHost, dbPort, dbName)

	// Connect to database
	db, err := sql.Open("postgres", connString)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test connection
	err = db.Ping()
	if err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	log.Println("Connected to database successfully")

	// Create migrations table if it doesn't exist
	err = createMigrationsTable(db)
	if err != nil {
		log.Fatalf("Failed to create migrations table: %v", err)
	}

	// Get list of migration files
	migrationFiles, err := getMigrationFiles("internal/database/migrations")
	if err != nil {
		log.Fatalf("Failed to get migration files: %v", err)
	}

	// Get list of already applied migrations
	appliedMigrations, err := getAppliedMigrations(db)
	if err != nil {
		log.Fatalf("Failed to get applied migrations: %v", err)
	}

	log.Printf("Found %d migration files, %d already applied", len(migrationFiles), len(appliedMigrations))

	// Apply pending migrations
	pendingMigrations := getPendingMigrations(migrationFiles, appliedMigrations)
	if len(pendingMigrations) == 0 {
		log.Println("No pending migrations to apply")
		return
	}

	log.Printf("Applying %d pending migrations...", len(pendingMigrations))

	// Apply each pending migration
	for _, migration := range pendingMigrations {
		log.Printf("Applying migration: %s", migration)
		err = applyMigration(db, migration)
		if err != nil {
			log.Fatalf("Failed to apply migration %s: %v", migration, err)
		}
		log.Printf("Successfully applied migration: %s", migration)
	}

	log.Println("All migrations applied successfully!")
}

func createMigrationsTable(db *sql.DB) error {
	query := `
		CREATE TABLE IF NOT EXISTS migrations (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			UNIQUE(name)
		)
	`
	_, err := db.Exec(query)
	return err
}

func getMigrationFiles(migrationsDir string) ([]string, error) {
	var files []string

	err := filepath.Walk(migrationsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".sql") {
			files = append(files, path)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	// Sort files by name (which should contain timestamps)
	sort.Strings(files)
	return files, nil
}

func getAppliedMigrations(db *sql.DB) (map[string]bool, error) {
	query := `SELECT name FROM migrations ORDER BY applied_at`
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	applied := make(map[string]bool)
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		applied[name] = true
	}

	return applied, nil
}

func getPendingMigrations(allFiles []string, applied map[string]bool) []string {
	var pending []string
	for _, file := range allFiles {
		fileName := filepath.Base(file)
		if !applied[fileName] {
			pending = append(pending, file)
		}
	}
	return pending
}

func applyMigration(db *sql.DB, migrationFile string) error {
	// Read migration file
	content, err := os.ReadFile(migrationFile)
	if err != nil {
		return fmt.Errorf("failed to read migration file: %w", err)
	}

	// Execute migration
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Execute SQL statements
	statements := strings.Split(string(content), ";")
	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}

		_, err := tx.Exec(stmt)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to execute statement '%s': %w", stmt, err)
		}
	}

	// Record migration
	fileName := filepath.Base(migrationFile)
	_, err = tx.Exec("INSERT INTO migrations (name) VALUES ($1)", fileName)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to record migration: %w", err)
	}

	// Commit transaction
	return tx.Commit()
}
