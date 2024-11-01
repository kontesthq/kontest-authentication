package ott

import (
	"github.com/kontesthq/go-security/Auth"
	"github.com/kontesthq/go-security/Auth/ott"
	error2 "github.com/kontesthq/go-security/examples/multi_ott_example/error"
	"kontest-authentication/service"
	"sync"
)

type LoginOneTimeTokenService struct {
	oneTimeTokenService ott.OneTimeTokenService
	once                sync.Once
}

// NewLoginOneTimeTokenService initializes and returns an instance of LoginOneTimeTokenService
func NewLoginOneTimeTokenService() *LoginOneTimeTokenService {
	return &LoginOneTimeTokenService{
		oneTimeTokenService: ott.NewInMemoryOneTimeTokenService(),
	}
}

func (o *LoginOneTimeTokenService) GenerateToken(user Auth.UserDetails) (*ott.OneTimeTokenAuthenticationToken, error) {
	return ott.GenerateOneTimeToken(user, o.oneTimeTokenService)
}

func (o *LoginOneTimeTokenService) DoAuthenticate(providedToken string) (string, error) {
	oneTimeToken := ott.NewOneTimeUnauthenticatedToken(providedToken)

	oneTimeTokenAuthenticationMethod := ott.NewOneTimeTokenAuthenticationMethod(*oneTimeToken, o.oneTimeTokenService, service.GetUserDetailsByEmail)

	authenticated, username, err := oneTimeTokenAuthenticationMethod.Authenticate()
	if err != nil {
		return "", err
	}

	if !authenticated {
		return "", &error2.WrongOTTError{}
	}

	return username, nil
}
