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

// for testing purposes
var now = time.Now

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
	year        int
	month       int
	fteRatio    float64
}

func NewReporter(attendances []odoo.Attendance, leaves []odoo.Leave) *Reporter {
	return &Reporter{
		attendances: attendances,
		leaves:      leaves,
		year:        now().UTC().Year(),
		month:       int(now().UTC().Month()),
		fteRatio:    float64(1),
	}
}

func (r *Reporter) SetMonth(year, month int) *Reporter {
	r.year = year
	r.month = month
	return r
}

func (r *Reporter) SetFteRatio(fteRatio float64) *Reporter {
	r.fteRatio = fteRatio
	return r
}

func (r *Reporter) CalculateReport() Report {
	filtered := r.filterAttendancesInMonth()
	blocks := reduceAttendancesToBlocks(filtered)
	dailySummaries := r.prepareWorkdays()

	r.addAttendanceBlocksToDailies(blocks, dailySummaries)

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

func (r *Reporter) prepareWorkdays() []*DailySummary {
	days := make([]*DailySummary, 0)

	firstDay := time.Date(r.year, time.Month(r.month), 1, 0, 0, 0, 0, time.UTC)
	lastDay := firstDay.AddDate(0, 1, 0)

	if lastDay.After(now().UTC()) {
		lastDay = getDateTomorrow()
	}

	for currentDay := firstDay; currentDay.Before(lastDay); currentDay = currentDay.AddDate(0, 0, 1) {
		days = append(days, NewDailySummary(r.fteRatio, currentDay))
	}

	return days
}

func getDateTomorrow() time.Time {
	return now().UTC().Truncate(24*time.Hour).AddDate(0, 0, 1)
}

func (r *Reporter) addAttendanceBlocksToDailies(blocks []AttendanceBlock, dailySums []*DailySummary) {
	for _, block := range blocks {
		existing, found := findDailySummaryByDate(dailySums, block.Start)
		if found {
			existing.addAttendanceBlock(block)
			continue
		}
	}
}

func sortAttendances(filtered []odoo.Attendance) {
	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].DateTime.ToTime().Unix() < filtered[j].DateTime.ToTime().Unix()
	})
}

func (r *Reporter) filterAttendancesInMonth() []odoo.Attendance {
	filteredAttendances := make([]odoo.Attendance, 0)
	for _, attendance := range r.attendances {
		if attendance.DateTime.IsWithinMonth(r.year, r.month) {
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
