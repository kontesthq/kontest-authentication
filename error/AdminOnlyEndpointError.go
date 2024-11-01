package error

type AdminOnlyEndpointError struct {
}

func (e *AdminOnlyEndpointError) Error() string {
	return "Error: admin only endpoint"
}
