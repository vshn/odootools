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

// String returns the location name.
// Returns empty string if nil.
func (tz *TimeZone) String() string {
	if tz == nil || tz.Location == nil {
		return ""
	}
	return tz.Location.String()
}

// IsEqualTo returns true if the given TimeZone is equal to other.
// If both are nil, it returns true.
func (tz *TimeZone) IsEqualTo(other *TimeZone) bool {
	return tz.String() == other.String()
}
