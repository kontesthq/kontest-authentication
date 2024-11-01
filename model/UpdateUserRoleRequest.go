package model

import "github.com/google/uuid"

type UpdateUserRoleRequest struct {
	UID uuid.UUID `json:"uid"` // UUID of the user
}
