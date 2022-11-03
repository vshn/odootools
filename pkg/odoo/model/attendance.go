package model

import (
	"context"
	"sort"
	"time"

	"github.com/vshn/odootools/pkg/odoo"
)

// Attendance is an entry or closing event of a shift.
type Attendance struct {
	// ID is an unique ID for each attendance entry
	ID int `json:"id,omitempty"`

	// DateTime is the entry timestamp in UTC
	// Format: '2006-01-02 15:04:05'
	DateTime *odoo.Date `json:"name,omitempty"`

	// Action is either "sign_in" or "sign_out"
	Action string `json:"action,omitempty"`

	// Reason describes the "action reason" from Odoo.
	// NOTE: This field has special meaning when calculating the overtime.
	Reason *ActionReason `json:"action_desc,omitempty"`
}

type AttendanceList odoo.List[Attendance]

// FetchAttendancesBetweenDates retrieves all attendances associated with the given employee between 2 dates (inclusive each).
func (o Odoo) FetchAttendancesBetweenDates(ctx context.Context, employeeID int, begin, end time.Time) (AttendanceList, error) {
	return o.fetchAttendances(ctx, []odoo.Filter{
		[]interface{}{"employee_id", "=", employeeID},
		[]string{"name", ">=", begin.Format(odoo.DateFormat)},
		[]string{"name", "<=", end.Format(odoo.DateFormat)},
	})
}

func (o Odoo) fetchAttendances(ctx context.Context, domainFilters []odoo.Filter) (AttendanceList, error) {
	result := AttendanceList{}
	err := o.querier.SearchGenericModel(ctx, odoo.SearchReadModel{
		Model:  "hr.attendance",
		Domain: domainFilters,
		Fields: []string{"employee_id", "name", "action", "action_desc"},
		Limit:  0,
		Offset: 0,
	}, &result)
	return result, err
}

// SortByDate sorts the attendances by date ascending (oldest first).
func (l AttendanceList) SortByDate() {
	items := l.Items
	sort.Slice(l.Items, func(i, j int) bool {
		return items[i].DateTime.ToTime().Unix() < items[j].DateTime.ToTime().Unix()
	})
}

// FilterAttendanceBetweenDates returns a new list that only contains items within the specified time range.
// It uses `from`'s location to set the timezone.
func (l AttendanceList) FilterAttendanceBetweenDates(from, to time.Time) AttendanceList {
	filteredAttendances := AttendanceList{}
	if l.Items != nil {
		filteredAttendances.Items = []Attendance{}
	}
	for _, attendance := range l.Items {
		if attendance.DateTime.WithLocation(from.Location()).IsWithinTimeRange(from, to) {
			filteredAttendances.Items = append(filteredAttendances.Items, attendance)
		}
	}
	return filteredAttendances
}

// AddCurrentTimeAsSignOut adds an Attendance with timesheet.ActionSignOut reason and with the current time.
// An attendance is only added if the last Attendance in the list is timesheet.ActionSignIn.
func (l AttendanceList) AddCurrentTimeAsSignOut(tz *time.Location) AttendanceList {
	if len(l.Items) == 0 {
		return l
	}
	lastAttendance := l.Items[len(l.Items)-1]
	if lastAttendance.Action != ActionSignIn {
		return l
	}

	now := odoo.Date(time.Now().In(tz))
	// fake a sign_out
	l.Items = append(l.Items, Attendance{
		DateTime: &now,
		Action:   ActionSignOut,
		Reason:   lastAttendance.Reason,
	})
	return l
}
