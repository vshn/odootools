package model

import "encoding/json"

// LeaveType describes the "leave type" from Odoo.
//
// Example raw values returned from Odoo:
//   - `false` (if no specific reason given)
//   - `[4, "Unpaid"]`
//   - `[5, "Military Service"]`
//   - `[7, "Special Occasions"]`
//   - `[9, "Public Holiday"]`
//   - `[16, "Legal Leaves 2020"]`
//   - `[17, "Legal Leaves 2021"]`
type LeaveType struct {
	ID   float64
	Name string
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
