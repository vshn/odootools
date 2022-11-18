package timesheet

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vshn/odootools/pkg/odoo"
	"github.com/vshn/odootools/pkg/odoo/model"
)

var (
	zurichTZ    *time.Location
	vancouverTZ *time.Location
)

func init() {
	zue, err := time.LoadLocation("Europe/Zurich")
	if err != nil {
		panic(err)
	}
	zurichTZ = zue
	van, err := time.LoadLocation("America/Vancouver")
	if err != nil {
		panic(err)
	}
	vancouverTZ = van
}

func hours(t *testing.T, date, hours string) time.Time {
	tm, err := time.Parse(odoo.DateTimeFormat, fmt.Sprintf("%s %s:00", date, hours))
	require.NoError(t, err)
	return tm
}

func hoursDuration(t *testing.T, hours float64) time.Duration {
	dur, err := time.ParseDuration(fmt.Sprintf("%fh", hours))
	require.NoError(t, err)
	return dur
}

func TestReporter_AddAttendanceShiftsToDailies(t *testing.T) {
	tests := map[string]struct {
		givenDailySummaries    []*DailySummary
		givenShifts            []AttendanceShift
		expectedDailySummaries []*DailySummary
	}{
		"GivenShiftsWithDifferentDates_ThenSeparateDaily": {
			givenDailySummaries: []*DailySummary{
				{Date: odoo.MustParseDate("2021-02-03").Time},
				{Date: odoo.MustParseDate("2021-02-04").Time},
			},
			givenShifts: []AttendanceShift{
				{Start: odoo.MustParseDateTime("2021-02-03 09:00:00").Time, End: odoo.MustParseDateTime("2021-02-03 18:00:00").Time},
				{Start: odoo.MustParseDateTime("2021-02-04 09:00:00").Time, End: odoo.MustParseDateTime("2021-02-04 12:00:00").Time},
				{Start: odoo.MustParseDateTime("2021-02-04 13:00:00").Time, End: odoo.MustParseDateTime("2021-02-04 19:00:00").Time},
			},
			expectedDailySummaries: []*DailySummary{
				{
					Date: odoo.MustParseDate("2021-02-03").Time,
					Shifts: []AttendanceShift{
						{Start: odoo.MustParseDateTime("2021-02-03 09:00:00").Time, End: odoo.MustParseDateTime("2021-02-03 18:00:00").Time},
					},
				},
				{
					Date: odoo.MustParseDate("2021-02-04").Time,
					Shifts: []AttendanceShift{
						{Start: odoo.MustParseDateTime("2021-02-04 09:00:00").Time, End: odoo.MustParseDateTime("2021-02-04 12:00:00").Time},
						{Start: odoo.MustParseDateTime("2021-02-04 13:00:00").Time, End: odoo.MustParseDateTime("2021-02-04 19:00:00").Time},
					},
				},
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			start := time.Date(2021, time.February, 1, 0, 0, 0, 0, zurichTZ)
			end := start.AddDate(0, 1, 0)
			r := ReportBuilder{
				from: start,
				to:   end,
			}
			r.addAttendanceShiftsToDailies(tt.givenShifts, tt.givenDailySummaries)

			assert.Equal(t, tt.expectedDailySummaries, tt.givenDailySummaries)
		})
	}
}

