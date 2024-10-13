package model

import (
	"github.com/google/uuid"
	"time"
)

// RefreshToken represents a refresh token in the application
type RefreshToken struct {
	TokenID            uuid.UUID `db:"token_id"`             // UUID as primary key
	RefreshToken       string    `db:"refresh_token"`        // The actual refresh token string
	Expiry             time.Time `db:"expiry"`               // Expiration time for the refresh token
	UserID             uuid.UUID `db:"user_id"`              // Foreign key to User
	AssociatedDeviceID string    `db:"associated_device_id"` // Foreign key to Device
}
