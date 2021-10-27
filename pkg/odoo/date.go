package odoo

import (
	"bytes"
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
	t, err := time.Parse(DateTimeFormat, string(ts))
	if err != nil {
		return err
	}

	*d = Date(t)
	return nil
}
func (d *Date) ToTime() time.Time {
	return time.Time(*d)
}

func (d Date) IsWithinMonth(year, month int) bool {
	firstDayOfMonth := time.Date(year, time.Month(month), 1, 0, 0, 1, 0, time.Now().Location())
	nextMonth := firstDayOfMonth.AddDate(0, 1, 0)
	date := d.ToTime()
	return date.After(firstDayOfMonth) && date.Before(nextMonth)
}