func TestReporter_ReduceAttendancesToShifts(t *testing.T) {
	tests := map[string]struct {
		givenAttendances []model.Attendance
		givenTimeZone    *time.Location
		expectedShifts   []AttendanceShift
	}{
		"GivenAttendancesInZurich_WhenReducing_ThenApplyLocalZone": {
			givenTimeZone: zurichTZ,
			givenAttendances: []model.Attendance{
				// these times are UTC
				{DateTime: odoo.MustParseDateTime("2021-02-03 19:00:00"), Action: model.ActionSignIn},
				{DateTime: odoo.MustParseDateTime("2021-02-03 22:59:00"), Action: model.ActionSignOut},
			},
			expectedShifts: []AttendanceShift{
				{Start: odoo.MustParseDateTime("2021-02-03 19:00:00").In(zurichTZ),
					End: odoo.MustParseDateTime("2021-02-03 22:59:00").In(zurichTZ),
				},
			},
		},
		"GivenAttendancesInZurich_WhenSplitOverMidnight_ThenSplitInTwoDays": {
			givenTimeZone: zurichTZ,
			givenAttendances: []model.Attendance{
				// these times are UTC
				{DateTime: odoo.MustParseDateTime("2021-02-03 19:00:00"), Action: model.ActionSignIn},
				{DateTime: odoo.MustParseDateTime("2021-02-03 22:59:00"), Action: model.ActionSignOut},
				{DateTime: odoo.MustParseDateTime("2021-02-03 23:00:00"), Action: model.ActionSignIn},
				{DateTime: odoo.MustParseDateTime("2021-02-04 00:00:00"), Action: model.ActionSignOut},
			},
			expectedShifts: []AttendanceShift{
				{
					Start: odoo.MustParseDateTime("2021-02-03 19:00:00").In(zurichTZ),
					End:   odoo.MustParseDateTime("2021-02-03 22:59:00").In(zurichTZ),
				},
				{
					Start: odoo.MustParseDateTime("2021-02-03 23:00:00").In(zurichTZ),
					End:   odoo.MustParseDateTime("2021-02-04 00:00:00").In(zurichTZ),
				},
			},
		},
		"GivenAttendancesInVancouver_ThenSplitCorrectly": {
			givenTimeZone: vancouverTZ,
			givenAttendances: []model.Attendance{
				// these times are UTC
				{DateTime: odoo.MustParseDateTime("2021-02-03 15:00:00"), Action: model.ActionSignIn},
				{DateTime: odoo.MustParseDateTime("2021-02-03 19:00:00"), Action: model.ActionSignOut},
				{DateTime: odoo.MustParseDateTime("2021-02-03 20:00:00"), Action: model.ActionSignIn},
				{DateTime: odoo.MustParseDateTime("2021-02-04 01:00:00"), Action: model.ActionSignOut},
			},
			expectedShifts: []AttendanceShift{
				{
					Start: odoo.MustParseDateTime("2021-02-03 15:00:00").In(vancouverTZ), // 2021-02-03 07:00:00 -0800 PST
					End:   odoo.MustParseDateTime("2021-02-03 19:00:00").In(vancouverTZ), // 2021-02-03 11:00:00 -0800 PST
				},
				{
					Start: odoo.MustParseDateTime("2021-02-03 20:00:00").In(vancouverTZ), // 2021-02-03 12:00:00 -0800 PST
					End:   odoo.MustParseDateTime("2021-02-04 01:00:00").In(vancouverTZ), // 2021-02-03 17:00:00 -0800 PST
				},
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			start := time.Date(2021, time.February, 1, 0, 0, 0, 0, tt.givenTimeZone)
			end := start.AddDate(0, 1, 0)
			r := ReportBuilder{
				from: start,
				to:   end,
			}
			list := model.AttendanceList{Items: tt.givenAttendances}
			result := r.reduceAttendancesToShifts(list)
			for _, shift := range result {
				t.Logf("shift start: %s, end: %s", shift.Start, shift.End)
			}

			assert.Equal(t, tt.expectedShifts, result)
		})
	}
}

func TestReporter_prepareWorkDays(t *testing.T) {
	tests := map[string]struct {
		givenYear      int
		givenMonth     int
		givenContracts []model.Contract
		expectedDays   []*DailySummary
		expectedError  string
		now            time.Time
	}{
		"GivenFullMonthInThePast_ThenReturnAllDays": {
			givenYear:  2021,
			givenMonth: 5,
			givenContracts: []model.Contract{
				{Start: odoo.MustParseDate("2021-01-01"), WorkingSchedule: &model.WorkingSchedule{Name: "100%"}},
			},
			expectedDays: generateMonth(2021, 5, 31, zurichTZ),
		},
		"GivenNoContracts_ThenExpectError": {
			givenYear:     2021,
			givenMonth:    5,
			expectedDays:  generateMonth(2021, 5, 31, zurichTZ),
			expectedError: "no contract found that covers date: 2021-05-01 00:00:00 +0200 CEST",
		},
		"GivenCurrentMonth_ThenReturnNoMoreThanToday": {
			givenYear:  2021,
			givenMonth: 3,
			givenContracts: []model.Contract{
				{Start: odoo.MustParseDate("2021-01-01"), WorkingSchedule: &model.WorkingSchedule{Name: "100%"}},
			},
			expectedDays: generateMonth(2021, 3, 7, zurichTZ),
			now:          time.Date(2021, time.March, 7, 10, 32, 16, 0, time.UTC),
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {

			start := time.Date(tt.givenYear, time.Month(tt.givenMonth), 1, 0, 0, 0, 0, zurichTZ)
			end := start.AddDate(0, 1, 0)
			r := &ReportBuilder{
				from:       start,
				to:         end,
				contracts:  model.ContractList{Items: tt.givenContracts},
				clampToNow: true,
				clock:      time.Now,
			}
			if !tt.now.IsZero() {
				r.clock = func() time.Time {
					return tt.now
				}
			}

			result, err := r.prepareDays()
			if tt.expectedError != "" {
				assert.EqualError(t, err, tt.expectedError)
				return
			}
			require.NoError(t, err)
			require.Len(t, result, len(tt.expectedDays), "count of days")
			for i := range result {
				assert.Equal(t, tt.expectedDays[i].Date, result[i].Date)
			}
		})
	}
}

