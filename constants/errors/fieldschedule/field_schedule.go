package error

import "errors"

var (
	ErrFieldScheduleNotFound = errors.New("Field schedule not found")
	ErrFieldScheduleIsExist  = errors.New("Field schedule already exist")
)

var FieldScheduleErrors = []error{
	ErrFieldScheduleNotFound,
	ErrFieldScheduleIsExist,
}
