package database

import (
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	error2 "kontest-authentication/error"
	"kontest-authentication/model"
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

	insertRoles()
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

	// Create Devices table
	devicesTable := `
	CREATE TABLE IF NOT EXISTS devices (
		id TEXT PRIMARY KEY
	);`

	if _, err := db.Exec(devicesTable); err != nil {
		log.Fatalf("Failed to create devices table: %v", err)
	}

	// Create RefreshTokens table
	refreshTokensTable := `
	CREATE TABLE IF NOT EXISTS refresh_tokens (
		token_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
		refresh_token TEXT NOT NULL,
		expiry TIMESTAMPTZ NOT NULL,  -- Use TIMESTAMPTZ to include timezone info
		user_id UUID NOT NULL,
		associated_device_id TEXT NOT NULL,
		FOREIGN KEY (associated_device_id) REFERENCES devices(id) ON DELETE CASCADE,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);`

	if _, err := db.Exec(refreshTokensTable); err != nil {
		log.Fatalf("Failed to create refresh_tokens table: %v", err)
	}

	log.Println("All tables created successfully.")
}

func insertRoles() {
	roles := []model.Role{
		model.GetRoleUser(),
		model.GetRoleAdmin(),
	}

	db := GetDB()          // Get the database connection
	tx, err := db.Beginx() // Begin a new transaction
	if err != nil {
		log.Fatalf("Failed to begin transaction: %v", err)
	}

	defer func() {
		if err != nil {
			err := tx.Rollback()
			if err != nil {
				log.Println("Cannot rollback transaction")
				return
			} // Rollback if there was an error
		} else {
			if commitErr := tx.Commit(); commitErr != nil {
				log.Fatalf("Failed to commit transaction: %v", commitErr)
			}
		}
	}()

	for _, role := range roles {
		if err := insertRoleIfNotExists(role, tx); err != nil {
			log.Fatalf("Can not insert role %s in DB with error: %s", role.Name, err)
		}
	}

	log.Println("Roles successfully added to DB")
}

func insertRoleIfNotExists(role model.Role, tx *sqlx.Tx) error {
	// Check if role already exists
	existingRole, err := FindRoleByName(role.Name)

	if err != nil && !errors.Is(err, &error2.RoleNotFoundError{}) {
		return fmt.Errorf("error checking if role exists: %v", err)
	}

	if existingRole != nil {
		log.Printf("Role %s already exists in DB", role.Name)
		return nil // Role already exists, do nothing
	}

	// Insert the role if it does not exist
	_, err = InsertRoleIntoDB(role, tx)
	if err != nil {
		return fmt.Errorf("failed to insert role %s: %w", role.Name, err)
	}

	return nil
}
