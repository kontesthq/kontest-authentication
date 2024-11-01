package service

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/kontesthq/go-security/Auth"
	"github.com/kontesthq/go-security/Auth/PasswordEncoder"
	"github.com/kontesthq/go-security/Auth/username_password"
	"kontest-authentication/config"
	"kontest-authentication/database"
	error2 "kontest-authentication/error"
	"kontest-authentication/model"
	"kontest-authentication/utils"
	"kontest-authentication/utils/spicedb_utils"
	"log"
	"log/slog"
	"regexp"
	"sync"
)

var (
	instance                  *UserService
	once                      sync.Once
	delegatingPasswordEncoder *PasswordEncoder.DelegatingPasswordEncoder
)

type UserService struct {
}

func NewUserService() *UserService {
	once.Do(func() {
		instance = &UserService{}
	})

	delegatingPasswordEncoder = config.GetDelegatePasswordEncoder()

	return instance
}

func GetUserDetailsByEmail(email string) (Auth.UserDetails, error) {
	user, err := GetUserByEmail(email)
	if err != nil {
		return nil, err
	}

	return model.NewUserPrincipal(*user), nil
}

func GetUserByEmail(email string) (*model.User, error) {
	user, err := database.FindUserByEmail(email)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func GetUserByUserID(userID string) (*model.User, error) {
	uid, err := utils.IsValidUUID(userID)
	if err != nil {
		return nil, err
	}

	user, err := database.FindUserByID(uid)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func changePassword(email, newPassword string) error {
	user, err := database.FindUserByEmail(email)
	if err != nil {
		return err
	}

	user.Password = newPassword

	tx, err := database.GetDB().Beginx()
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			tx.Rollback() // Rollback on error
		}
	}()

	err = database.UpdateUser(*user, tx)
	if err != nil {
		return err
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	return nil // Password change successful
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

	dbUser, err := database.FindUserByEmail(user.Email)

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
	// Encode the password before storing
	encodedPassword, err := delegatingPasswordEncoder.Encode(user.Password)
	if err != nil {
		// save simple password in case of error and log the error
		log.Printf("Error encoding password for the userID: %v with error: %v", user.ID, err)
		encodedPassword = user.Password
	}
	user.Password = encodedPassword // Set the encoded password back to the user model

	fmt.Println("Registering user")

	_, err = database.InsertUserIntoDB(user, tx)

	if err != nil {
		return uuid.Nil, err
	}

	_, err = database.AssignRoleToUser(user.ID, 1, tx)
	if err != nil {
		return uuid.Nil, errors.New("can not assign role to user")
	}

	// save user to spicedb
	err = spicedb_utils.SaveUserInSpiceDBWithDefaults(user.ID.String())
	if err != nil {
		return uuid.Nil, err
	}

	// Publish the registration message to kafka
	if err := PublishRegistrationMessage(user.Email); err != nil {
		log.Printf("Failed to publish registration message: %v", err)
	}

	return uid, nil
}

func getUserByEmail(email string) (model.User, error) {
	user, err := database.FindUserByEmail(email)
	if err != nil {
		return model.User{}, err
	}

	return *user, nil
}

func (us *UserService) GetUserByUserID(uid uuid.UUID) (*model.User, error) {
	user, err := database.FindUserByID(uid)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (us *UserService) DoAuthenticate(email, password string) (uuid.UUID, error) {
	usernamePasswordAuthenticationMethod := username_password.NewUsernamePasswordAuthenticationProvider(delegatingPasswordEncoder, true, GetUserDetailsByEmail, changePassword)

	usernamePasswordToken := username_password.NewUsernamePasswordAuthenticationToken(email, password)

	authentication, err := usernamePasswordAuthenticationMethod.Authenticate(usernamePasswordToken)
	if err != nil || !authentication.IsAuthenticated() {
		log.Printf("Authentication failed with error: %s", err)
		return uuid.Nil, err
	}

	userDetails, err := getUserByEmail(email)

	if err != nil {
		log.Printf("Error getting user details: %s", err)
		return uuid.Nil, err
	}

	fmt.Printf("authenticated: %v\n", authentication.IsAuthenticated())

	return userDetails.ID, nil
}

// IsValidDeviceID validates the device ID format
func (us *UserService) IsValidDeviceID(deviceID string) bool {
	// Regular expression to check if the device ID is a 128-character hex string
	regex := regexp.MustCompile("^[a-fA-F0-9]{128}$")

	// Check if deviceID is not nil and matches the pattern
	return deviceID != "" && regex.MatchString(deviceID)
}

func (us *UserService) MakeUserMember(uid uuid.UUID) (bool, error) {
	user, err := database.FindUserByID(uid)

	if err != nil {
		return false, fmt.Errorf("failed to find user by ID: %w", err)
	}
	if user == nil {
		return false, &error2.UserNotFoundError{}
	}

	tx := database.GetDB().MustBegin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			slog.Error(fmt.Sprintf("Panic occurred: %v", r))
		} else if err != nil {
			tx.Rollback()
		} else {
			if commitErr := tx.Commit(); commitErr != nil {
				slog.Error(fmt.Sprintf("Failed to commit transaction: %v", commitErr))
			}
		}
	}()

	err = us.updateUserRoles(uid, []int{model.GetRoleUser().ID}, tx)
	if err != nil {
		return false, err
	}

	// Changing roles in spicedb
	err = spicedb_utils.MakeUserMember(uid.String())
	if err != nil {
		return false, err
	}

	slog.Info(fmt.Sprintf("User with uid %s has been made member", uid.String()))
	return true, nil
}

