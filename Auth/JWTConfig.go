package Auth

import (
	"os"
	"time"
)

// JWTConfig holds the configuration for JWT
var (
	JWTSecret              = getJWTSecret()  // Get JWT secret from environment variable
	JWTTokenExpiryDuration = 5 * time.Minute // Set token expiration duration
)

// getJWTSecret retrieves the JWT secret from the environment variable
func getJWTSecret() string {
	secret := os.Getenv("KONTEST_JWT_SECRET")
	if secret == "" {
		secret = "JWT_SECRET"
	}
	return secret
}
