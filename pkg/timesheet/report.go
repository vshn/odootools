package timesheet

import (
	"fmt"
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

// DefaultTimeZone is the zone to which apply a last-resort default.
var DefaultTimeZone *time.Location

type AttendanceShift struct {
	Start model.Attendance
	End   model.Attendance
}

// String implements fmt.Stringer.
func (s *AttendanceShift) String() string {
	if s == nil {
		return ""
	}
	return fmt.Sprintf("AttendanceShift[Start: %s, End: %s, Duration, %s, Reason: %s]", s.Start.DateTime, s.End.DateTime, s.Duration(), s.Start.Reason)
}

// Duration returns the difference between AttendanceShift.Start and AttendanceShift.End.
func (s *AttendanceShift) Duration() time.Duration {
	if s == nil {
		return 0
	}
	return s.End.DateTime.Sub(s.Start.DateTime.Time)
}

type AbsenceBlock struct {
	Date   time.Time
	Reason string
}

type Summary struct {
	TotalOvertime        time.Duration
	TotalExcusedTime     time.Duration
	TotalWorkedTime      time.Duration
	TotalOutOfOfficeTime time.Duration
	// TotalLeave is the amount of paid leave days.
	// This value respects FTE ratio, e.g. in a 50% ratio a public holiday is still counted as '1d'.
	TotalLeave      float64
	AverageWorkload float64
}

type Report struct {
	DailySummaries []*DailySummary
	Summary        Summary
	Employee       model.Employee
	// From is the first day (inclusive) of the time range.
	From time.Time
	// To is the last day (inclusive) of the time range
	To time.Time
}

type ReportBuilder struct {
	attendances model.AttendanceList
	leaves      odoo.List[model.Leave]
	employee    model.Employee
	from        time.Time
	to          time.Time
	contracts   model.ContractList
	clampToNow  bool
	clock       func() time.Time
}

func NewReporter(attendances model.AttendanceList, leaves odoo.List[model.Leave], employee model.Employee, contracts model.ContractList) *ReportBuilder {
	return &ReportBuilder{
		attendances: attendances,
		leaves:      leaves,
		employee:    employee,
		contracts:   contracts,
		clampToNow:  true,
		clock:       time.Now,
	}
}

// SkipClampingToNow ignores the current time when preparing the daily summaries within the time range.
// By default, the reporter doesn't include days that are happening in the future and thus calculate overtime wrongly.
func (r *ReportBuilder) SkipClampingToNow(skip bool) *ReportBuilder {
	r.clampToNow = !skip
	return r
}

// CalculateReport creates a report with the given information in the builder.
// For example, to build a monthly report, set `from` to the first day of the month from midnight, and `to` to the first day of the next month at midnight.
func (r *ReportBuilder) CalculateReport(from time.Time, to time.Time) (Report, error) {
	r.from = from
	r.to = to
	filteredAttendances := r.attendances.FilterAttendanceBetweenDates(r.from, r.to.Add(-1*time.Second))
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

	r.addAttendancesToDailyShifts(filteredAttendances, dailySummaries)
	r.addAbsencesToDailies(absences, dailySummaries)

	summary := Summary{}
	for _, dailySummary := range dailySummaries {
		overtimeSummary := dailySummary.CalculateOvertimeSummary()
		summary.TotalOvertime += overtimeSummary.Overtime()
		summary.TotalExcusedTime += overtimeSummary.ExcusedTime()
		summary.TotalWorkedTime += overtimeSummary.WorkingTime()
		summary.TotalOutOfOfficeTime += overtimeSummary.OutOfOfficeTime
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

func (r *ReportBuilder) getTimeZone() *time.Location {
	return r.from.Location()
}

func (r *ReportBuilder) addAttendancesToDailyShifts(attendances model.AttendanceList, dailies []*DailySummary) {
	monthTz := r.getTimeZone()
	dailyMap := make(map[string]*DailySummary, len(dailies))
	for _, dailySummary := range dailies {
		dailyMap[dailySummary.Date.Format(odoo.DateFormat)] = dailySummary
	}

	for _, attendance := range attendances.Items {
		tz := attendance.Timezone.LocationOrDefault(monthTz)
		date := attendance.DateTime.In(tz)
		daily, exists := dailyMap[date.Format(odoo.DateFormat)]
		if !exists {
			continue // irrelevant attendance
		}
		daily.Date = daily.Date.In(tz) // Update the timezone of the day
		var shift AttendanceShift
		shiftCount := len(daily.Shifts)
		newShift := false
		if shiftCount == 0 {
			shift = AttendanceShift{}
			newShift = true
		} else {
			shift = daily.Shifts[shiftCount-1]
		}
		startDate := shift.Start.DateTime.Time
		endDate := shift.End.DateTime.Time
		if !endDate.IsZero() && (date.Equal(endDate) || date.After(endDate)) {
			// new shift
			shift = AttendanceShift{}
			newShift = true
		}
		if !startDate.IsZero() && (attendance.Action == model.ActionSignIn) {
			// start of shift already defined, which means we have 2 consecutive sign_ins.
			// This is semantically invalid.
			shift = AttendanceShift{}
			newShift = true
		}
		attendance.DateTime.Time = date // correct timezone
		switch attendance.Action {
		case model.ActionSignIn:
			shift.Start = attendance
		case model.ActionSignOut:
			shift.End = attendance
		}
		if newShift {
			daily.Shifts = append(daily.Shifts, shift)
		} else {
			daily.Shifts[shiftCount-1] = shift
		}
	}
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
			from := leave.DateFrom
			blocks = append(blocks, AbsenceBlock{
				Reason: leave.Type.String(),
				Date:   odoo.Midnight(from.Time),
			})
		}
	}
	return blocks
}

func (r *ReportBuilder) prepareDays() ([]*DailySummary, error) {
	days := make([]*DailySummary, 0)

	firstDay := r.from
	lastDay := r.to

	tz := r.getTimeZone()
	now := r.clock().In(tz)
	if r.clampToNow && lastDay.After(now) {
		lastDay = r.getDateTomorrow()
	}

	contractStartDate := odoo.LocalizeTime(r.contracts.GetEarliestStartContractDate(), tz)
	for currentDay := firstDay; currentDay.Before(lastDay); currentDay = currentDay.AddDate(0, 0, 1) {
		if currentDay.Before(contractStartDate) {
			continue
		}
		currentRatio, err := r.contracts.GetFTERatioForDay(currentDay)
		if err != nil {
			return days, err
		}
		days = append(days, NewDailySummary(currentRatio, currentDay.In(tz)))
	}

	return days, nil
}

func (r *ReportBuilder) getDateTomorrow() time.Time {
	tz := r.getTimeZone()
	now := r.clock().In(tz)
	return time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, tz)
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
			tz := r.getTimeZone()
			from := split.DateFrom
			date := odoo.Midnight(from.In(tz))
			if odoo.IsWithinTimeRange(date, r.from, r.to) {
				split.DateFrom.Time = odoo.LocalizeTime(split.DateFrom.Time, tz)
				split.DateTo.Time = odoo.LocalizeTime(split.DateTo.Time, tz)
				filteredLeaves = append(filteredLeaves, split)
			}
		}
	}
	return filteredLeaves
}
