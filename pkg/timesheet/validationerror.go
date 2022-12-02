package timesheet

import (
	"errors"
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
	if l == nil || len(l.Errors) == 0 {
		return ""
	}
	dateList := make([]string, len(l.Errors))
	for i, err := range l.Errors {
		dateList[i] = err.Date.Format(odoo.DateFormat)
	}
	joinedList := strings.Join(dateList, ", ")
	return fmt.Sprintf("Report invalid for date(s): [%s]", joinedList)
}

// AppendValidationError appends an err to the given list, if err is of type ValidationError.
func AppendValidationError(list *ValidationErrorList, err error) {
	if list == nil && err == nil {
		return
	}
	var validationError *ValidationError
	if errors.As(err, &validationError) {
		list.Errors = append(list.Errors, validationError)
	}
}
