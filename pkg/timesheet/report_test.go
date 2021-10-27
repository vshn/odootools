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
	tm, err := time.Parse(odoo.DateFormat, date)
	require.NoError(t, err)
	return &tm
}

func TestReduceAttendanceBlocks(t *testing.T) {
	tests := map[string]struct {
		givenDailySummaries    []*DailySummary
		givenBlocks            []AttendanceBlock
		expectedDailySummaries []*DailySummary
	}{
		"GivenBlocksWithDifferentDates_ThenSeparateDaily": {
			givenDailySummaries: []*DailySummary{
				{Date: *date(t, "2021-02-03")},
				{Date: *date(t, "2021-02-04")},
			},
			givenBlocks: []AttendanceBlock{
				{Start: parse(t, "2021-02-03 09:00"), End: parse(t, "2021-02-03 18:00")},
				{Start: parse(t, "2021-02-04 09:00"), End: parse(t, "2021-02-04 12:00")},
				{Start: parse(t, "2021-02-04 13:00"), End: parse(t, "2021-02-04 19:00")},
			},
			expectedDailySummaries: []*DailySummary{
				{
					Date: *date(t, "2021-02-03"),
					Blocks: []AttendanceBlock{
						{Start: parse(t, "2021-02-03 09:00"), End: parse(t, "2021-02-03 18:00")},
					},
				},
				{
					Date: *date(t, "2021-02-04"),
					Blocks: []AttendanceBlock{
						{Start: parse(t, "2021-02-04 09:00"), End: parse(t, "2021-02-04 12:00")},
						{Start: parse(t, "2021-02-04 13:00"), End: parse(t, "2021-02-04 19:00")},
					},
				},
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			r := Reporter{
				year:  2021,
				month: 2,
			}
			r.addAttendanceBlocksToDailies(tt.givenBlocks, tt.givenDailySummaries)

			assert.Equal(t, tt.expectedDailySummaries, tt.givenDailySummaries)
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
			givenYear:  time.Now().Year(),
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

			r := &Reporter{
				year:  tt.givenYear,
				month: tt.givenMonth,
			}
			result := r.prepareWorkdays()
			assert.Equal(t, tt.expectedDays, result)
		})
	}
}
