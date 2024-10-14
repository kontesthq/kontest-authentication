package database

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	error2 "kontest-authentication/error"
	"kontest-authentication/model"
	"log"
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
	var result sql.Result
	var err error

	if tx != nil {
		result, err = tx.NamedExec(`
		INSERT INTO refresh_tokens (token_id, refresh_token, expiry, user_id, associated_device_id)
		VALUES (:token_id, :refresh_token, :expiry, :user_id, :associated_device_id)`, &refreshToken)
	} else {
		result, err = db.NamedExec(`
		INSERT INTO refresh_tokens (token_id, refresh_token, expiry, user_id, associated_device_id)
		VALUES (:token_id, :refresh_token, :expiry, :user_id, :associated_device_id)`, &refreshToken)
	}

	if err != nil {
		return nil, fmt.Errorf("error adding refresh token for user %s: %v", refreshToken.UserID, err)
	}
	return result, nil
}

// InsertDeviceIntoDB Function to insert a device into the database
func InsertDeviceIntoDB(device model.Device, tx *sqlx.Tx) (sql.Result, error) {
	result, err := tx.NamedExec(`
		INSERT INTO devices (id)
		VALUES (:id)`, &device)
	if err != nil {
		return nil, fmt.Errorf("error adding device for refresh token ID %s: %v", device.RefreshTokenID, err)
	}
	return result, nil
}

func InsertDeviceIntoDBIfNotExists(device model.Device, tx *sqlx.Tx) error {
	// Check if device already exists
	existingDevice, err := FindDeviceByID(device.ID)

	if err != nil {
		if !errors.Is(err, &error2.DeviceNotFoundInDBError{}) {
			return fmt.Errorf("error checking if device exists: %v", err)
		}
	}

	if existingDevice == nil {
		_, err := InsertDeviceIntoDB(device, tx)
		if err != nil {
			return fmt.Errorf("error inserting device: %v", err)
		}
	}

	return nil
}

func FindDeviceByID(deviceID string) (*model.Device, error) {
	var device model.Device

	err := GetDB().Get(&device, `
	SELECT id FROM devices WHERE id = $1
`, deviceID)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &error2.DeviceNotFoundInDBError{}
		}

		return nil, fmt.Errorf("error fetching device: %v", err)
	}

	return &device, nil

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

func FindUserByEmail(email string) (*model.User, error) {
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

func FindUserByID(userID uuid.UUID) (*model.User, error) {
	var user model.User

	query := `SELECT * FROM users WHERE id = :id`

	// Use NamedQuery to retrieve the user by ID
	rows, err := GetDB().NamedQuery(query, map[string]interface{}{
		"id": userID,
	})

	if err != nil {
		return nil, fmt.Errorf("error executing query: %v", err)
	}
	defer rows.Close()

	// Check if any rows are returned
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

// GetRefreshToken retrieves a refresh token for a specified user ID and device ID using the provided transaction.
func GetRefreshToken(userID uuid.UUID, deviceID string, tx *sqlx.Tx) (*model.RefreshToken, error) {
	var refreshToken model.RefreshToken

	query := `
		SELECT *
		FROM refresh_tokens rt
		WHERE rt.user_id = $1 AND rt.associated_device_id = $2`

	// Use the transaction to execute the query
	rows, err := tx.Query(query, userID, deviceID)
	defer rows.Close()

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		log.Printf("Error retrieving refresh token for user ID %s and device ID %s: %v", userID.String(), deviceID, err)
		return nil, err
	}

	// Scan the result into the refreshToken struct
	if rows.Next() {
		// Scan the result into the refreshToken struct
		err = rows.Scan(
			&refreshToken.TokenID,
			&refreshToken.RefreshToken,
			&refreshToken.Expiry,
			&refreshToken.UserID,
			&refreshToken.AssociatedDeviceID,
		)
		if err != nil {
			log.Printf("Error scanning refresh token for user ID %s and device ID %s: %v", userID.String(), deviceID, err)
			return nil, err
		}
		return &refreshToken, nil
	}

	// If no rows were returned, the refresh token was not found
	return nil, nil
}

func DeleteRefreshToken(tokenID uuid.UUID, tx *sqlx.Tx) error {
	var err error

	// Use either the transaction or the default database connection
	if tx != nil {
		_, err = tx.Exec(`
			DELETE FROM refresh_tokens
			WHERE token_id = $1`, tokenID)
	} else {
		_, err = db.Exec(`
			DELETE FROM refresh_tokens
			WHERE token_id = $1`, tokenID)
	}

	// Handle any errors from the database operation
	if err != nil {
		return fmt.Errorf("error deleting refresh token with ID %s: %v", tokenID, err)
	}
	return nil
}

func GetRefreshTokenByRefreshToken(refreshToken string) (*model.RefreshToken, error) {
	var token model.RefreshToken

	query := `
	SELECT * FROM refresh_tokens
	WHERE refresh_token = $1`

	rows, err := db.Query(query, refreshToken)
	if err != nil {
		return nil, fmt.Errorf("error querying refresh token: %v", err)
	}
	defer rows.Close()

	// Check if a row was returned
	if rows.Next() {
		// Scan the result into the token struct
		if err := rows.Scan(&token.TokenID, &token.RefreshToken, &token.Expiry, &token.UserID, &token.AssociatedDeviceID); err != nil {
			return nil, fmt.Errorf("error scanning refresh token: %v", err)
		}
		return &token, nil
	}

	// No matching token found
	return nil, &error2.RefreshTokenNotFoundInDBError{}
}
