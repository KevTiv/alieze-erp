package main

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

func main() {
	// Seed random number generator
	rand.Seed(time.Now().UnixNano())

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

	// Get test organization ID
	orgID := getTestOrganizationID(db)
	if orgID == uuid.Nil {
		log.Fatal("Test organization not found")
	}

	log.Printf("Using test organization: %s", orgID)

	// Generate test data
	log.Println("Generating test data...")

	// Create test products
	productIDs := generateTestProducts(db, orgID, 20)
	log.Printf("Created %d test products", len(productIDs))

	// Create test warehouses and locations
	warehouseID, locationIDs := generateTestInventoryStructure(db, orgID)
	log.Printf("Created warehouse and %d locations", len(locationIDs))

	// Create initial stock
	generateInitialStock(db, orgID, productIDs, locationIDs)
	log.Println("Created initial stock quantities")

	// Generate some inventory movements
	generateTestMovements(db, orgID, productIDs, locationIDs)
	log.Println("Generated test inventory movements")

	// Set reorder points for some products
	generateReorderPoints(db, orgID, productIDs)
	log.Println("Set reorder points for products")

	log.Println("Test data generation completed successfully!")
}

func getTestOrganizationID(db *sql.DB) uuid.UUID {
	query := `SELECT id FROM organizations WHERE code = 'TEST-ORG' LIMIT 1`
	var id uuid.UUID
	err := db.QueryRow(query).Scan(&id)
	if err != nil {
		log.Printf("Warning: Could not find test organization: %v", err)
		return uuid.Nil
	}
	return id
}

func generateTestProducts(db *sql.DB, orgID uuid.UUID, count int) []uuid.UUID {
	var productIDs []uuid.UUID

	productTypes := []string{"storable", "consumable", "service"}
	categories := []string{"Electronics", "Clothing", "Office", "Furniture", "Tools"}

	for i := 0; i < count; i++ {
		productID := uuid.New()
		productIDs = append(productIDs, productID)

		productType := productTypes[rand.Intn(len(productTypes))]
		category := categories[rand.Intn(len(categories))]
		listPrice := 10.0 + rand.Float64()*990.0
		standardPrice := listPrice * 0.6

		query := `
			INSERT INTO products (
				id, organization_id, name, default_code, product_type,
				list_price, standard_price, active, created_at, updated_at
			) VALUES (
				$1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW()
			) ON CONFLICT (id) DO NOTHING
		`

		productName := fmt.Sprintf("Test Product %d", i+1)
		defaultCode := fmt.Sprintf("PROD-%04d", i+1)

		_, err := db.Exec(query,
			productID, orgID, productName, defaultCode, productType,
			listPrice, standardPrice, true)
		if err != nil {
			log.Printf("Failed to create product %s: %v", productName, err)
		}
	}

	return productIDs
}

func generateTestInventoryStructure(db *sql.DB, orgID uuid.UUID) (uuid.UUID, []uuid.UUID) {
	// Create warehouse
	warehouseID := uuid.New()
	query := `
		INSERT INTO warehouses (
			id, organization_id, name, code, reception_steps, delivery_steps, active
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7
		) ON CONFLICT (id) DO NOTHING
	`
	_, err := db.Exec(query, warehouseID, orgID, "Test Warehouse", "TEST-WH", "one_step", "ship_only", true)
	if err != nil {
		log.Printf("Failed to create warehouse: %v", err)
	}

	// Create locations
	var locationIDs []uuid.UUID
	locationTypes := []string{"internal", "customer", "supplier", "inventory", "production"}

	for i := 0; i < 5; i++ {
		locationID := uuid.New()
		locationIDs = append(locationIDs, locationID)

		locationType := locationTypes[i%len(locationTypes)]
		locationName := fmt.Sprintf("Test Location %d", i+1)

		locQuery := `
			INSERT INTO stock_locations (
				id, organization_id, name, usage, removal_strategy, active
			) VALUES (
				$1, $2, $3, $4, $5, $6
			) ON CONFLICT (id) DO NOTHING
		`
		_, err := db.Exec(locQuery, locationID, orgID, locationName, locationType, "fifo", true)
		if err != nil {
			log.Printf("Failed to create location %s: %v", locationName, err)
		}
	}

	return warehouseID, locationIDs
}

func generateInitialStock(db *sql.DB, orgID uuid.UUID, productIDs []uuid.UUID, locationIDs []uuid.UUID) {
	for _, productID := range productIDs {
		for _, locationID := range locationIDs {
			// Only create stock for some combinations
			if rand.Float32() > 0.3 {
				quantity := float64(rand.Intn(100) + 1)

				query := `
					INSERT INTO stock_quants (
						id, organization_id, product_id, location_id, quantity,
						reserved_quantity, in_date, created_at, updated_at
					) VALUES (
						gen_random_uuid(), $1, $2, $3, $4, 0, NOW(), NOW(), NOW()
					) ON CONFLICT (product_id, location_id, organization_id)
					DO UPDATE SET quantity = EXCLUDED.quantity
				`

				_, err := db.Exec(query, orgID, productID, locationID, quantity)
				if err != nil {
					log.Printf("Failed to create stock for product %s in location %s: %v",
						productID, locationID, err)
				}
			}
		}
	}
}

func generateTestMovements(db *sql.DB, orgID uuid.UUID, productIDs []uuid.UUID, locationIDs []uuid.UUID) {
	if len(locationIDs) < 2 {
		return
	}

	for i := 0; i < 50; i++ {
		productID := productIDs[rand.Intn(len(productIDs))]
		sourceLoc := locationIDs[rand.Intn(len(locationIDs))]
		destLoc := locationIDs[rand.Intn(len(locationIDs))]

		// Ensure source and destination are different
		if sourceLoc == destLoc && len(locationIDs) > 1 {
			destLoc = locationIDs[(rand.Intn(len(locationIDs)-1)+1)%len(locationIDs)]
		}

		quantity := float64(rand.Intn(50) + 1)
		moveDate := time.Now().AddDate(0, 0, -rand.Intn(30))

		query := `
			INSERT INTO stock_moves (
				id, organization_id, product_id, location_id, location_dest_id,
				product_uom_qty, state, date, created_at, updated_at
			) VALUES (
				gen_random_uuid(), $1, $2, $3, $4, $5, 'done', $6, NOW(), NOW()
			)
		`

		_, err := db.Exec(query, orgID, productID, sourceLoc, destLoc, quantity, moveDate)
		if err != nil {
			log.Printf("Failed to create movement: %v", err)
		}
	}
}

func generateReorderPoints(db *sql.DB, orgID uuid.UUID, productIDs []uuid.UUID) {
	for _, productID := range productIDs {
		// Set reorder points for 50% of products
		if rand.Float32() > 0.5 {
			reorderPoint := float64(rand.Intn(50) + 10)
			safetyStock := reorderPoint * 0.3
			leadTime := rand.Intn(14) + 1

			query := `
				UPDATE products SET
					reorder_point = $1,
					safety_stock = $2,
					lead_time_days = $3
				WHERE id = $4 AND organization_id = $5
			`

			_, err := db.Exec(query, reorderPoint, safetyStock, leadTime, productID, orgID)
			if err != nil {
				log.Printf("Failed to set reorder points for product %s: %v", productID, err)
			}
		}
	}
}
