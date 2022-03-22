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
	d = d.Round(time.Second)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	d -= m * time.Minute
	s := d / time.Second
	return fmt.Sprintf("%s%d:%02d:%02d", sign, h, m, s)
}

// FormatFloat returns a string of the given float with desired digits after comma.
func (v BaseView) FormatFloat(value float64, precision int) string {
	return strconv.FormatFloat(value, 'f', precision, 64)
}

// GetNextMonth returns the numerical next month of the given input (month 1-12)
// The year is increased if month is >= 12.
func (v BaseView) GetNextMonth(year, month int) (int, int) {
	if month >= 12 {
		return year + 1, 1
	}
	return year, month + 1
}

// GetPreviousMonth returns the numerical previous month of the given input (month 1-12)
// The year is decreased if month is <= 1.
func (v BaseView) GetPreviousMonth(year, month int) (int, int) {
	if month <= 1 {
		return year - 1, 12
	}
	return year, month - 1
}
