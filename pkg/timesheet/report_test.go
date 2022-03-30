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

func parse(t *testing.T, pattern string) time.Time {
	tm, err := time.Parse(odoo.DateTimeFormat, fmt.Sprintf("%s:00", pattern))
	require.NoError(t, err)
	return tm
}

func date(t *testing.T, date string) *time.Time {
	zone, err := time.LoadLocation("Europe/Zurich")
	require.NoError(t, err)
	tm, err := time.Parse(odoo.DateFormat, date)
	require.NoError(t, err)
	tmzone := tm.In(zone)
	return &tmzone
}

func newDateTime(t *testing.T, value string) *odoo.Date {
	tm, err := time.Parse(odoo.DateTimeFormat, fmt.Sprintf("%s:00", value))
	require.NoError(t, err)
	ptr := odoo.Date(tm)
	return &ptr
}

func localzone(t *testing.T) *time.Location {
	zone, err := time.LoadLocation("Europe/Zurich")
	require.NoError(t, err)
	return zone
}

func TestReporter_AddAttendanceShiftsToDailies(t *testing.T) {
	tests := map[string]struct {
		givenDailySummaries    []*DailySummary
		givenShifts            []AttendanceShift
		expectedDailySummaries []*DailySummary
	}{
		"GivenShiftsWithDifferentDates_ThenSeparateDaily": {
			givenDailySummaries: []*DailySummary{
				{Date: *date(t, "2021-02-03")},
				{Date: *date(t, "2021-02-04")},
			},
			givenShifts: []AttendanceShift{
				{Start: parse(t, "2021-02-03 09:00"), End: parse(t, "2021-02-03 18:00")},
				{Start: parse(t, "2021-02-04 09:00"), End: parse(t, "2021-02-04 12:00")},
				{Start: parse(t, "2021-02-04 13:00"), End: parse(t, "2021-02-04 19:00")},
			},
			expectedDailySummaries: []*DailySummary{
				{
					Date: *date(t, "2021-02-03"),
					Shifts: []AttendanceShift{
						{Start: parse(t, "2021-02-03 09:00"), End: parse(t, "2021-02-03 18:00")},
					},
				},
				{
					Date: *date(t, "2021-02-04"),
					Shifts: []AttendanceShift{
						{Start: parse(t, "2021-02-04 09:00"), End: parse(t, "2021-02-04 12:00")},
						{Start: parse(t, "2021-02-04 13:00"), End: parse(t, "2021-02-04 19:00")},
					},
				},
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			start := time.Date(2021, time.February, 1, 0, 0, 0, 0, time.UTC)
			end := start.AddDate(0, 1, 0)
			r := ReportBuilder{
				from:     start,
				to:       end,
				timezone: localzone(t),
			}
			r.addAttendanceShiftsToDailies(tt.givenShifts, tt.givenDailySummaries)

			assert.Equal(t, tt.expectedDailySummaries, tt.givenDailySummaries)
		})
	}
}

func TestReporter_ReduceAttendancesToShifts(t *testing.T) {
	tests := map[string]struct {
		givenAttendances []model.Attendance
		expectedShifts   []AttendanceShift
	}{
		"GivenAttendancesInUTC_WhenReducing_ThenApplyLocalZone": {
			givenAttendances: []model.Attendance{
				{DateTime: newDateTime(t, "2021-02-03 19:00"), Action: ActionSignIn}, // these times are UTC
				{DateTime: newDateTime(t, "2021-02-03 22:59"), Action: ActionSignOut},
			},
			expectedShifts: []AttendanceShift{
				{Start: newDateTime(t, "2021-02-03 19:00").ToTime().In(localzone(t)),
					End: newDateTime(t, "2021-02-03 22:59").ToTime().In(localzone(t)),
				},
			},
		},
		"GivenAttendancesInUTC_WhenSplitOverMidnight_ThenSplitInTwoDays": {
			givenAttendances: []model.Attendance{
				{DateTime: newDateTime(t, "2021-02-03 19:00"), Action: ActionSignIn}, // these times are UTC
				{DateTime: newDateTime(t, "2021-02-03 22:59"), Action: ActionSignOut},
				{DateTime: newDateTime(t, "2021-02-03 23:00"), Action: ActionSignIn},
				{DateTime: newDateTime(t, "2021-02-04 00:00"), Action: ActionSignOut},
			},
			expectedShifts: []AttendanceShift{
				{
					Start: newDateTime(t, "2021-02-03 19:00").ToTime().In(localzone(t)),
					End:   newDateTime(t, "2021-02-03 22:59").ToTime().In(localzone(t)),
				},
				{
					Start: newDateTime(t, "2021-02-03 23:00").ToTime().In(localzone(t)),
					End:   newDateTime(t, "2021-02-04 00:00").ToTime().In(localzone(t)),
				},
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			start := time.Date(2021, time.February, 1, 0, 0, 0, 0, time.UTC)
			end := start.AddDate(0, 1, 0)
			r := ReportBuilder{
				from:     start,
				to:       end,
				timezone: localzone(t),
			}
			list := model.AttendanceList{Items: tt.givenAttendances}
			result := r.reduceAttendancesToShifts(list)

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
		nowF           func() time.Time
	}{
		"GivenFullMonthInThePast_ThenReturnAllDays": {
			givenYear:  2021,
			givenMonth: 5,
			givenContracts: []model.Contract{
				{Start: odoo.MustParseDate("2021-01-01"), WorkingSchedule: &model.WorkingSchedule{Name: "100%"}},
			},
			expectedDays: generateMonth(t, 2021, 5, 31),
		},
		"GivenNoContracts_ThenExpectError": {
			givenYear:     2021,
			givenMonth:    5,
			expectedDays:  generateMonth(t, 2021, 5, 31),
			expectedError: "no contract found that covers date: 2021-05-01 00:00:00",
		},
		"GivenCurrentMonth_ThenReturnNoMoreThanToday": {
			givenYear:  2021,
			givenMonth: 3,
			givenContracts: []model.Contract{
				{Start: odoo.MustParseDate("2021-01-01"), WorkingSchedule: &model.WorkingSchedule{Name: "100%"}},
			},
			expectedDays: generateMonth(t, 2021, 3, 7),
			nowF: func() time.Time {
				return time.Unix(1615113136, 0) // Sunday, March 7, 2021 10:32:16
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if tt.nowF != nil {
				// Set fixed clock if given
				currentF := now
				now = tt.nowF
				defer func() {
					now = currentF
				}()
			}

			start := time.Date(tt.givenYear, time.Month(tt.givenMonth), 1, 0, 0, 0, 0, time.UTC)
			end := start.AddDate(0, 1, 0)
			r := &ReportBuilder{
				from:      start,
				to:        end,
				timezone:  localzone(t),
				contracts: model.ContractList{Items: tt.givenContracts},
			}
			result, err := r.prepareDays()
			if tt.expectedError != "" {
				assert.EqualError(t, err, tt.expectedError)
				return
			}
			require.NoError(t, err)
			require.Len(t, result, len(tt.expectedDays))
			for i := range result {
				assert.Equal(t, tt.expectedDays[i].Date, result[i].Date)
			}
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

func generateMonth(t *testing.T, year, month, lastDay int) []*DailySummary {
	days := make([]*DailySummary, lastDay)
	for i := 0; i < lastDay; i++ {
		days[i] = &DailySummary{Date: *date(t, fmt.Sprintf("%d-%02d-%02d", year, month, i+1))}
	}
	return days
}
