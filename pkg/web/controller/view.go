package controller

import (
	"fmt"
	"strconv"
	"time"
)

// BaseView contains some utility methods.
type BaseView struct {
}

// FormatDurationInHours returns a human friendly "0:00"-formatted duration.
// Seconds within a minute are rounded up or down to the nearest full minute.
// A sign ("-") is prefixed if duration is negative.
func (v BaseView) FormatDurationInHours(d time.Duration) string {
	sign := ""
	if d.Seconds() < 0 {
		sign = "-"
		d = time.Duration(d.Nanoseconds() * -1)
	}
	d = d.Round(time.Minute)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	return fmt.Sprintf("%s%d:%02d", sign, h, m)
}

// FormatFloat returns a string of the given float with desired digits after comma.
func (v BaseView) FormatFloat(value float64, precision int) string {
	return strconv.FormatFloat(value, 'f', precision, 64)
}
