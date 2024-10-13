package database

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"log"
	"sync"
)

var (
	db   *sqlx.DB
	once sync.Once
)

// InitializeDatabase initializes the database connection.
func InitializeDatabase(dbName, dbPort, dbHost, dbUser, dbPassword, sslmode string) {
	var connStr string
	if dbPassword == "" {
		connStr = fmt.Sprintf("host=%s user=%s dbname=%s port=%s sslmode=%s",
			dbHost, dbUser, dbName, dbPort, sslmode)
	} else {
		connStr = fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
			dbHost, dbUser, dbPassword, dbName, dbPort, sslmode)
	}

	var err error
	once.Do(func() {
		db, err = sqlx.Connect("postgres", connStr)
		if err != nil {
			log.Fatalf("Failed to connect to database: %v", err)
		}
	})

	log.Println("Database connection established successfully.")
}

// GetDB returns the database connection.
func GetDB() *sqlx.DB {
	return db
}

// CloseDB closes the database connection.
func CloseDB() {
	if db != nil {
		err := db.Close()
		if err != nil {
			log.Printf("Error closing database: %v", err)
		} else {
			log.Println("Database connection closed.")
		}
	}
}

func SetupDatabase() {
	// Create the uuid-ossp extension
	if _, err := db.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\";"); err != nil {
		log.Fatalf("Failed to create uuid-ossp extension: %v", err)
	}

	createTables()
}

func createTables() {
	// Create Users table
	usersTable := `
	CREATE TABLE IF NOT EXISTS users (
		id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
		email VARCHAR(255) UNIQUE NOT NULL,
		password TEXT NOT NULL
	);`

	if _, err := db.Exec(usersTable); err != nil {
		log.Fatalf("Failed to create users table: %v", err)
	}

	// Create Roles table
	rolesTable := `
	CREATE TABLE IF NOT EXISTS roles (
		id SERIAL PRIMARY KEY,
		name VARCHAR(50) UNIQUE NOT NULL
	);`

	if _, err := db.Exec(rolesTable); err != nil {
		log.Fatalf("Failed to create roles table: %v", err)
	}

	// Create UserRoles (junction) table
	userRolesTable := `
	CREATE TABLE IF NOT EXISTS user_roles (
		user_id UUID NOT NULL,
		role_id INT NOT NULL,
		PRIMARY KEY (user_id, role_id),
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
		FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE
	);`

	if _, err := db.Exec(userRolesTable); err != nil {
		log.Fatalf("Failed to create user_roles table: %v", err)
	}

	// Create RefreshTokens table
	refreshTokensTable := `
	CREATE TABLE IF NOT EXISTS refresh_tokens (
		token_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
		refresh_token TEXT NOT NULL,
		expiry TIMESTAMP NOT NULL,
		user_id UUID NOT NULL,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);`

	if _, err := db.Exec(refreshTokensTable); err != nil {
		log.Fatalf("Failed to create refresh_tokens table: %v", err)
	}

	// Create Devices table
	devicesTable := `
	CREATE TABLE IF NOT EXISTS devices (
		id SERIAL PRIMARY KEY,
		refresh_token_id UUID NOT NULL,
		FOREIGN KEY (refresh_token_id) REFERENCES refresh_tokens(token_id) ON DELETE CASCADE
	);`

	if _, err := db.Exec(devicesTable); err != nil {
		log.Fatalf("Failed to create devices table: %v", err)
	}

	log.Println("All tables created successfully.")
}
