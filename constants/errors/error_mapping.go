package errors

import (
	errorField "payment-services/constants/errors/field"
	errorFieldSchedule "payment-services/constants/errors/fieldschedule"
	errorTime "payment-services/constants/errors/time"
)

func ErrMapping(err error) bool {
	var (
		GeneralErrors       = GeneralErrors
		FieldErrors         = errorField.FieldErrors
		FieldScheduleErrors = errorFieldSchedule.FieldScheduleErrors
		TimeErrors          = errorTime.TimeErrors
	)

	allErrors := make([]error, 0)
	allErrors = append(allErrors, GeneralErrors...)
	allErrors = append(allErrors, FieldErrors...)
	allErrors = append(allErrors, FieldScheduleErrors...)
	allErrors = append(allErrors, TimeErrors...)

	for _, item := range allErrors {
		if err.Error() == item.Error() {
			return true
		}
	}

	return false
}
