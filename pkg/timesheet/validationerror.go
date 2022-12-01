package timesheet

import (
	"fmt"
	"strings"
	"time"

	"github.com/vshn/odootools/pkg/odoo"
)

type ValidationError struct {
	Date time.Time
	Err  error
}

type ValidationErrorList struct {
	Errors []*ValidationError
}

func NewValidationError(forDate time.Time, err error) *ValidationError {
	if err == nil {
		return nil
	}
	return &ValidationError{Date: forDate, Err: err}
}

func (e *ValidationError) Error() string {
	return e.Err.Error()
}

func (e *ValidationError) Unwrap() error {
	return e.Err
}

// Error returns a comma-separated list of dates that have a validation error.
func (l *ValidationErrorList) Error() string {
	dateList := make([]string, len(l.Errors))
	for i, err := range l.Errors {
		dateList[i] = err.Date.Format(odoo.DateFormat)
	}
	joinedList := strings.Join(dateList, ", ")
	return fmt.Sprintf("validation invalid for date(s): [%s]", joinedList)
}

func AppendValidationError(list *ValidationErrorList, err *ValidationError) {
	if list == nil && err == nil {
		return
	}
	if list == nil {
		*list = ValidationErrorList{}
	}
	list.Errors = append(list.Errors, err)
}
