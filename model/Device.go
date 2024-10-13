package model

import "github.com/google/uuid"

// Device represents a device in the application
type Device struct {
	ID             string    `db:"id"`               // Unique identifier for the device
	RefreshTokenID uuid.UUID `db:"refresh_token_id"` // Foreign key to RefreshToken
}
