package service

import (
	"errors"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"kontest-authentication/database"
	"kontest-authentication/model"
	"log"
	"time"
)

type RefreshTokenService struct {
	refreshTokenValidity time.Duration
}

func NewRefreshTokenService() *RefreshTokenService {
	return &RefreshTokenService{
		refreshTokenValidity: 10 * time.Second,
	}
}

func (r *RefreshTokenService) generateNewToken() model.RefreshToken {
	refreshToken := model.RefreshToken{
		TokenID:      uuid.New(),
		RefreshToken: uuid.New().String(),
		Expiry:       time.Now().Add(r.refreshTokenValidity),
	}

	return refreshToken
}

func (r *RefreshTokenService) CreateRefreshToken(userId uuid.UUID, deviceID string) (*model.RefreshToken, error) {
	// Check if user exists in the database
	user, err := database.FindUserByID(userId)

	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, errors.New("user does not exist")
	}

	// Check if there is already a refresh token in the database for the user with the same device ID
	tx, err := database.GetDB().Beginx()
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			tx.Rollback() // rollback if there was an error
		} else {
			err = tx.Commit() // commit only if no error occurred
		}
	}()

	refreshToken, err := database.GetRefreshToken(userId, deviceID, tx)

	if err != nil {
		return nil, err
	}

	if refreshToken == nil {
		return r.createAndStoreRefreshToken(userId, deviceID, tx)
	} else if refreshToken.Expiry.Before(time.Now()) {
		// delete the old refresh token
		err := database.DeleteRefreshToken(refreshToken.TokenID, tx)

		if err != nil {
			log.Println("Failed to delete old refresh token")
		}

		return r.createAndStoreRefreshToken(userId, deviceID, tx)
	} else {
		return refreshToken, nil
	}
}

func (r *RefreshTokenService) createAndStoreRefreshToken(userId uuid.UUID, deviceID string, tx *sqlx.Tx) (*model.RefreshToken, error) {
	// Generate a new refresh token
	newRefreshToken := r.generateNewToken()
	newRefreshToken.UserID = userId
	newRefreshToken.AssociatedDeviceID = deviceID

	err := database.InsertDeviceIntoDBIfNotExists(model.Device{ID: deviceID, RefreshTokenID: newRefreshToken.TokenID}, tx)
	if err != nil {
		return nil, err
	}

	// Store the refresh token in the database
	_, err = database.InsertRefreshTokenIntoDB(newRefreshToken, tx)
	if err != nil {
		return nil, err
	}

	return &newRefreshToken, nil
}
