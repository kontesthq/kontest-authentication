package service

import (
	"errors"
	"fmt"
	"github.com/ayushs-2k4/go-security/Auth/FromJava/PasswordEncoder"
	"log"
	"net/http"
)

type UsernamePasswordAuthenticationMethod struct {
	Username                  string
	Password                  string
	DelegatingPasswordEncoder *PasswordEncoder.DelegatingPasswordEncoder
	GetUserDetails            func(username string) (*UserDetails, error)
	ChangePasswordFunc        func(username, newPassword string) error
}

func NewUsernamePasswordAuthenticationMethod(username, password string, delegatingPasswordEncoder *PasswordEncoder.DelegatingPasswordEncoder, getUserDetailsFunc func(username string) (*UserDetails, error), changePasswordFunc func(username, newPassword string) error) *UsernamePasswordAuthenticationMethod {

	if delegatingPasswordEncoder == nil {
		idForEncode := "scrypt"
		encoders := PasswordEncoder.GetPasswordEncoders()
		var err error
		delegatingPasswordEncoder, err = PasswordEncoder.NewDelegatingPasswordEncoder(idForEncode, encoders)
		if err != nil {
			log.Fatalf("Error creating DelegatingPasswordEncoder: %s", err)
		}
	}

	return &UsernamePasswordAuthenticationMethod{
		Username:                  username,
		Password:                  password,
		DelegatingPasswordEncoder: delegatingPasswordEncoder,
		GetUserDetails:            getUserDetailsFunc,
		ChangePasswordFunc:        changePasswordFunc,
	}
}

func (u *UsernamePasswordAuthenticationMethod) Authenticate(w http.ResponseWriter, r *http.Request) (bool, error) {
	inputUsername := u.Username
	inputPassword := u.Password

	// Call the custom authentication function
	if u.GetUserDetails != nil {
		user, err := u.GetUserDetails(inputUsername)

		if err != nil || user == nil {
			return false, err
		}

		dbPassword := user.GetPassword() // prefixEncodedPassword
		fmt.Println("dbPassword: " + dbPassword)

		// check if password matches
		if passwordMatches, err := u.DelegatingPasswordEncoder.Matches(inputPassword, dbPassword); err != nil || !passwordMatches {
			return false, errors.New("password is wrong")
		}

		// Authentication is successful
		shouldUpgradeEncoding := u.DelegatingPasswordEncoder.UpgradeEncoding(dbPassword)

		if shouldUpgradeEncoding {
			passwordWithNewEncoding, err := u.DelegatingPasswordEncoder.Encode(inputPassword)

			if err != nil {
				log.Printf("Cannot upgrade encoding due to error: %s\n", err)
			} else {
				log.Printf("Upgrading encoding for user: %s\n", user.GetUsername())
				u.ChangePasswordFunc(user.GetUsername(), passwordWithNewEncoding)
			}
		}

		return true, nil

	} else {
		return false, errors.New("no GetUserDetails function provided")
	}
}
