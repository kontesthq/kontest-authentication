package error

type RoleNotFoundError struct {
}

func (e *RoleNotFoundError) Error() string {
	return "Role not found"
}
