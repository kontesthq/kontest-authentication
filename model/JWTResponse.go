package model

type JWTResponse struct {
	JWTToken     string `json:"jwt_token"`
	RefreshToken string `json:"refresh_token"`
	Username     string `json:"username"`
}
