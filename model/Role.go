package model

import "github.com/google/uuid"

// Role represents a role in the application
type Role struct {
	ID      int         `db:"id"`       // Primary key
	Name    string      `db:"name"`     // Role name
	UserIDs []uuid.UUID `db:"user_ids"` // Many-to-many relationship with User (only IDs stored)
}
