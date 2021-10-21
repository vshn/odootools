package timesheet

import (
	"sort"
	"time"

	"github.com/mhutter/vshn-ftb/pkg/odoo"
)

type AttendanceBlock struct {
	Start       time.Time
	End         time.Time
	Reason      string
	LoggedHours float64
}

type Summary struct {
	TotalWorkedHours float64
}

type DailySummary struct {
	Date    time.Time
	Entries []AttendanceBlock
}

type Report struct {
	DailySummaries []DailySummary
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

func (r *Reporter) CalculateReportForMonth(year, month int) Report {
	filtered := filterAttendancesInMonth(year, month, r)

	sortedEntries := make([]AttendanceBlock, 0)

	sortAttendances(filtered)

	var entry AttendanceBlock
	for _, attendance := range filtered {
		if attendance.Action == "sign_in" {
			entry = AttendanceBlock{
				Start: attendance.Name.ToTime(),
			}
		}
		if attendance.Action == "sign_out" {
			entry.End = attendance.Name.ToTime()
			entry.LoggedHours = entry.End.Sub(entry.Start).Hours()

			sortedEntries = append(sortedEntries, entry)
		}
	}

	dailySummaries := reduceAttendanceBlocks(sortedEntries)

	summary := Summary{}
	for _, dailySummary := range dailySummaries {
		summary.TotalWorkedHours += dailySummary.CalculateOvertime()
	}
	return Report{
		DailySummaries: dailySummaries,
		Summary:        summary,
	}
}

func reduceAttendanceBlocks(blocks []AttendanceBlock) []DailySummary {
	dailySums := make([]DailySummary, 0)

	var dailySumTemp *DailySummary
	for _, block := range blocks {
		if dailySumTemp == nil {
			dailySumTemp = &DailySummary{
				Entries: []AttendanceBlock{block},
				Date:    block.Start.Truncate(24 * time.Hour),
			}
			continue
		}
		if block.Start.Day() == dailySumTemp.Date.Day() {
			dailySumTemp.Entries = append(dailySumTemp.Entries, block)
			continue
		}
		dailySums = append(dailySums, *dailySumTemp)
		dailySumTemp = nil
	}
	return dailySums
}

func (s DailySummary) CalculateOvertime() float64 {
	workHours := float64(0)
	for _, block := range s.Entries {
		if block.Reason == "" {
			workHours += block.End.Sub(block.Start).Hours()
			continue
		}
		// TODO: Add 1.5 factor in special cases
		// TODO: respect FTE percentage
		// TODO: respect attendance reasons (e.g. no overtime in sick leave)

	}

	return workHours - 8
}

func sortAttendances(filtered []odoo.Attendance) {
	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].Name.ToTime().Unix() < filtered[j].Name.ToTime().Unix()
	})
}

func filterAttendancesInMonth(year int, month int, r *Reporter) []odoo.Attendance {
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
	date := attendance.Name.ToTime()
	return date.After(firstDayOfMonth) && date.Before(nextMonth)
}
