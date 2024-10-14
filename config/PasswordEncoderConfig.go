package config

import (
	"github.com/ayushs-2k4/go-security/Auth/FromJava/PasswordEncoder"
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
