package database

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	error2 "kontest-authentication/error"
	"kontest-authentication/model"
)

// InsertUserIntoDB Function to insert a user into the database
func InsertUserIntoDB(user model.User, tx *sqlx.Tx) (sql.Result, error) {
	result, err := tx.NamedExec(`
		INSERT INTO users (id, email, password)
		VALUES (:id, :email, :password)`, &user)
	if err != nil {
		return nil, fmt.Errorf("error adding user %s: %v", user.Email, err)
	}

	return result, nil
}

// InsertRefreshTokenIntoDB Function to insert a refresh token into the database
func InsertRefreshTokenIntoDB(refreshToken model.RefreshToken, tx *sqlx.Tx) (sql.Result, error) {
	result, err := tx.NamedExec(`
		INSERT INTO refresh_tokens (token_id, refresh_token, expiry, user_id)
		VALUES (:token_id, :refresh_token, :expiry, :user_id)`, &refreshToken)
	if err != nil {
		return nil, fmt.Errorf("error adding refresh token for user %s: %v", refreshToken.UserID, err)
	}
	return result, nil
}

// InsertDeviceIntoDB Function to insert a device into the database
func InsertDeviceIntoDB(device model.Device, tx *sqlx.Tx) (sql.Result, error) {
	result, err := tx.NamedExec(`
		INSERT INTO devices (refresh_token_id)
		VALUES (:refresh_token_id)`, &device)
	if err != nil {
		return nil, fmt.Errorf("error adding device for refresh token ID %s: %v", device.RefreshTokenID, err)
	}
	return result, nil
}

// InsertRoleIntoDB Function to insert a role into the database
func InsertRoleIntoDB(role model.Role, tx *sqlx.Tx) (sql.Result, error) {
	result, err := tx.NamedExec(`
		INSERT INTO roles (id, name)
		VALUES (:id, :name)`, &role)
	if err != nil {
		return nil, fmt.Errorf("error adding role %s: %v", role.Name, err)
	}
	return result, nil
}

// FindRoleByName Function to find a role by its name
func FindRoleByName(roleName string) (*model.Role, error) {
	var role model.Role

	err := GetDB().Get(&role, `
	SELECT id, name FROM roles WHERE name = $1
`, roleName)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &error2.RoleNotFoundError{}
		}

		return nil, fmt.Errorf("error fetching role: %v", err)
	}

	return &role, nil
}

// AssignRoleToUser Function to assign a role to a user
func AssignRoleToUser(userID uuid.UUID, roleID int, tx *sqlx.Tx) (sql.Result, error) {
	result, err := tx.NamedExec(`
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

func FindByEmail(email string) (*model.User, error) {
	var user model.User

	query := `
	SELECT id, email, password FROM users WHERE email = :email`

	rows, err := GetDB().NamedQuery(query,
		map[string]interface{}{
			"email": email,
		})

	if err != nil {
		return nil, fmt.Errorf("error executing query: %v", err)
	}
	defer rows.Close()

	if rows.Next() {
		if err := rows.StructScan(&user); err != nil {
			return nil, fmt.Errorf("error scanning result: %v", err)
		}
		return &user, nil
	}

	// If no rows were returned, the user was not found
	return nil, &error2.UserNotFoundError{}
}

func UpdateUser(user model.User, tx *sqlx.Tx) error {
	_, err := tx.NamedExec(`
		UPDATE users
		SET password = :password
		WHERE id = :id`, &user)

	if err != nil {
		return fmt.Errorf("error updating details for user %s: %v", user.Email, err)
	}

	return nil
}
