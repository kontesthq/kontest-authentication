package config

import (
	"github.com/kontesthq/go-security/Auth/PasswordEncoder"
	"log"
	"sync"
)

var (
	delegatingPasswordEncoder *PasswordEncoder.DelegatingPasswordEncoder
	once                      sync.Once
)

func initDelegatePasswordEncoder() {
	idForEncode := "argon2"
	encoders := PasswordEncoder.GetPasswordEncoders()
	var err error
	delegatingPasswordEncoder, err = PasswordEncoder.NewDelegatingPasswordEncoder(idForEncode, encoders)
	if err != nil {
		log.Fatalf("Error creating DelegatingPasswordEncoder: %s", err)
	}
}

func GetDelegatePasswordEncoder() *PasswordEncoder.DelegatingPasswordEncoder {
	once.Do(initDelegatePasswordEncoder)

	return delegatingPasswordEncoder
}
