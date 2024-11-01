package ott

import (
	"errors"
	"fmt"
	Auth2 "github.com/kontesthq/go-security/Auth"
	"github.com/kontesthq/go-security/Auth/ott"
	"kontest-authentication/Auth"
	error2 "kontest-authentication/error"
	"kontest-authentication/model"
	"kontest-authentication/service"
	"log/slog"
	"sync"
)

type OTTType int

const (
	Login = iota
	ForgotPassword
)

var (
	instance *OTTService
	once     sync.Once
)

type OTTService struct {
	loginOneTimeTokenService *LoginOneTimeTokenService
}

func NewOTTService() *OTTService {
	once.Do(func() {
		instance = &OTTService{
			loginOneTimeTokenService: NewLoginOneTimeTokenService(),
		}
	})

	return instance
}

func (o *OTTService) GenerateOTTByUser(user Auth2.UserDetails, ottType OTTType) (*ott.OneTimeTokenAuthenticationToken, error) {
	if user == nil {
		return nil, &error2.UserNotFoundError{}
	}

	switch ottType {
	case Login:
		token, err := o.loginOneTimeTokenService.GenerateToken(user)
		if err != nil {
			return nil, err
		}

		return token, nil
	case ForgotPassword:
		return nil, errors.New("forgot password not yet implemented")

	default:
		return nil, errors.New("invalid ott type")
	}
}

func (o *OTTService) GenerateOTTByUsername(username string, ottType OTTType) (*ott.OneTimeTokenAuthenticationToken, error) {
	user, err := service.GetUserDetailsByEmail(username)

	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, &error2.UserNotFoundError{}
	}

	return o.GenerateOTTByUser(user, ottType)
}

func (o *OTTService) ValidateOTT(providedToken string, ottType OTTType) (string, error) {
	switch ottType {
	case Login:
		return o.loginOneTimeTokenService.DoAuthenticate(providedToken)

	case ForgotPassword:
		return "", errors.New("forgot password not yet implemented")

	default:
		return "", errors.New("invalid ott type")
	}
}

func (o *OTTService) HandleLoginOTT(username string) error {
	user, err := service.GetUserByEmail(username)

	if err != nil {
		return err
	}

	if user == nil {
		return &error2.UserNotFoundError{}
	}

	oneTimeToken, err := o.GenerateOTTByUser(model.NewUserPrincipal(*user), Login)

	if err != nil {
		return err
	}

	err = service.PublishLoginOTTEmailEventToKafka(user.Email, oneTimeToken.GetTokenValue())
	if err != nil {
		slog.Error(fmt.Sprintf("Error publishing login ott email event to kafka: %v\n", err))
	} else {
		slog.Info(fmt.Sprintf("Login OTT email event published to kafka for user: %s\n", user.Email))
	}

	return nil
}

func (o *OTTService) ValidateLoginOTT(providedToken string) (string, string, error) {
	email, err := o.loginOneTimeTokenService.DoAuthenticate(providedToken)
	if err != nil {
		return "", "", err
	}

	user, err := service.GetUserByEmail(email)

	if err != nil {
		return "", "", err
	}

	if user == nil {
		return "", "", &error2.UserNotFoundError{}
	}

	return Auth.GenerateJWT(user.ID.String(), []byte(Auth.JWTSecret), Auth.JWTTokenExpiryDuration)
}
