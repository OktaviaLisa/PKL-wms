package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

var DB *sql.DB

func Connect(databaseURL string) error {
	var err error
	for i := 0; i < 30; i++ {
		DB, err = sql.Open("postgres", databaseURL)
		if err != nil {
			log.Printf("Failed to connect to database (attempt %d): %v", i+1, err)
			time.Sleep(2 * time.Second)
			continue
		}

		if err = DB.Ping(); err != nil {
			log.Printf("Failed to ping database (attempt %d): %v", i+1, err)
			time.Sleep(2 * time.Second)
			continue
		}
		break
	}

	if err != nil {
		return fmt.Errorf("could not connect to database after 30 attempts: %v", err)
	}

	fmt.Println("Connected to database successfully")
	return createTables()
}

func createTables() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS auth_user (
			id SERIAL PRIMARY KEY,
			username VARCHAR(150) UNIQUE NOT NULL,
			email VARCHAR(254) NOT NULL,
			password VARCHAR(128) NOT NULL,
			is_staff BOOLEAN DEFAULT FALSE,
			is_superuser BOOLEAN DEFAULT FALSE,
			is_active BOOLEAN DEFAULT TRUE,
			first_name VARCHAR(150) DEFAULT '',
			last_name VARCHAR(150) DEFAULT '',
			roles TEXT DEFAULT 'user',
			user_type VARCHAR(50) DEFAULT 'user',
			company_name VARCHAR(200) DEFAULT '',
			warehouse_id INTEGER DEFAULT NULL,
			date_joined TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS warehouses (
			id SERIAL PRIMARY KEY,
			name VARCHAR(200) NOT NULL,
			location VARCHAR(300) DEFAULT '',
			manager_id INTEGER DEFAULT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS warehouse_category (
			id SERIAL PRIMARY KEY,
			name VARCHAR(100) NOT NULL,
			description TEXT DEFAULT '',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS warehouse_product (
			id SERIAL PRIMARY KEY,
			name VARCHAR(200) NOT NULL,
			sku VARCHAR(50) UNIQUE NOT NULL,
			category_id INTEGER REFERENCES warehouse_category(id),
			description TEXT DEFAULT '',
			price DECIMAL(10,2) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS inbound_requests (
			id SERIAL PRIMARY KEY,
			product_id INTEGER REFERENCES warehouse_product(id),
			quantity INTEGER NOT NULL,
			supplier VARCHAR(200) NOT NULL,
			status VARCHAR(50) DEFAULT 'pending',
			notes TEXT DEFAULT '',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS outbound_requests (
			id SERIAL PRIMARY KEY,
			product_id INTEGER REFERENCES warehouse_product(id),
			quantity INTEGER NOT NULL,
			destination VARCHAR(200) NOT NULL,
			status VARCHAR(50) DEFAULT 'pending',
			notes TEXT DEFAULT '',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS orders (
			id SERIAL PRIMARY KEY,
			order_number VARCHAR(100) UNIQUE NOT NULL,
			customer VARCHAR(200) NOT NULL,
			status VARCHAR(50) DEFAULT 'pending',
			total_amount DECIMAL(10,2) DEFAULT 0,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS inventory (
			id SERIAL PRIMARY KEY,
			product_id INTEGER REFERENCES warehouse_product(id),
			quantity INTEGER DEFAULT 0,
			min_stock INTEGER DEFAULT 0,
			location_id INTEGER REFERENCES locations(id) DEFAULT 1,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(product_id, location_id)
		)`,
		`CREATE TABLE IF NOT EXISTS stock_movements (
			id SERIAL PRIMARY KEY,
			product_id INTEGER REFERENCES warehouse_product(id),
			movement_type VARCHAR(50) NOT NULL,
			quantity INTEGER NOT NULL,
			reference VARCHAR(200) DEFAULT '',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS suppliers (
			id SERIAL PRIMARY KEY,
			name VARCHAR(200) NOT NULL,
			contact_person VARCHAR(200) DEFAULT '',
			phone VARCHAR(50) DEFAULT '',
			email VARCHAR(200) DEFAULT '',
			address TEXT DEFAULT '',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS customers (
			id SERIAL PRIMARY KEY,
			name VARCHAR(200) NOT NULL,
			contact_person VARCHAR(200) DEFAULT '',
			phone VARCHAR(50) DEFAULT '',
			email VARCHAR(200) DEFAULT '',
			address TEXT DEFAULT '',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS units (
			id SERIAL PRIMARY KEY,
			name VARCHAR(100) NOT NULL,
			symbol VARCHAR(20) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS locations (
			id SERIAL PRIMARY KEY,
			name VARCHAR(200) NOT NULL,
			code VARCHAR(50) NOT NULL,
			description TEXT DEFAULT '',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS receiving (
			id SERIAL PRIMARY KEY,
			document_number VARCHAR(100) UNIQUE NOT NULL,
			receive_date DATE NOT NULL,
			supplier_id INTEGER REFERENCES suppliers(id),
			product_id INTEGER REFERENCES warehouse_product(id),
			quantity INTEGER NOT NULL,
			unit_id INTEGER REFERENCES units(id),
			location_id INTEGER REFERENCES locations(id),
			remarks TEXT DEFAULT '',
			status VARCHAR(50) DEFAULT 'pending',
			created_by INTEGER DEFAULT 1,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS issuing (
			id SERIAL PRIMARY KEY,
			document_number VARCHAR(100) UNIQUE NOT NULL,
			issue_date DATE NOT NULL,
			customer_id INTEGER REFERENCES customers(id),
			product_id INTEGER REFERENCES warehouse_product(id),
			quantity INTEGER NOT NULL,
			unit_id INTEGER REFERENCES units(id),
			location_id INTEGER REFERENCES locations(id),
			remarks TEXT DEFAULT '',
			status VARCHAR(50) DEFAULT 'pending',
			created_by INTEGER DEFAULT 1,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS receptions (
			id SERIAL PRIMARY KEY,
			product_name VARCHAR(200) NOT NULL,
			category VARCHAR(100),
			quantity INTEGER NOT NULL,
			location VARCHAR(200),
			notes TEXT,
			received_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			status VARCHAR(50) DEFAULT 'quality_check'
		)`,
		`CREATE TABLE IF NOT EXISTS quality_checks (
			id SERIAL PRIMARY KEY,
			reception_id INTEGER,
			product_name VARCHAR(200),
			quantity INTEGER,
			status VARCHAR(50),
			notes TEXT,
			checked_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS returns (
			id SERIAL PRIMARY KEY,
			product_name VARCHAR(255) NOT NULL,
			quantity INTEGER NOT NULL,
			return_type VARCHAR(50) NOT NULL DEFAULT 'QUALITY_FAIL',
			reason TEXT,
			status VARCHAR(50) NOT NULL DEFAULT 'PENDING',
			supplier VARCHAR(255),
			reception_id INTEGER,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS dispatches (
			id SERIAL PRIMARY KEY,
			product_name VARCHAR(255) NOT NULL,
			customer VARCHAR(255),
			quantity INTEGER NOT NULL,
			location VARCHAR(255),
			notes TEXT,
			dispatch_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			status VARCHAR(50) DEFAULT 'pending'
		)`,

	}

	for _, query := range queries {
		if _, err := DB.Exec(query); err != nil {
			log.Printf("Error creating table: %v", err)
		}
	}

	return createDefaultData()
}

func createDefaultData() error {
	// Create default admin
	if err := createDefaultAdmin(); err != nil {
		return err
	}
	
	// Create default master data
	return createDefaultMasterData()
}

func createDefaultAdmin() error {
	// Create admin user
	var count int
	err := DB.QueryRow("SELECT COUNT(*) FROM auth_user WHERE username = 'admin'").Scan(&count)
	if err == nil && count == 0 {
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("admin"), bcrypt.DefaultCost)
		
		_, err = DB.Exec(`
			INSERT INTO auth_user (username, email, password, is_staff, is_superuser, is_active, first_name, last_name, roles, date_joined) 
			VALUES ('admin', 'admin@admin.com', $1, true, true, true, 'Admin', 'User', 'admin', NOW())
		`, string(hashedPassword))
		
		if err != nil {
			return fmt.Errorf("error creating admin user: %v", err)
		}
		
		fmt.Println("Default admin user created (admin/admin)")
	}

	// Create QC Inspector user
	err = DB.QueryRow("SELECT COUNT(*) FROM auth_user WHERE username = 'qc_inspector'").Scan(&count)
	if err == nil && count == 0 {
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("qc123"), bcrypt.DefaultCost)
		
		_, err = DB.Exec(`
			INSERT INTO auth_user (username, email, password, is_staff, is_superuser, is_active, first_name, last_name, roles, date_joined) 
			VALUES ('qc_inspector', 'qc@wms.com', $1, false, false, true, 'QC', 'Inspector', 'qc_inspector', NOW())
		`, string(hashedPassword))
		
		if err != nil {
			return fmt.Errorf("error creating qc_inspector user: %v", err)
		}
		
		fmt.Println("Default QC Inspector user created (qc_inspector/qc123)")
	}

	return nil
}

func createDefaultMasterData() error {
	// Insert default units
	units := [][]string{
		{"Pieces", "pcs"},
		{"Box", "box"},
		{"Kilogram", "kg"},
		{"Liter", "ltr"},
	}
	
	for _, unit := range units {
		_, err := DB.Exec(`INSERT INTO units (name, symbol) VALUES ($1, $2) ON CONFLICT DO NOTHING`, unit[0], unit[1])
		if err != nil {
			log.Printf("Error inserting unit %s: %v", unit[0], err)
		}
	}
	
	// Insert default locations
	locations := [][]string{
		{"Warehouse A", "WH-A", "Main warehouse storage"},
		{"Warehouse B", "WH-B", "Secondary warehouse storage"},
		{"Cold Storage", "CS-01", "Temperature controlled storage"},
	}
	
	for _, location := range locations {
		_, err := DB.Exec(`INSERT INTO locations (name, code, description) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING`, location[0], location[1], location[2])
		if err != nil {
			log.Printf("Error inserting location %s: %v", location[0], err)
		}
	}
	
	// Insert default category
	_, err := DB.Exec(`INSERT INTO warehouse_category (name, description) VALUES ('Electronics', 'Electronic items') ON CONFLICT DO NOTHING`)
	if err != nil {
		log.Printf("Error inserting category: %v", err)
	}
	
	fmt.Println("Default master data created")
	return nil
}