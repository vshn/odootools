package timesheet

import (
	"fmt"
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

	StateApproved  = "validate"
	StateToApprove = "confirm"
	StateDraft     = "draft"
)

// for testing purposes
var now = time.Now()

const (
	ActionSignIn  = "sign_in"
	ActionSignOut = "sign_out"
)

type AttendanceShift struct {
	// Start is the localized beginning time of the attendance
	Start time.Time
	// End is the localized finish time of the attendance
	End    time.Time
	Reason string
}

type AbsenceBlock struct {
	Date   time.Time
	Reason string
}

type Summary struct {
	TotalOvertime    time.Duration
	TotalLeave       time.Duration
	TotalExcusedTime time.Duration
	TotalWorkedTime  time.Duration
}

type MonthlyReport struct {
	DailySummaries []*DailySummary
	Summary        Summary
	Employee       *odoo.Employee
	Year           int
	Month          int
}

type ReportBuilder struct {
	attendances []odoo.Attendance
	leaves      []odoo.Leave
	employee    *odoo.Employee
	year        int
	month       int
	contracts   odoo.ContractList
	timezone    *time.Location
}

func NewReporter(attendances []odoo.Attendance, leaves []odoo.Leave, employee *odoo.Employee, contracts []odoo.Contract) *ReportBuilder {
	return &ReportBuilder{
		attendances: attendances,
		leaves:      leaves,
		employee:    employee,
		year:        now().UTC().Year(),
		month:       int(now().UTC().Month()),
		contracts:   contracts,
		timezone:    time.Local,
	}
}

func (r *ReportBuilder) SetMonth(year, month int) *ReportBuilder {
	r.year = year
	r.month = month
	return r
}

func (r *ReportBuilder) SetTimeZone(zone string) *ReportBuilder {
	loc, err := time.LoadLocation(zone)
	if err == nil {
		r.timezone = loc
	}
	return r
}

func (r *ReportBuilder) CalculateMonthlyReport() MonthlyReport {
	filteredAttendances := r.filterAttendancesInMonth()
	shifts := r.reduceAttendancesToShifts(filteredAttendances)
	filteredLeaves := r.filterLeavesInMonth()
	absences := r.reduceLeavesToBlocks(filteredLeaves)
	dailySummaries := r.prepareDays()

	r.addAttendanceShiftsToDailies(shifts, dailySummaries)
	r.addAbsencesToDailies(absences, dailySummaries)

	summary := Summary{}
	for _, dailySummary := range dailySummaries {
		summary.TotalOvertime += dailySummary.CalculateOvertime()
		summary.TotalLeave += dailySummary.CalculateAbsenceTime()
		summary.TotalExcusedTime += dailySummary.CalculateExcusedTime()
		summary.TotalWorkedTime += dailySummary.CalculateWorkingTime()
	}
	return MonthlyReport{
		DailySummaries: dailySummaries,
		Summary:        summary,
		Employee:       r.employee,
		Year:           r.year,
		Month:          r.month,
	}
}

func (r *ReportBuilder) reduceAttendancesToShifts(attendances []odoo.Attendance) []AttendanceShift {
	sortAttendances(attendances)
	shifts := make([]AttendanceShift, 0)
	var tmpShift AttendanceShift
	for _, attendance := range attendances {
		if attendance.Action == ActionSignIn {
			tmpShift = AttendanceShift{
				Start:  attendance.DateTime.ToTime().In(r.timezone),
				Reason: attendance.Reason.String(),
			}
		}
		if attendance.Action == ActionSignOut {
			tmpShift.End = attendance.DateTime.ToTime().In(r.timezone)
			shifts = append(shifts, tmpShift)
		}
	}
	return shifts
}

func (r *ReportBuilder) reduceLeavesToBlocks(leaves []odoo.Leave) []AbsenceBlock {
	blocks := make([]AbsenceBlock, 0)
	for _, leave := range leaves {
		// Only consider approved leaves
		if leave.State == StateApproved {
			blocks = append(blocks, AbsenceBlock{
				Reason: leave.Type.String(),
				Date:   leave.DateFrom.ToTime().In(r.timezone).Truncate(24 * time.Hour),
			})
		}
	}
	return blocks
}

func (r *ReportBuilder) prepareDays() []*DailySummary {
	days := make([]*DailySummary, 0)

	firstDay := time.Date(r.year, time.Month(r.month), 1, 0, 0, 0, 0, time.UTC)
	lastDay := firstDay.AddDate(0, 1, 0)

	if lastDay.After(now().In(r.timezone)) {
		lastDay = r.getDateTomorrow()
	}

	for currentDay := firstDay; currentDay.Before(lastDay); currentDay = currentDay.AddDate(0, 0, 1) {
		currentRatio, err := r.contracts.GetFTERatioForDay(odoo.Date(currentDay))
		if err != nil {
			fmt.Println(err)
			currentRatio = 0
		}
		days = append(days, NewDailySummary(currentRatio, currentDay.In(r.timezone)))
	}

	return days
}

func (r *ReportBuilder) getDateTomorrow() time.Time {
	return now().In(r.timezone).Truncate(24*time.Hour).AddDate(0, 0, 1)
}

func (r *ReportBuilder) addAttendanceShiftsToDailies(shifts []AttendanceShift, dailySums []*DailySummary) {
	for _, shift := range shifts {
		existing, found := findDailySummaryByDate(dailySums, shift.Start)
		if found {
			existing.addAttendanceShift(shift)
			continue
		}
	}
}

func (r *ReportBuilder) addAbsencesToDailies(absences []AbsenceBlock, summaries []*DailySummary) {
	for _, block := range absences {
		existing, found := findDailySummaryByDate(summaries, block.Date)
		if found {
			existing.addAbsenceBlock(block)
			continue
		}
	}
}

func sortAttendances(filtered []odoo.Attendance) {
	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].DateTime.ToTime().Unix() < filtered[j].DateTime.ToTime().Unix()
	})
}

func (r *ReportBuilder) filterAttendancesInMonth() []odoo.Attendance {
	filteredAttendances := make([]odoo.Attendance, 0)
	for _, attendance := range r.attendances {
		if attendance.DateTime.WithLocation(r.timezone).IsWithinMonth(r.year, r.month) {
			filteredAttendances = append(filteredAttendances, attendance)
		}
	}
	return filteredAttendances
}

func (r *ReportBuilder) filterLeavesInMonth() []odoo.Leave {
	filteredLeaves := make([]odoo.Leave, 0)
	for _, leave := range r.leaves {
		splits := leave.SplitByDay()
		for _, split := range splits {
			date := split.DateFrom.WithLocation(r.timezone)
			if date.IsWithinMonth(r.year, r.month) && date.ToTime().Weekday() != time.Sunday && date.ToTime().Weekday() != time.Saturday {
				filteredLeaves = append(filteredLeaves, split)
			}
		}
	}
	return filteredLeaves
}
