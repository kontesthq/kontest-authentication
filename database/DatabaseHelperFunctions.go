package database

import (
	"database/sql"
	"fmt"
	"github.com/google/uuid"
	"kontest-authentication/model"
)

// InsertUserIntoDB Function to insert a user into the database
func InsertUserIntoDB(user model.User) (sql.Result, error) {
	result, err := GetDB().NamedExec(`
		INSERT INTO users (id, email, password)
		VALUES (:id, :email, :password)`, &user)
	if err != nil {
		return nil, fmt.Errorf("error adding user %s: %v", user.Email, err)
	}

	return result, nil
}

// InsertRefreshTokenIntoDB Function to insert a refresh token into the database
func InsertRefreshTokenIntoDB(refreshToken model.RefreshToken) (sql.Result, error) {
	result, err := GetDB().NamedExec(`
		INSERT INTO refresh_tokens (token_id, refresh_token, expiry, user_id)
		VALUES (:token_id, :refresh_token, :expiry, :user_id)`, &refreshToken)
	if err != nil {
		return nil, fmt.Errorf("error adding refresh token for user %s: %v", refreshToken.UserID, err)
	}
	return result, nil
}

// InsertDeviceIntoDB Function to insert a device into the database
func InsertDeviceIntoDB(device model.Device) (sql.Result, error) {
	result, err := GetDB().NamedExec(`
		INSERT INTO devices (refresh_token_id)
		VALUES (:refresh_token_id)`, &device)
	if err != nil {
		return nil, fmt.Errorf("error adding device for refresh token ID %s: %v", device.RefreshTokenID, err)
	}
	return result, nil
}

// InsertRoleIntoDB Function to insert a role into the database
func InsertRoleIntoDB(role model.Role) (sql.Result, error) {
	result, err := GetDB().NamedExec(`
		INSERT INTO roles (id, name)
		VALUES (:id, :name)`, &role)
	if err != nil {
		return nil, fmt.Errorf("error adding role %s: %v", role.Name, err)
	}
	return result, nil
}

// AssignRoleToUser Function to assign a role to a user
func AssignRoleToUser(userID uuid.UUID, roleID int) (sql.Result, error) {
	result, err := GetDB().NamedExec(`
		INSERT INTO user_roles (user_id, role_id)
		VALUES (:user_id, :role_id)`, map[string]interface{}{
		"user_id": userID,
		"role_id": roleID,
	})
	if err != nil {
		return nil, fmt.Errorf("error assigning role ID %d to user ID %s: %v", roleID, userID, err)
	}
	return result, nil
}
