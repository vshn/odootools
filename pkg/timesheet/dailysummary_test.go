package timesheet

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDailySummary_CalculateOvertime(t *testing.T) {
	tests := map[string]struct {
		givenBlocks      []AttendanceBlock
		expectedOvertime time.Duration
	}{
		"GivenSingleBlock_WhenMoreThanDailyMax_ThenReturnOvertime": {
			givenBlocks: []AttendanceBlock{
				{Start: hours(t, "09:00"), End: hours(t, "18:00")},
			},
			expectedOvertime: hoursDuration(t, 1),
		},
		"GivenMultipleBlocks_WhenMoreThanDailyMax_ThenReturnOvertime": {
			givenBlocks: []AttendanceBlock{
				{Start: hours(t, "09:00"), End: hours(t, "12:00")},
				{Start: hours(t, "13:00"), End: hours(t, "19:00")},
			},
			expectedOvertime: hoursDuration(t, 1),
		},
		"GivenMultipleBlocks_WhenLessThanDailyMax_ThenReturnUndertime": {
			givenBlocks: []AttendanceBlock{
				{Start: hours(t, "09:00"), End: hours(t, "12:00")},
				{Start: hours(t, "13:00"), End: hours(t, "17:00")},
			},
			expectedOvertime: hoursDuration(t, -1),
		},
		"GivenSickLeaveBlocks_WhenSickLeaveIsFilling_ThenReturnZero": {
			givenBlocks: []AttendanceBlock{
				{Start: hours(t, "09:00"), End: hours(t, "12:00")},
				{Start: hours(t, "13:00"), End: hours(t, "18:00"), Reason: ReasonSickLeave},
			},
			expectedOvertime: hoursDuration(t, 0),
		},
		"GivenSickLeaveBlocks_WhenSickLeaveIsLessThanDailyMax_ThenReturnUndertime": {
			givenBlocks: []AttendanceBlock{
				{Start: hours(t, "09:00"), End: hours(t, "12:00")},
				{Start: hours(t, "13:00"), End: hours(t, "17:00"), Reason: ReasonSickLeave},
			},
			expectedOvertime: hoursDuration(t, -1),
		},
		"GivenSickLeaveBlocks_WhenCombinedHoursExceedDailyMax_ThenCapOvertime": {
			givenBlocks: []AttendanceBlock{
				{Start: hours(t, "09:00"), End: hours(t, "12:00")},
				{Start: hours(t, "13:00"), End: hours(t, "18:30"), Reason: ReasonSickLeave},
			},
			expectedOvertime: hoursDuration(t, 0),
		},
		"GivenSickLeaveBlocks_WhenWorkingHoursIsMoreThanDailyMax_ThenIgnoreSickLeaveCompletely": {
			givenBlocks: []AttendanceBlock{
				{Start: hours(t, "09:00"), End: hours(t, "18:00")},
				{Start: hours(t, "19:00"), End: hours(t, "20:00"), Reason: ReasonSickLeave},
			},
			expectedOvertime: hoursDuration(t, 1),
		},
		"GivenSickLeaveAndOutsideOfficeHoursBlocks_WhenWorkingHoursIsMoreThanDailyMax_ThenIgnoreSickLeaveCompletely": {
			givenBlocks: []AttendanceBlock{
				{Start: hours(t, "09:00"), End: hours(t, "18:00")},                                   // 1h overtime
				{Start: hours(t, "19:00"), End: hours(t, "20:00"), Reason: ReasonSickLeave},          // no overtime
				{Start: hours(t, "20:00"), End: hours(t, "22:00"), Reason: ReasonOutsideOfficeHours}, // 3h overtime
			},
			expectedOvertime: hoursDuration(t, 4),
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			s := DailySummary{
				Blocks: tt.givenBlocks,
				FTERatio: 1,
			}
			result := s.CalculateOvertime()
			assert.Equal(t, tt.expectedOvertime, result)
		})
	}
}
