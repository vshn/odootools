package controller

import (
	"fmt"
	"strconv"
	"time"

	"github.com/vshn/odootools/pkg/odoo"
	"github.com/vshn/odootools/pkg/timesheet"
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

// FormatDailySummary returns Values with sensible format.
func (v BaseView) FormatDailySummary(report timesheet.Report, daily *timesheet.DailySummary) Values {
	overtimeSummary := daily.CalculateOvertimeSummary()
	basic := Values{
		"Weekday":           daily.Date.Weekday(),
		"Date":              daily.Date.Format(odoo.DateFormat),
		"Workload":          daily.FTERatio * 100,
		"ExcusedHours":      v.FormatDurationInHours(overtimeSummary.ExcusedTime()),
		"WorkedHours":       v.FormatDurationInHours(overtimeSummary.WorkingTime()),
		"OvertimeHours":     v.FormatDurationInHours(overtimeSummary.Overtime()),
		"OvertimeClassname": v.OvertimeClassnameThreshold(overtimeSummary.Overtime(), overtimeSummary.DailyMax),
		"LeaveType":         "",
		"ValidationError":   daily.ValidateTimesheetEntries(),
	}
	if daily.HasAbsences() {
		basic["LeaveType"] = daily.Absences[0].Reason
	}
	if report.From.Location() != daily.Date.Location() {
		basic["Timezone"] = daily.Date.Location().String()
	}
	return basic
}

func (v BaseView) OvertimeClassname(duration time.Duration) string {
	overtimeClassname := ""
	if duration == 0 {
		return overtimeClassname
	}
	if duration.Minutes() > 15 {
		overtimeClassname = "Overtime"
	}
	if duration.Minutes() < -15 {
		overtimeClassname = "Undertime"
	}
	return overtimeClassname
}

func (v BaseView) OvertimeClassnameThreshold(duration time.Duration, dailyMax time.Duration) string {
	overtimeClassname := ""
	threshold := dailyMax.Seconds() + duration.Seconds()
	if threshold > dailyMax.Seconds()*1.03 {
		overtimeClassname = "Overtime"
	}
	if threshold < dailyMax.Seconds()*0.97 {
		overtimeClassname = "Undertime"
	}
	return overtimeClassname
}
