package utils

import "github.com/google/uuid"

func IsValidUUID(uuidStr string) (uuid.UUID, error) {
	return uuid.Parse(uuidStr)
}
