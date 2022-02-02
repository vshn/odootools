package model

import (
	"context"
	"strconv"
	"time"

	"github.com/vshn/odootools/pkg/odoo"
)

type Leave struct {
	// ID is an unique ID for each leave entry
	ID int `json:"id"`

	// DateFrom is the starting timestamp of the leave in UTC
	// Format: DateTimeFormat
	DateFrom *odoo.Date `json:"date_from"`

	// DateTo is the ending timestamp of the leave in UTC
	// Format: DateTimeFormat
	DateTo *odoo.Date `json:"date_to"`

	// Type describes the "leave type" from Odoo.
	Type *odoo.LeaveType `json:"holiday_status_id,omitempty"`

	// State is the leave request state.
	// Example raw values returned from Odoo:
	//  * `draft` (To Submit)
	//  * `confirm` (To Approve)
	//  * `validate` (Approved)
	State string `json:"state,omitempty"`
}

// LeaveList contains a slice of Leave.
type LeaveList struct {
	Items []Leave `json:"records,omitempty"`
}

func (o Odoo) FetchAllLeaves(employeeID int) (LeaveList, error) {
	return o.readLeaves([]odoo.Filter{
		[]string{"employee_id", "=", strconv.Itoa(employeeID)},
		[]string{"type", "=", "remove"}, // Only return used leaves. With type = "add" we would get leaves that add days to holiday budget
	})
}

func (o Odoo) FetchLeavesBetweenDates(employeeID int, begin, end time.Time) (LeaveList, error) {
	beginStr := begin.Format(odoo.DateFormat)
	endStr := end.Format(odoo.DateFormat)
	return o.readLeaves([]odoo.Filter{
		[]string{"type", "=", "remove"}, // Only return used leaves. With type = "add" we would get leaves that add days to holiday budget
		[]interface{}{"employee_id", "=", employeeID},
		"|",
		"|",
		"&",
		[]string{"date_from", ">=", beginStr},
		[]string{"date_from", "<=", endStr},
		"&",
		[]string{"date_from", "<=", beginStr},
		[]string{"date_to", ">=", beginStr},
		"&",
		[]string{"date_from", "<=", endStr},
		[]string{"date_to", ">=", beginStr},
	})
}

func (o Odoo) readLeaves(domainFilters []odoo.Filter) (LeaveList, error) {
	result := LeaveList{}
	err := o.querier.SearchGenericModel(context.Background(), odoo.SearchReadModel{
		Model:  "hr.holidays",
		Domain: domainFilters,
		Fields: []string{"date_from", "date_to", "holiday_status_id", "state"},
		Limit:  0,
		Offset: 0,
	}, &result)
	return result, err
}

func (l Leave) SplitByDay() []Leave {
	arr := make([]Leave, 0)
	if l.DateFrom.ToTime().Day() == l.DateTo.ToTime().Day() {
		arr = append(arr, l)
		return arr
	}
	totalDuration := l.DateTo.ToTime().Sub(l.DateFrom.ToTime())
	days := totalDuration / (time.Hour * 24)
	hoursPerDay := days * 8 * time.Hour
	startDate := l.DateFrom.ToTime()
	endDate := l.DateTo.ToTime()
	for currentDate := startDate; currentDate.Before(endDate); currentDate = currentDate.AddDate(0, 0, 1) {
		from := odoo.Date(currentDate)
		to := odoo.Date(currentDate.Add(hoursPerDay))
		newLeave := Leave{
			DateFrom: &from,
			DateTo:   &to,
			Type:     l.Type,
			State:    l.State,
		}
		arr = append(arr, newLeave)
	}
	return arr
}