func TestReportBuilder_getDateTomorrow(t *testing.T) {
	tests := map[string]struct {
		now          time.Time
		fromDate     time.Time
		expectedDate time.Time
	}{
		"UTC": {
			now:          time.Date(2021, time.February, 1, 2, 3, 4, 0, time.UTC),
			fromDate:     time.Date(2021, time.February, 1, 0, 0, 0, 0, time.UTC),
			expectedDate: time.Date(2021, time.February, 2, 0, 0, 0, 0, time.UTC),
		},
		"Zurich": {
			now:          time.Date(2021, time.February, 1, 2, 3, 4, 0, zurichTZ),
			fromDate:     time.Date(2021, time.February, 1, 0, 0, 0, 0, zurichTZ),
			expectedDate: time.Date(2021, time.February, 2, 0, 0, 0, 0, zurichTZ),
		},
		"VancouverInSameDay": {
			now:          time.Date(2021, time.February, 1, 14, 3, 4, 0, zurichTZ),
			fromDate:     time.Date(2021, time.February, 1, 0, 0, 0, 0, vancouverTZ),
			expectedDate: time.Date(2021, time.February, 2, 0, 0, 0, 0, vancouverTZ),
		},
		"VancouverInOtherDay": {
			now:          time.Date(2021, time.February, 2, 5, 3, 4, 0, zurichTZ),
			fromDate:     time.Date(2021, time.February, 1, 0, 0, 0, 0, vancouverTZ),
			expectedDate: time.Date(2021, time.February, 2, 0, 0, 0, 0, vancouverTZ),
		},
		"EndOfMonth": {
			now:          time.Date(2021, time.February, 28, 14, 3, 4, 0, zurichTZ),
			fromDate:     time.Date(2021, time.February, 1, 0, 0, 0, 0, zurichTZ),
			expectedDate: time.Date(2021, time.March, 1, 0, 0, 0, 0, zurichTZ),
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			b := ReportBuilder{from: tc.fromDate}
			b.clock = func() time.Time {
				return tc.now
			}
			result := b.getDateTomorrow()
			t.Logf("now:      %s", tc.now)
			t.Logf("from:     %s", tc.fromDate)
			t.Logf("tomorrow: %s", result)
			assert.Equal(t, tc.expectedDate, result, "date")
		})
	}
}

func TestReportBuilder_calculateAverageWorkload(t *testing.T) {
	tests := map[string]struct {
		givenDailies    []*DailySummary
		expectedAverage float64
	}{
		"GivenNoDaily_ThenExpectZero": {
			givenDailies:    []*DailySummary{},
			expectedAverage: 0.0,
		},
		"GivenSingleDaily_ThenExpectEqualValue": {
			givenDailies: []*DailySummary{
				{FTERatio: 0.7},
			},
			expectedAverage: 0.7,
		},
		"GivenMultipleDaily_ThenExpectAverageValue": {
			givenDailies: []*DailySummary{
				{FTERatio: 0.5},
				{FTERatio: 0.7},
			},
			expectedAverage: 0.6,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			builder := ReportBuilder{}
			result := builder.calculateAverageWorkload(tc.givenDailies)
			assert.InDelta(t, tc.expectedAverage, result, 0.01, "average workload")
		})
	}
}

func generateMonth(year, month, lastDay int, zone *time.Location) []*DailySummary {
	days := make([]*DailySummary, lastDay)
	for i := 0; i < lastDay; i++ {
		day := time.Date(year, time.Month(month), i+1, 0, 0, 0, 0, zone)
		days[i] = &DailySummary{Date: day}
	}
	return days
}

