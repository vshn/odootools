package odoo

import (
	"bytes"
	"fmt"
	"time"
)

type TimeZone time.Location

// UnmarshalJSON implements json.Unmarshaler.
func (tz *TimeZone) UnmarshalJSON(b []byte) error {
	ts := bytes.Trim(b, `"`)
	loc, err := time.LoadLocation(string(ts))
	if err != nil {
		return fmt.Errorf("cannot unmarshal json: %w", err)
	}
	*tz = TimeZone(*loc)
	return nil
}

func (tz *TimeZone) Location() *time.Location {
	if tz == nil {
		return nil
	}
	l := time.Location(*tz)
	return &l
}
