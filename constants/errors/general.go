package errors

import "errors"

var (
	ErrInternalServerError = errors.New("Internal server error")
	ErrSQLError            = errors.New("Database server failed to execute query")
	ErrTooManyRequests     = errors.New("Too many requests")
	ErrUnauthorized        = errors.New("Unauthorized")
	ErrInvalidToken        = errors.New("Invalid token")
	ErrForbidden           = errors.New("Forbidden")
	ErrInvalidUploadFile   = errors.New("Invalid upload file")
	ErrSizeTooBig          = errors.New("File size is too big")
)

var GeneralErrors = []error{
	ErrInternalServerError,
	ErrSQLError,
	ErrTooManyRequests,
	ErrUnauthorized,
	ErrInvalidToken,
	ErrForbidden,
	ErrInvalidUploadFile,
	ErrSizeTooBig,
}
