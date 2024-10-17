package model

import "github.com/google/uuid"

// Role represents a role in the application
type Role struct {
	ID      int         `db:"id"`       // Primary key
	Name    string      `db:"name"`     // Role name
	UserIDs []uuid.UUID `db:"user_ids"` // Many-to-many relationship with User (only IDs stored)
}

// Private variables (unexported)
var (
	roleUser = Role{
		ID:   1,
		Name: "USER",
	}

	roleAdmin = Role{
		ID:   2,
		Name: "ADMIN",
	}
)

// Getter functions to access roles

func GetRoleUser() Role {
	return roleUser
}

func GetRoleAdmin() Role {
	return roleAdmin
}
