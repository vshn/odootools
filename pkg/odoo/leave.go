package odoo

import (
	"fmt"
	"strconv"
	"time"
)

type Leave struct {
	// ID is an unique ID for each leave entry
	ID int `json:"id"`

	// DateFrom is the starting timestamp of the leave in UTC
	// Format: DateTimeFormat
	DateFrom *Date `json:"date_from"`

	// DateTo is the ending timestamp of the leave in UTC
	// Format: DateTimeFormat
	DateTo *Date `json:"date_to"`

	// Type describes the "leave type" from Odoo.
	Type *LeaveType `json:"holiday_status_id,omitempty"`

	// State is the leave request state.
	// Example raw values returned from Odoo:
	//  * `draft` (To Submit)
	//  * `confirm` (To Approve)
	//  * `validate` (Approved)
	State string `json:"state,omitempty"`
}

func (c *Client) FetchAllLeaves(sid string, employeeID int) ([]Leave, error) {
	return c.readLeaves(sid, []Filter{
		[]string{"employee_id", "=", strconv.Itoa(employeeID)},
		[]string{"type", "=", "remove"}, // Only return used leaves. With type = "add" we would get leaves that add days to holiday budget
	})
}

func (c *Client) FetchLeavesBetweenDates(sid string, employeeID int, begin, end Date) ([]Leave, error) {
	beginStr := begin.ToTime().Format(DateFormat)
	endStr := end.ToTime().Format(DateFormat)
	return c.readLeaves(sid, []Filter{
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

func (c Client) readLeaves(sid string, domainFilters []Filter) ([]Leave, error) {
	// Prepare "search leaves" request
	body, err := NewJsonRpcRequest(&ReadModelRequest{
		Model:  "hr.holidays",
		Domain: domainFilters,
		Fields: []string{"date_from", "date_to", "holiday_status_id", "state"},
		Limit:  0,
		Offset: 0,
	}).Encode()
	if err != nil {
		return nil, fmt.Errorf("encoding request: %w", err)
	}

	res, err := c.makeRequest(sid, body)
	if err != nil {
		return nil, err
	}

	type readResult struct {
		Length  int     `json:"length,omitempty"`
		Records []Leave `json:"records,omitempty"`
	}
	result := &readResult{}
	if err := c.unmarshalResponse(res.Body, result); err != nil {
		return nil, err
	}
	return result.Records, nil
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
		from := Date(currentDate)
		to := Date(currentDate.Add(hoursPerDay))
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
