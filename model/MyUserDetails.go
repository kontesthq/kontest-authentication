package model

//type UserDetails struct {
//	UID      uuid.UUID
//	Username string
//	Password string
//}
//
//func (m *UserDetails) GetUsername() string {
//	return m.Username
//}
//
//func (m *UserDetails) GetPassword() string {
//	return m.Password
//}
//
//func (m *UserDetails) GetUID() uuid.UUID {
//	return m.UID
//}
//
//func NewMyUserDetailsImpl(UID uuid.UUID, username, password string) *UserDetails {
//	return &UserDetails{
//		UID:      UID,
//		Username: username,
//		Password: password,
//	}
//}

type UserPrincipal struct {
	User User
}

func (u UserPrincipal) GetUsername() string {
	return u.User.Email
}

func (u UserPrincipal) GetPassword() string {
	return u.User.Password
}

func NewUserPrincipal(user User) *UserPrincipal {
	return &UserPrincipal{
		User: user,
	}
}
