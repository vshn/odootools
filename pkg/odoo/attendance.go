package odoo

import (
	"fmt"
)

type Attendance struct {
	// ID is an unique ID for each attendance entry
	ID int `json:"id,omitempty"`

	// DateTime is the entry timestamp in UTC
	// Format: '2006-01-02 15:04:05'
	DateTime *Date `json:"name,omitempty"`

	// Action is either "sign_in" or "sign_out"
	Action string `json:"action,omitempty"`

	// Reason describes the "action reason" from Odoo.
	// NOTE: This field has special meaning when calculating the overtime.
	Reason *ActionReason `json:"action_desc,omitempty"`
}

func (c Client) FetchAllAttendances(sid string, employeeID int) ([]Attendance, error) {
	return c.fetchAttendances(sid, []Filter{[]interface{}{"employee_id", "=", employeeID}})
}

func (c Client) FetchAttendancesBetweenDates(sid string, employeeID int, begin, end Date) ([]Attendance, error) {
	return c.fetchAttendances(sid, []Filter{
		[]interface{}{"employee_id", "=", employeeID},
		[]string{"name", ">=", begin.ToTime().Format(DateFormat)},
		[]string{"name", "<=", end.ToTime().Format(DateFormat)},
	})
}

func (c Client) fetchAttendances(sid string, domainFilters []Filter) ([]Attendance, error) {
	body, err := NewJsonRpcRequest(&ReadModelRequest{
		Model:  "hr.attendance",
		Domain: domainFilters,
		Fields: []string{"employee_id", "name", "action", "action_desc"},
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
		Length  int          `json:"length,omitempty"`
		Records []Attendance `json:"records,omitempty"`
	}
	result := &readResult{}
	if err := c.unmarshalResponse(res.Body, result); err != nil {
		return nil, err
	}
	return result.Records, nil
}
