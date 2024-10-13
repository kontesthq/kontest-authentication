package model

import "github.com/google/uuid"

// User represents a user in the application
type User struct {
	ID              uuid.UUID   `db:"id"`                // UUID as primary key
	Email           string      `db:"email"`             // Unique email address
	Password        string      `db:"password"`          // Password (hashed)
	DeviceIDs       []uuid.UUID `db:"device_ids"`        // One-to-many relationship with Device (only IDs stored)
	RoleIDs         []int       `db:"role_ids"`          // Many-to-many relationship with Role
	RefreshTokenIDs []uuid.UUID `db:"refresh_token_ids"` // Many-to-one relationship with RefreshToken (only IDs stored)
}
