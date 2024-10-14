package Auth

import (
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"time"
)

func GenerateJWT(subject string, secret []byte, expiry time.Duration) (string, string, error) {
	// Generate JWT
	jwtToken, err := GenerateJWTOnly(subject, secret, expiry)
	if err != nil {
		return "", "", err
	}

	// Generate Refresh Token
	refreshToken, err := GenerateRefreshToken()
	if err != nil {
		return "", "", err
	}

	return jwtToken, refreshToken, nil
}

func GenerateJWTOnly(subject string, secret []byte, expiry time.Duration) (string, error) {
	expirationTime := time.Now().Add(expiry)

	claims := &jwt.StandardClaims{
		ExpiresAt: expirationTime.Unix(),
		Subject:   subject,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(secret)

	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// GenerateRefreshToken generates a simple UUID as a refresh token.
func GenerateRefreshToken() (string, error) {
	// Generate a New UUID for the refresh token
	refreshToken := uuid.New().String()

	// Store the refresh token along with the associated subject

	return refreshToken, nil
}
