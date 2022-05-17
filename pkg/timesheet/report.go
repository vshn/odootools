package timesheet

import (
	"time"

	"github.com/vshn/odootools/pkg/odoo"
	"github.com/vshn/odootools/pkg/odoo/model"
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
var now = time.Now

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
	TotalExcusedTime time.Duration
	TotalWorkedTime  time.Duration
	// TotalLeave is the amount of paid leave days.
	// This value respects FTE ratio, e.g. in a 50% ratio a public holiday is still counted as '1d'.
	TotalLeave      float64
	AverageWorkload float64
}

type Report struct {
	DailySummaries []*DailySummary
	Summary        Summary
	Employee       *model.Employee
	// From is the first day (inclusive) of the time range
	From time.Time
	// To is the last day (inclusive) of the time range
	To time.Time
}

type ReportBuilder struct {
	attendances model.AttendanceList
	leaves      odoo.List[model.Leave]
	employee    *model.Employee
	from        time.Time
	to          time.Time
	contracts   model.ContractList
	timezone    *time.Location
	clampToNow  bool
}

func NewReporter(attendances model.AttendanceList, leaves odoo.List[model.Leave], employee *model.Employee, contracts model.ContractList) *ReportBuilder {
	return &ReportBuilder{
		attendances: attendances,
		leaves:      leaves,
		employee:    employee,
		contracts:   contracts,
		timezone:    time.Local,
		clampToNow:  true,
	}
}

// SetRange sets the time range in which the report should consider calculations.
// For example, to build a monthly report, set `from` to the first day of the month from midnight, and `to` to the first day of the next month at midnight.
func (r *ReportBuilder) SetRange(from, to time.Time) *ReportBuilder {
	r.from = from
	r.to = to
	return r
}

func (r *ReportBuilder) SetTimeZone(zone string) *ReportBuilder {
	loc, err := time.LoadLocation(zone)
	if err == nil {
		r.timezone = loc
	}
	return r
}

// SkipClampingToNow ignores the current time when preparing the daily summaries within the time range.
// By default, the reporter doesn't include days that are happening in the future and thus calculate overtime wrongly.
func (r *ReportBuilder) SkipClampingToNow(skip bool) *ReportBuilder {
	r.clampToNow = !skip
	return r
}

func (r *ReportBuilder) CalculateReport() (Report, error) {
	filteredAttendances := r.attendances.FilterAttendanceBetweenDates(r.from.In(r.timezone), r.to)
	shifts := r.reduceAttendancesToShifts(filteredAttendances)
	filteredLeaves := r.filterLeavesInTimeRange()
	absences := r.reduceLeavesToBlocks(filteredLeaves)
	dailySummaries, err := r.prepareDays()
	if err != nil {
		return Report{
			Employee: r.employee,
			From:     r.from,
			To:       r.to,
		}, err
	}

	r.addAttendanceShiftsToDailies(shifts, dailySummaries)
	r.addAbsencesToDailies(absences, dailySummaries)

	summary := Summary{}
	for _, dailySummary := range dailySummaries {
		overtimeSummary := dailySummary.CalculateOvertimeSummary()
		summary.TotalOvertime += overtimeSummary.Overtime()
		summary.TotalExcusedTime += overtimeSummary.ExcusedTime()
		summary.TotalWorkedTime += overtimeSummary.WorkingTime()
		if dailySummary.IsHoliday() {
			summary.TotalLeave += 1
		}
	}
	summary.AverageWorkload = r.calculateAverageWorkload(dailySummaries)
	return Report{
		DailySummaries: dailySummaries,
		Summary:        summary,
		Employee:       r.employee,
		From:           r.from,
		To:             r.to,
	}, nil
}

func (r *ReportBuilder) reduceAttendancesToShifts(attendances model.AttendanceList) []AttendanceShift {
	attendances.SortByDate()
	shifts := make([]AttendanceShift, 0)
	var tmpShift AttendanceShift
	for _, attendance := range attendances.Items {
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

func (r *ReportBuilder) calculateAverageWorkload(dailies []*DailySummary) float64 {
	if len(dailies) == 0 {
		return 0.0
	}
	avg := 0.0
	for _, dailySummary := range dailies {
		avg += dailySummary.FTERatio
	}
	return avg / float64(len(dailies))
}

func (r *ReportBuilder) reduceLeavesToBlocks(leaves []model.Leave) []AbsenceBlock {
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

func (r *ReportBuilder) prepareDays() ([]*DailySummary, error) {
	days := make([]*DailySummary, 0)

	firstDay := r.from
	lastDay := r.to

	if r.clampToNow && lastDay.After(now().In(r.timezone)) {
		lastDay = r.getDateTomorrow()
	}

	for currentDay := firstDay; currentDay.Before(lastDay); currentDay = currentDay.AddDate(0, 0, 1) {
		currentRatio, err := r.contracts.GetFTERatioForDay(odoo.Date(currentDay))
		if err != nil {
			return days, err
		}
		days = append(days, NewDailySummary(currentRatio, currentDay.In(r.timezone)))
	}

	return days, nil
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

func (r *ReportBuilder) filterLeavesInTimeRange() []model.Leave {
	filteredLeaves := make([]model.Leave, 0)
	for _, leave := range r.leaves.Items {
		splits := leave.SplitByDay()
		for _, split := range splits {
			date := split.DateFrom.WithLocation(r.timezone)
			if date.IsWithinTimeRange(r.from, r.to) && date.ToTime().Weekday() != time.Sunday && date.ToTime().Weekday() != time.Saturday {
				filteredLeaves = append(filteredLeaves, split)
			}
		}
	}
	return filteredLeaves
}
