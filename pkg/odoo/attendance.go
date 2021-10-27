package odoo

import (
	"encoding/json"
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
	//
	// Example raw values returned from Odoo:
	//  * `false` (if no specific reason given)
	//  * `[1, "Outside office hours"]`
	//  * `[2, "Outside office hours"]`
	//  * `[3, "Sick / Medical Consultation"]`
	//  * `[4, "Sick / Medical Consultation"]`
	//  * `[5, "Authorities"]`
	//  * `[6, "Authorities"]`
	//  * `[27, "Requested Public Service"]`
	//  * `[28, "Requested Public Service"]`
	//
	// NOTE: This field has special meaning when calculating the overtime.
	Reason *ActionReason `json:"action_desc,omitempty"`
}

type ActionReason struct {
	ID   float64
	Name string
}

func (reason ActionReason) MarshalJSON() ([]byte, error) {
	if reason.Name == "" {
		return []byte("false"), nil
	}
	arr := []interface{}{reason.ID, reason.Name}
	return json.Marshal(arr)
}
func (reason *ActionReason) UnmarshalJSON(b []byte) error {
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
			*reason = ActionReason{
				ID:   arr[0].(float64),
				Name: v,
			}
		}
	}
	return nil
}
func (reason *ActionReason) String() string {
	if reason == nil {
		return ""
	}
	return reason.Name
}

type readAttendancesResult struct {
	Length  int          `json:"length,omitempty"`
	Records []Attendance `json:"records,omitempty"`
}

func (c Client) ReadAllAttendances(sid string, uid int) ([]Attendance, error) {
	// Prepare "search attendances" request
	body, err := NewJsonRpcRequest(&ReadModelRequest{
		Model:  "hr.attendance",
		Domain: []Filter{{"employee_id.user_id.id", "=", uid}},
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

	result := &readAttendancesResult{}
	if err := c.unmarshalResponse(res.Body, result); err != nil {
		return nil, err
	}
	return result.Records, nil
}
