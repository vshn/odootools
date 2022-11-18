package odoo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"
)

// TimeZone represents a time zone in Odoo.
type TimeZone struct {
	*time.Location
}

func NewTimeZone(loc *time.Location) *TimeZone {
	return &TimeZone{Location: loc}
}

// UnmarshalJSON implements json.Unmarshaler.
func (tz *TimeZone) UnmarshalJSON(b []byte) error {
	var f bool
	if err := json.Unmarshal(b, &f); err == nil || string(b) == "false" || string(b) == "" {
		return nil
	}
	ts := bytes.Trim(b, `"`)
	loc, err := time.LoadLocation(string(ts))
	if err != nil {
		return fmt.Errorf("cannot unmarshal json: %w", err)
	}
	tz.Location = loc
	return nil
}

// MarshalJSON implements json.Marshaler.
func (tz *TimeZone) MarshalJSON() ([]byte, error) {
	if tz.IsEmpty() || tz.Location == time.Local {
		return []byte(`null`), nil
	}
	return []byte(fmt.Sprintf(`"%s"`, tz.Location)), nil
}

// LocationOrDefault returns the location if it's defined, or given default it not.
func (tz *TimeZone) LocationOrDefault(def *time.Location) *time.Location {
	if tz.IsEmpty() {
		return def
	}
	return tz.Location
}

// IsEmpty returns true if the location is nil.
func (tz *TimeZone) IsEmpty() bool {
	if tz == nil || tz.Location == nil {
		return true
	}
	return false
}
