package timesheet

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vshn/odootools/pkg/odoo"
)

func hours(t *testing.T, hours string) time.Time {
	tm, err := time.Parse(odoo.AttendanceDateTimeFormat, fmt.Sprintf("2021-02-03 %s:00", hours))
	require.NoError(t, err)
	return tm
}

func hoursDuration(t *testing.T, hours float64) time.Duration {
	dur, err := time.ParseDuration(fmt.Sprintf("%fh", hours))
	require.NoError(t, err)
	return dur
}

func parse(t *testing.T, pattern string) time.Time {
	tm, err := time.Parse(odoo.AttendanceDateTimeFormat, fmt.Sprintf("%s:00", pattern))
	require.NoError(t, err)
	return tm
}

func date(t *testing.T, date string) time.Time {
	tm, err := time.Parse(odoo.AttendanceDateFormat, date)
	require.NoError(t, err)
	return tm
}

func TestReduceAttendanceBlocks(t *testing.T) {
	tests := map[string]struct {
		givenBlocks            []AttendanceBlock
		expectedDailySummaries []*DailySummary
	}{
		"GivenBlocksWithDifferentDates_ThenSeparateDaily": {
			givenBlocks: []AttendanceBlock{
				{Start: parse(t, "2021-02-03 09:00"), End: parse(t, "2021-02-03 18:00")},
				{Start: parse(t, "2021-02-04 09:00"), End: parse(t, "2021-02-04 12:00")},
				{Start: parse(t, "2021-02-04 13:00"), End: parse(t, "2021-02-04 19:00")},
			},
			expectedDailySummaries: []*DailySummary{
				{
					Date: date(t, "2021-02-03"),
					Blocks: []AttendanceBlock{
						{Start: parse(t, "2021-02-03 09:00"), End: parse(t, "2021-02-03 18:00")},
					},
				},
				{
					Date: date(t, "2021-02-04"),
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
			result := reduceAttendanceBlocks(tt.givenBlocks, 0)
			assert.Equal(t, tt.expectedDailySummaries, result)
		})
	}
}
