package odoo

import (
	"encoding/json"
	"fmt"
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
	//
	// Example raw values returned from Odoo:
	//  * `false` (if no specific reason given)
	//  * `[4, "Unpaid"]`
	//  * `[5, "Military Service"]`
	//  * `[7, "Special Occasions"]`
	//  * `[9, "Public Holiday"]`
	//  * `[16, "Legal Leaves 2020"]`
	//  * `[17, "Legal Leaves 2021"]`
	Type *LeaveType `json:"holiday_status_id,omitempty"`

	// State is the leave request state.
	// Example raw values returned from Odoo:
	//  * `draft` (To Submit)
	//  * `confirm` (To Approve)
	//  * `validate` (Approved)
	State string `json:"state,omitempty"`
}

type LeaveType struct {
	ID   float64
	Name string
}

func (c *Client) ReadAllLeaves(sid string, uid int) ([]Leave, error) {
	// Prepare "search leaves" request
	body, err := NewJsonRpcRequest(&ReadModelRequest{
		Model:  "hr.holidays",
		Domain: []Filter{
			{"employee_id.user_id.id", "=", uid},
			{"type", "=", "remove"}, // Only return used leaves. With type = "add" we would get leaves that add days to holiday budget
		},
		Fields: []string{"employee_id", "date_from", "date_to", "holiday_status_id", "state"},
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

	result := &readLeavesResult{}
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

////////////////// Boilerplate

type readLeavesResult struct {
	Length  int     `json:"length,omitempty"`
	Records []Leave `json:"records,omitempty"`
}

func (leaveType LeaveType) MarshalJSON() ([]byte, error) {
	if leaveType.Name == "" {
		return []byte("false"), nil
	}
	arr := []interface{}{leaveType.ID, leaveType.Name}
	return json.Marshal(arr)
}
func (leaveType *LeaveType) UnmarshalJSON(b []byte) error {
	var f bool
	if err := json.Unmarshal(b, &f); err == nil || string(b) == "false" {
		return nil
	}
	var arr []interface{}
	if err := json.Unmarshal(b, &arr); err != nil {
		return err
	}
	if len(arr) >= 2 {
		if v, ok := arr[1].(string); ok {
			*leaveType = LeaveType{
				ID:   arr[0].(float64),
				Name: v,
			}
		}
	}
	return nil
}
func (leaveType *LeaveType) String() string {
	if leaveType == nil {
		return ""
	}
	return leaveType.Name
}
