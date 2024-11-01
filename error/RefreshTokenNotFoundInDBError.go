package error

type RefreshTokenNotFoundInDBError struct {
}

func (e *RefreshTokenNotFoundInDBError) Error() string { return "Refresh Token not found in DB" }