func (us *UserService) MakeUserAdmin(uid uuid.UUID) (bool, error) {
	user, err := database.FindUserByID(uid)

	if err != nil {
		return false, fmt.Errorf("failed to find user by ID: %w", err)
	}
	if user == nil {
		return false, &error2.UserNotFoundError{}
	}

	tx := database.GetDB().MustBegin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			slog.Error(fmt.Sprintf("Panic occurred: %v", r))
		} else if err != nil {
			tx.Rollback()
		} else {
			if commitErr := tx.Commit(); commitErr != nil {
				slog.Error(fmt.Sprintf("Failed to commit transaction: %v", commitErr))
			}
		}
	}()

	// Updating roles in DB
	err = us.updateUserRoles(uid, []int{model.GetRoleUser().ID, model.GetRoleAdmin().ID}, tx)
	if err != nil {
		return false, err
	}

	// Updating roles in spicedb
	err = spicedb_utils.MakeUserAdmin(uid.String())
	if err != nil {
		return false, err
	}

	slog.Info(fmt.Sprintf("User with uid %s has been made an admin", uid.String()))
	return true, nil
}

func (us *UserService) updateUserRoles(uid uuid.UUID, roleIDs []int, tx *sqlx.Tx) error {
	_, err := database.UpdateUserRoles(uid, roleIDs, tx)
	if err != nil {
		return fmt.Errorf("failed to update roles for user %s: %v", uid, err)
	}
	return nil
}

func (us *UserService) DeleteUser(uid uuid.UUID) (bool, error) {
	user, err := database.FindUserByID(uid)

	if err != nil || user == nil {
		return false, &error2.UserNotFoundError{}
	}

	hasUserDeleted, err := database.DeleteUser(uid, nil)

	if err != nil || !hasUserDeleted {
		return false, errors.New("can not delete user")
	}

	// Delete from spicedb
	spicedb_utils.DeleteUserFromSpiceDB(uid.String())

	// Publish the deletion message to kafka
	if err := PublishAccountDeletionEventToKafka(user.Email); err != nil {
		slog.Error(fmt.Sprintf("failed to publish registration message: %v", err))
	}

	return true, nil
}
