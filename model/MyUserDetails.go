package model

import "github.com/google/uuid"

type UserDetails struct {
	UID      uuid.UUID
	Username string
	Password string
}

func (m *UserDetails) GetUsername() string {
	return m.Username
}

func (m *UserDetails) GetPassword() string {
	return m.Password
}

func (m *UserDetails) GetUID() uuid.UUID {
	return m.UID
}

func NewMyUserDetailsImpl(UID uuid.UUID, username, password string) *UserDetails {
	return &UserDetails{
		UID:      UID,
		Username: username,
		Password: password,
	}
}