// This is the test that puts together a full example report.
// Should cover an 80-20 approach.
func TestReportBuilder_CalculateReport(t *testing.T) {
	givenAttendances := model.AttendanceList{Items: []model.Attendance{
		{DateTime: odoo.NewDate(2021, 1, 1, 8, 0, 0, zurichTZ), Action: model.ActionSignIn},
		{DateTime: odoo.NewDate(2021, 1, 1, 17, 0, 0, zurichTZ), Action: model.ActionSignOut}, // normal day, 1h overtime

		{DateTime: odoo.NewDate(2021, 1, 2, 20, 0, 0, zurichTZ), Action: model.ActionSignIn, Reason: &model.ActionReason{Name: ReasonOutsideOfficeHours}},
		{DateTime: odoo.NewDate(2021, 1, 2, 22, 0, 0, zurichTZ), Action: model.ActionSignOut, Reason: &model.ActionReason{Name: ReasonOutsideOfficeHours}}, // on-call, 3h overtime

		{DateTime: odoo.NewDate(2021, 1, 4, 8, 0, 0, zurichTZ), Action: model.ActionSignIn, Reason: &model.ActionReason{Name: ReasonSickLeave}},
		{DateTime: odoo.NewDate(2021, 1, 4, 16, 0, 0, zurichTZ), Action: model.ActionSignOut, Reason: &model.ActionReason{Name: ReasonSickLeave}}, // whole day with sick leave

		{DateTime: odoo.NewDate(2021, 1, 5, 8, 0, 0, zurichTZ), Action: model.ActionSignIn, Reason: &model.ActionReason{Name: ReasonSickLeave}},
		{DateTime: odoo.NewDate(2021, 1, 5, 10, 0, 0, zurichTZ), Action: model.ActionSignOut, Reason: &model.ActionReason{Name: ReasonSickLeave}}, // partially sick, count only 1h
		{DateTime: odoo.NewDate(2021, 1, 5, 10, 0, 0, zurichTZ), Action: model.ActionSignIn},
		{DateTime: odoo.NewDate(2021, 1, 5, 17, 0, 0, zurichTZ), Action: model.ActionSignOut}, // 7h worked

		{DateTime: odoo.NewDate(2021, 1, 7, 8, 0, 0, zurichTZ), Action: model.ActionSignIn},
		{DateTime: odoo.NewDate(2021, 1, 7, 17, 5, 0, zurichTZ), Action: model.ActionSignOut}, // faked signed out, still working though
	}}
	givenLeaves := odoo.List[model.Leave]{Items: []model.Leave{
		{DateFrom: odoo.NewDate(2021, 01, 06, 0, 0, 0, zurichTZ), DateTo: odoo.NewDate(2021, 01, 06, 23, 59, 0, zurichTZ), Type: &model.LeaveType{Name: TypeLegalLeavesPrefix}, State: StateApproved},
	}}
	givenEmployee := model.Employee{Name: "💃"}
	givenContracts := model.ContractList{Items: []model.Contract{
		{Start: odoo.NewDate(2021, 01, 01, 0, 0, 0, time.UTC), WorkingSchedule: &model.WorkingSchedule{Name: "100%"}},
	}}
	b := NewReporter(givenAttendances, givenLeaves, givenEmployee, givenContracts)
	now := time.Date(2021, 01, 07, 17, 5, 0, 0, zurichTZ)
	b.clock = func() time.Time {
		// fixed clock
		return now
	}
	start := time.Date(2021, 01, 01, 0, 0, 0, 0, zurichTZ)
	end := start.AddDate(0, 1, 0)
	report, err := b.CalculateReport(start, end)
	assert.NoError(t, err)
	assert.Equal(t, report.Employee.Name, givenEmployee.Name, "employee name")
	assert.Equal(t, ((9+3+7+9)*time.Hour)+(5*time.Minute), report.Summary.TotalWorkedTime, "total worked time")
	assert.Equal(t, ((1+3+1)*time.Hour)+(5*time.Minute), report.Summary.TotalOvertime, "total over time")
	assert.Equal(t, 1.0, report.Summary.TotalLeave, "total leave")
	assert.Equal(t, (8+2)*time.Hour, report.Summary.TotalExcusedTime, "total excused time")
	assert.Equal(t, 1.0, report.Summary.AverageWorkload, "average workload")
	assert.Equal(t, (2)*time.Hour, report.Summary.TotalOutOfOfficeTime, "total out of office time")
}
