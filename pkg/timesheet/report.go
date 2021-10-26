package timesheet

import (
	"sort"
	"time"

	"github.com/mhutter/vshn-ftb/pkg/odoo"
)

const (
	ReasonSickLeave          = "Sick / Medical Consultation"
	ReasonOutsideOfficeHours = "Outside office hours"
	ReasonAuthorities        = "Authorities"
	ReasonPublicService      = "Requested Public Service"
)

type AttendanceBlock struct {
	Start  time.Time
	End    time.Time
	Reason string
}

type Summary struct {
	TotalWorkedHours time.Duration
}

type Report struct {
	DailySummaries []*DailySummary
	Summary        Summary
}

type Reporter struct {
	attendances []odoo.Attendance
	leaves      []odoo.Leave
}

func NewReport() *Reporter {
	return &Reporter{}
}

func (r *Reporter) SetAttendances(attendances []odoo.Attendance) *Reporter {
	r.attendances = attendances
	return r
}

func (r *Reporter) SetLeaves(leaves []odoo.Leave) *Reporter {
	r.leaves = leaves
	return r
}

func (r *Reporter) CalculateReportForMonth(year, month int, fteRatio float64) Report {
	filtered := r.filterAttendancesInMonth(year, month)

	sortedEntries := make([]AttendanceBlock, 0)

	sortAttendances(filtered)

	var entry AttendanceBlock
	for _, attendance := range filtered {
		if attendance.Action == "sign_in" {
			entry = AttendanceBlock{
				Start:  attendance.DateTime.ToTime(),
				Reason: attendance.Reason.String(),
			}
		}
		if attendance.Action == "sign_out" {
			entry.End = attendance.DateTime.ToTime()
			sortedEntries = append(sortedEntries, entry)
		}
	}

	dailySummaries := reduceAttendanceBlocks(sortedEntries, fteRatio)

	summary := Summary{}
	for _, dailySummary := range dailySummaries {
		summary.TotalWorkedHours += dailySummary.CalculateOvertime()
	}
	return Report{
		DailySummaries: dailySummaries,
		Summary:        summary,
	}
}

func reduceAttendanceBlocks(blocks []AttendanceBlock, ratio float64) []*DailySummary {
	dailySums := make([]*DailySummary, 0)

	for _, block := range blocks {
		existing, found := findDailySummaryByDate(dailySums, block.Start)
		if found {
			existing.addAttendanceBlock(block)
			continue
		}
		newDaily := NewDailySummary(ratio)
		newDaily.addAttendanceBlock(block)
		dailySums = append(dailySums, newDaily)
	}
	return dailySums
}

func sortAttendances(filtered []odoo.Attendance) {
	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].DateTime.ToTime().Unix() < filtered[j].DateTime.ToTime().Unix()
	})
}

func (r *Reporter) filterAttendancesInMonth(year int, month int) []odoo.Attendance {
	filteredAttendances := make([]odoo.Attendance, 0)
	for _, attendance := range r.attendances {
		if isInMonth(attendance, year, month) {
			filteredAttendances = append(filteredAttendances, attendance)
		}
	}
	return filteredAttendances
}

func isInMonth(attendance odoo.Attendance, year, month int) bool {
	firstDayOfMonth := time.Date(year, time.Month(month), 1, 0, 0, 1, 0, time.Now().Location())
	nextMonth := firstDayOfMonth.AddDate(0, 1, 0)
	date := attendance.DateTime.ToTime()
	return date.After(firstDayOfMonth) && date.Before(nextMonth)
}
