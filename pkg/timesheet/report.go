package timesheet

import (
	"sort"
	"time"

	"github.com/vshn/odootools/pkg/odoo"
)

const (
	ReasonSickLeave          = "Sick / Medical Consultation"
	ReasonOutsideOfficeHours = "Outside office hours"
	ReasonAuthorities        = "Authorities"
	ReasonPublicService      = "Requested Public Service"

	TypePublicHoliday     = "Public Holiday"
	TypeMilitaryService   = "Military Service"
	TypeSpecialOccasions  = "Special Occasions"
	TypeUnpaid            = "Unpaid"
	TypeLegalLeavesPrefix = "Legal Leaves"
)

type AttendanceBlock struct {
	Start  time.Time
	End    time.Time
	Reason string
}

type Summary struct {
	TotalWorkedHours time.Duration
	TotalLeaveDays   time.Duration
}

type Report struct {
	DailySummaries []*DailySummary
	Summary        Summary
}

type Reporter struct {
	attendances []odoo.Attendance
	leaves      []odoo.Leave
}

func NewReporter(attendances []odoo.Attendance, leaves []odoo.Leave) *Reporter {
	return &Reporter{
		attendances: attendances,
		leaves:      leaves,
	}
}

func (r *Reporter) CalculateReportForMonth(year, month int, fteRatio float64) Report {
	filtered := r.filterAttendancesInMonth(year, month)
	blocks := reduceAttendancesToBlocks(filtered)
	dailySummaries := reduceAttendanceBlocksToDailies(blocks, fteRatio)



	summary := Summary{}
	for _, dailySummary := range dailySummaries {
		summary.TotalWorkedHours += dailySummary.CalculateOvertime()
	}
	return Report{
		DailySummaries: dailySummaries,
		Summary:        summary,
	}
}

func reduceAttendancesToBlocks(attendances []odoo.Attendance) []AttendanceBlock {
	sortAttendances(attendances)
	blocks := make([]AttendanceBlock, 0)
	var tmpBlock AttendanceBlock
	for _, attendance := range attendances {
		if attendance.Action == "sign_in" {
			tmpBlock = AttendanceBlock{
				Start:  attendance.DateTime.ToTime(),
				Reason: attendance.Reason.String(),
			}
		}
		if attendance.Action == "sign_out" {
			tmpBlock.End = attendance.DateTime.ToTime()
			blocks = append(blocks, tmpBlock)
		}
	}
	return blocks
}

func reduceAttendanceBlocksToDailies(blocks []AttendanceBlock, ratio float64) []*DailySummary {
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
		if attendance.DateTime.IsWithinMonth(year, month) {
			filteredAttendances = append(filteredAttendances, attendance)
		}
	}
	return filteredAttendances
}

func (r *Reporter) filterLeavesInMonth(year int, month int) []odoo.Leave {
	filteredLeaves := make([]odoo.Leave, 0)
	for _, leave := range r.leaves {
		splits := leave.SplitByDay()
		for _, split := range splits {
			date := split.DateFrom
			if date.IsWithinMonth(year, month) && date.ToTime().Weekday() != time.Sunday && date.ToTime().Weekday() != time.Saturday {
				filteredLeaves = append(filteredLeaves, split)
			}
		}
	}
	return filteredLeaves
}
