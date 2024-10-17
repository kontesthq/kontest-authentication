package model

import "github.com/google/uuid"

type DeleteUserRequest struct {
	UID uuid.UUID `json:"uid"` // UUID of the user
}
