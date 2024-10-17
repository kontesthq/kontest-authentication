package model

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
