package odoo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const (
	AttendanceDateFormat     = "2006-01-02"
	AttendanceTimeFormat     = "15:04:05"
	AttendanceDateTimeFormat = AttendanceDateFormat + " " + AttendanceTimeFormat
)

type Attendance struct {
	// ID is an unique ID for each attendance entry
	ID int `json:"id,omitempty"`

	// Name is the entry timestamp in UTC
	// Format: '2006-01-02 15:04:05'
	Name *AttendanceTime `json:"name,omitempty"`

	// Action is either "sign_in" or "sign_out"
	Action string `json:"action,omitempty"`

	// ActionDesc describes the "action reason" from Odoo.
	//
	// Example values:
	// * `[1, "Outside office hours"]`
	// * `[2, "Outside office hours"]`
	// * `[3, "Sick / Medical Consultation"]`
	// * `[4, "Sick / Medical Consultation"]`
	// * `[5, "Authorities"]`
	// * `[6, "Authorities"]`
	// * `[27, "Requested Public Service"]`
	// * `[28, "Requested Public Service"]`
	//
	// NOTE: This field has special meaning when calculating the working hours:
	// * "Outside office hours" - 1.5x bonus
	ActionDesc *json.RawMessage `json:"action_desc,omitempty"`

	// WorkedHours is the amount of time Odoo determined.
	// Will always be "0.0" if "action" is "sign_in". Values DO NOT reflect
	// special boni like the 1.5x bonus for "Outside office hours".
	WorkedHours float64
}

type AttendanceTime time.Time

func (at *AttendanceTime) String() string {
	t := time.Time(*at)
	return t.Format(AttendanceDateTimeFormat)
}
func (at *AttendanceTime) Date() time.Time {
	t := time.Time(*at)
	return t.Truncate(24 * time.Hour)
}
func (at *AttendanceTime) Time() string {
	t := time.Time(*at)
	return t.Format(AttendanceTimeFormat)
}
func (at AttendanceTime) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, at.String())), nil
}
func (at *AttendanceTime) UnmarshalJSON(b []byte) error {
	ts := bytes.Trim(b, `"`)
	t, err := time.Parse(AttendanceDateTimeFormat, string(ts))
	if err != nil {
		return err
	}

	*at = AttendanceTime(t)
	return nil
}
func (at *AttendanceTime) ToTime() time.Time {
	return time.Time(*at)
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

	// Create request
	req, err := http.NewRequest("POST", c.baseURL+"/web/dataset/search_read", body)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("content-type", "application/json")
	req.Header.Set("cookie", "session_id="+sid)

	// Send request
	res, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("sending HTTP request: %w", err)
	} else if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("expected HTTP status 200 OK, got %s", res.Status)
	}

	// decode response
	var result readAttendancesResult
	if err := DecodeResult(res.Body, &result); err != nil {
		return nil, fmt.Errorf("decoding result: %w", err)
	}

	return result.Records, nil
}
