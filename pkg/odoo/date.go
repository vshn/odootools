package odoo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"
)

const (
	DateFormat     = "2006-01-02"
	TimeFormat     = "15:04:05"
	DateTimeFormat = DateFormat + " " + TimeFormat
)

// Date is an Odoo-specific format of a timestamp
type Date struct {
	time.Time
}

// NewDate returns a new Date.
func NewDate(year int, month time.Month, day, hour, minute, second int, loc *time.Location) Date {
	return Date{
		Time: time.Date(year, month, day, hour, minute, second, 0, loc),
	}
}

func (d *Date) MarshalJSON() ([]byte, error) {
	if d.IsZero() {
		return []byte("false"), nil
	}
	return []byte(fmt.Sprintf(`"%s"`, d.Format(DateTimeFormat))), nil
}

func (d *Date) UnmarshalJSON(b []byte) error {
	var f bool
	if err := json.Unmarshal(b, &f); err == nil || string(b) == "false" {
		return nil
	}
	ts := bytes.Trim(b, `"`)
	// try parsing date + time
	t, dateTimeErr := time.Parse(DateTimeFormat, string(ts))
	if dateTimeErr != nil {
		// second attempt parsing date only
		t, dateTimeErr = time.Parse(DateFormat, string(ts))
		if dateTimeErr != nil {
			return dateTimeErr
		}
	}

	*d = Date{Time: t}
	return nil
}

// IsWithinTimeRange returns true if the date is between the given times.
// The date is considered in the range if from and to equal to the date respectively.
func IsWithinTimeRange(date, from, to time.Time) bool {
	// time.After doesn't return true if the unix seconds are the same.
	// Yet some users record attendances exactly midnight 00:00:00 and that causes same-timestamp issues.
	isBetween := date.After(from) && date.Before(to)
	return isBetween || date.Unix() == from.Unix() || date.Unix() == to.Unix()
}

// LocalizeTime returns the same time but with a different location.
// As opposed to time.In(loc), the returned time is not just updated with the location.
func LocalizeTime(tm time.Time, loc *time.Location) time.Time {
	return time.Date(tm.Year(), tm.Month(), tm.Day(), tm.Hour(), tm.Minute(), tm.Second(), tm.Nanosecond(), loc)
}

// Midnight returns a new time object in midnight (most recently past).
func Midnight(tm time.Time) time.Time {
	return time.Date(tm.Year(), tm.Month(), tm.Day(), 0, 0, 0, 0, tm.Location())
}

// MustParseDateTime parses the given value in DateTimeFormat or panics if it fails.
func MustParseDateTime(value string) Date {
	tm, err := ParseDateTime(value)
	if err != nil {
		panic(err)
	}
	return Date{Time: tm}
}

// MustParseDate parses the given value in DateFormat or panics if it fails.
func MustParseDate(value string) Date {
	tm, err := ParseDate(value)
	if err != nil {
		panic(err)
	}
	return Date{Time: tm}
}

// ParseDate parses the given value in DateFormat in UTC.
func ParseDate(value string) (time.Time, error) {
	tm, err := time.Parse(DateFormat, value)
	return tm, err
}

// ParseDateTime parses the given value in DateTimeFormat in UTC.
func ParseDateTime(value string) (time.Time, error) {
	tm, err := time.Parse(DateTimeFormat, value)
	return tm, err
}
