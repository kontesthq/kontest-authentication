package service

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"kontest-authentication/database"
	error2 "kontest-authentication/error"
	"kontest-authentication/model"
	"log"
)

type UserService struct {
}

func NewUserService() *UserService {
	return &UserService{}
}

func (us *UserService) Register(user model.User) (uuid.UUID, error) {
	// Begin a new transaction
	tx, err := database.GetDB().Beginx()
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to begin transaction: %v", err)
	}

	defer func() {
		// Ensure that the transaction is rolled back if not committed
		if err != nil {
			err := tx.Rollback()
			if err != nil {
				log.Println("Cannot rollback transaction")
				return
			}
		} else {
			if commitErr := tx.Commit(); commitErr != nil {
				log.Fatalf("Failed to commit transaction: %v", commitErr)
			}
		}
	}()

	dbUser, err := database.FindByEmail(user.Email)

	if (err != nil && !errors.Is(err, &error2.UserNotFoundError{})) {
		return uuid.Nil, fmt.Errorf("error checking if user exists: %v", err)
	}

	if dbUser != nil {
		return uuid.Nil, &error2.UserAlreadyPresentError{}
	}

	// Giving user a unique id
	uid, err := uuid.NewV7()

	if err != nil {
		return uuid.Nil, fmt.Errorf("cannot generate a unique ID")
	}

	user.ID = uid

	fmt.Println("Registering user")

	_, err = database.InsertUserIntoDB(user, tx)

	if err != nil {
		return uuid.Nil, err
	}

	_, err = database.AssignRoleToUser(user.ID, 1, tx)
	if err != nil {
		return uuid.Nil, errors.New("can not assign role to user")
	}

	return uid, nil
}
