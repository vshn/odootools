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
type Date time.Time

func (d *Date) String() string {
	t := time.Time(*d)
	return t.Format(DateTimeFormat)
}

func (d Date) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, d.String())), nil
}

func (d *Date) UnmarshalJSON(b []byte) error {
	ts := bytes.Trim(b, `"`)
	var f bool
	if err := json.Unmarshal(b, &f); err == nil || string(b) == "false" {
		return nil
	}
	// try parsing date + time
	t, dateTimeErr := time.Parse(DateTimeFormat, string(ts))
	if dateTimeErr != nil {
		// second attempt parsing date only
		t, dateTimeErr = time.Parse(DateFormat, string(ts))
		if dateTimeErr != nil {
			return dateTimeErr
		}
	}

	*d = Date(t)
	return nil
}

// IsZero returns true if Date is nil or Time.IsZero()
func (d *Date) IsZero() bool {
	return d == nil || d.ToTime().IsZero()
}

func (d Date) ToTime() time.Time {
	return time.Time(d)
}

func (d Date) WithLocation(loc *time.Location) Date {
	return Date(d.ToTime().In(loc))
}

func (d Date) IsWithinMonth(year, month int) bool {
	firstDayOfMonth := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, d.ToTime().Location())
	nextMonth := firstDayOfMonth.AddDate(0, 1, 0)
	date := d.ToTime()
	// time.After doesn't return true if the unix seconds are the same.
	// Yet some users record attendances exactly midnight 00:00:00 and that causes same-timestamp issues.
	isBetween := date.After(firstDayOfMonth) && date.Before(nextMonth)
	return isBetween || date.Unix() == firstDayOfMonth.Unix()
}

// MustParseDateTime parses the given value in DateTimeFormat or panics if it fails.
func MustParseDateTime(value string) *Date {
	tm, err := time.Parse(DateTimeFormat, value)
	if err != nil {
		panic(err)
	}
	d := Date(tm)
	return &d
}

// MustParseDate parses the given value in DateFormat or panics if it fails.
func MustParseDate(value string) *Date {
	tm, err := time.Parse(DateFormat, value)
	if err != nil {
		panic(err)
	}
	d := Date(tm)
	return &d
}
