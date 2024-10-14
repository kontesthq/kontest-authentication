package error

type DeviceNotFoundInDBError struct {
}

func (e *DeviceNotFoundInDBError) Error() string { return "Device not found in DB" }
