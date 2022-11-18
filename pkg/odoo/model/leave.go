package model

import (
	"context"
	"time"

	"github.com/vshn/odootools/pkg/odoo"
)

type Leave struct {
	// ID is an unique ID for each leave entry
	ID int `json:"id"`

	// DateFrom is the starting timestamp of the leave in UTC
	// Format: DateTimeFormat
	DateFrom odoo.Date `json:"date_from,omitempty"`

	// DateTo is the ending timestamp of the leave in UTC
	// Format: DateTimeFormat
	DateTo odoo.Date `json:"date_to,omitempty"`

	// Type describes the "leave type" from Odoo.
	Type *LeaveType `json:"holiday_status_id,omitempty"`

	// State is the leave request state.
	// Example raw values returned from Odoo:
	//  * `draft` (To Submit)
	//  * `confirm` (To Approve)
	//  * `validate` (Approved)
	State string `json:"state,omitempty"`
}

func (o Odoo) FetchLeavesBetweenDates(ctx context.Context, employeeID int, begin, end time.Time) (odoo.List[Leave], error) {
	beginStr := begin.Format(odoo.DateFormat)
	endStr := end.Format(odoo.DateFormat)
	return o.readLeaves(ctx, []odoo.Filter{
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

func (o Odoo) readLeaves(ctx context.Context, domainFilters []odoo.Filter) (odoo.List[Leave], error) {
	result := odoo.List[Leave]{}
	err := o.querier.SearchGenericModel(ctx, odoo.SearchReadModel{
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
	if l.DateFrom.Day() == l.DateTo.Day() {
		arr = append(arr, l)
		return arr
	}
	totalDuration := l.DateTo.Sub(l.DateFrom.Time)
	days := totalDuration / (time.Hour * 24)
	hoursPerDay := days * 8 * time.Hour
	startDate := l.DateFrom.Time
	endDate := l.DateTo.Time
	for currentDate := startDate; currentDate.Before(endDate); currentDate = currentDate.AddDate(0, 0, 1) {
		from := odoo.Date{Time: currentDate}
		to := odoo.Date{Time: currentDate.Add(hoursPerDay)}
		newLeave := Leave{
			DateFrom: from,
			DateTo:   to,
			Type:     l.Type,
			State:    l.State,
		}
		arr = append(arr, newLeave)
	}
	return arr
}
