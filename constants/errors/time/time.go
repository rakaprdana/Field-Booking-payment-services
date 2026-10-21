package error

import "errors"

var (
	ErrTimeNotFound = errors.New("Time not found")
)

var TimeErrors = []error{
	ErrTimeNotFound,
}
