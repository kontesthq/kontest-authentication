package model

type JWTRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`

	DeviceID string `json:"device_id"`
}
