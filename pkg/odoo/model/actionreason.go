package model

import "encoding/json"

// ActionReason describes the "action reason" from Odoo.
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
type ActionReason struct {
	ID   float64
	Name string
}

// String implements fmt.Stringer.
func (reason *ActionReason) String() string {
	if reason == nil {
		return ""
	}
	return reason.Name
}

// MarshalJSON implements json.Marshaler.
func (reason ActionReason) MarshalJSON() ([]byte, error) {
	if reason.Name == "" {
		return []byte("false"), nil
	}
	arr := []interface{}{reason.ID, reason.Name}
	return json.Marshal(arr)
}

// UnmarshalJSON implements json.Unmarshaler.
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
