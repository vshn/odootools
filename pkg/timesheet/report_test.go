package timesheet

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vshn/odootools/pkg/odoo"
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
			r := ReportBuilder{
				year:  2021,
				month: 2,
			}
			r.addAttendanceShiftsToDailies(tt.givenShifts, tt.givenDailySummaries)

			assert.Equal(t, tt.expectedDailySummaries, tt.givenDailySummaries)
		})
	}
}

func TestReporter_ReduceAttendancesToShifts(t *testing.T) {
	tests := map[string]struct {
		givenAttendances []odoo.Attendance
		expectedShifts   []AttendanceShift
	}{
		"GivenAttendancesInUTC_WhenReducing_ThenApplyLocalZone": {
			givenAttendances: []odoo.Attendance{
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
			givenAttendances: []odoo.Attendance{
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
			r := ReportBuilder{
				year:     2021,
				month:    2,
				timezone: localzone(t),
			}
			result := r.reduceAttendancesToShifts(tt.givenAttendances)

			assert.Equal(t, tt.expectedShifts, result)
		})
	}
}

func TestReporter_prepareWorkDays(t *testing.T) {
	tests := map[string]struct {
		givenYear    int
		givenMonth   int
		expectedDays []*DailySummary
		nowF         func() time.Time
	}{
		"GivenFullMonthInThePast_ThenReturnOnlyWorkingDays": {
			givenYear:  2021,
			givenMonth: 5,
			expectedDays: []*DailySummary{
				{Date: *date(t, "2021-05-01")},
				{Date: *date(t, "2021-05-02")},
				{Date: *date(t, "2021-05-03")},
				{Date: *date(t, "2021-05-04")},
				{Date: *date(t, "2021-05-05")},
				{Date: *date(t, "2021-05-06")},
				{Date: *date(t, "2021-05-07")},
				{Date: *date(t, "2021-05-08")},
				{Date: *date(t, "2021-05-09")},
				{Date: *date(t, "2021-05-10")},
				{Date: *date(t, "2021-05-11")},
				{Date: *date(t, "2021-05-12")},
				{Date: *date(t, "2021-05-13")},
				{Date: *date(t, "2021-05-14")},
				{Date: *date(t, "2021-05-15")},
				{Date: *date(t, "2021-05-16")},
				{Date: *date(t, "2021-05-17")},
				{Date: *date(t, "2021-05-18")},
				{Date: *date(t, "2021-05-19")},
				{Date: *date(t, "2021-05-20")},
				{Date: *date(t, "2021-05-21")},
				{Date: *date(t, "2021-05-22")},
				{Date: *date(t, "2021-05-23")},
				{Date: *date(t, "2021-05-24")},
				{Date: *date(t, "2021-05-25")},
				{Date: *date(t, "2021-05-26")},
				{Date: *date(t, "2021-05-27")},
				{Date: *date(t, "2021-05-28")},
				{Date: *date(t, "2021-05-29")},
				{Date: *date(t, "2021-05-30")},
				{Date: *date(t, "2021-05-31")},
			},
		},
		"GivenCurrentMonth_ThenReturnNoMoreThanToday": {
			givenYear:  2021,
			givenMonth: 3,
			expectedDays: []*DailySummary{
				{Date: *date(t, "2021-03-01")},
				{Date: *date(t, "2021-03-02")},
				{Date: *date(t, "2021-03-03")},
			},
			nowF: func() time.Time {
				return time.Unix(1614788672, 0) // Wednesday, March 3, 2021 4:24:32 PM
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

			r := &ReportBuilder{
				year:     tt.givenYear,
				month:    tt.givenMonth,
				timezone: localzone(t),
			}
			result := r.prepareDays()
			require.Len(t, result, len(tt.expectedDays))
			for i := range result {
				assert.Equal(t, tt.expectedDays[i].Date, result[i].Date)
			}
		})
	}
}
