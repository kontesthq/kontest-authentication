package error

type UserAlreadyPresentError struct {
}

func (e *UserAlreadyPresentError) Error() string {
	return "User Already present"
}
