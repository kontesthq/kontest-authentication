package error

type UserNotFoundError struct {
}

func (e *UserNotFoundError) Error() string {
	return "User not found"
}
