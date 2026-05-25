package response

import "errors"

// Common domain errors that your services or repositories can return
var (
	ErrNotFound       = errors.New("requested resource not found")
	ErrUnauthorized   = errors.New("invalid or missing authentication credentials")
	ErrAlreadyExists  = errors.New("resource with these details already exists")
	ErrInvalidInput   = errors.New("provided request payload is malformed or invalid")
	ErrInternalServer = errors.New("internal processing error")
)