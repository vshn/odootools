package timesheet

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDailySummary_CalculateOvertime(t *testing.T) {
	weekday := "2021-02-03"
	weekendDay := "2021-02-06"
	tests := map[string]struct {
		givenBlocks      []AttendanceBlock
		givenDate        *time.Time
		expectedOvertime time.Duration
	}{
		"GivenSingleBlock_WhenMoreThanDailyMax_ThenReturnOvertime": {
			givenBlocks: []AttendanceBlock{
				{Start: hours(t, weekday, "09:00"), End: hours(t, weekday, "18:00")},
			},
			expectedOvertime: hoursDuration(t, 1),
		},
		"GivenMultipleBlocks_WhenMoreThanDailyMax_ThenReturnOvertime": {
			givenBlocks: []AttendanceBlock{
				{Start: hours(t, weekday, "09:00"), End: hours(t, weekday, "12:00")},
				{Start: hours(t, weekday, "13:00"), End: hours(t, weekday, "19:00")},
			},
			expectedOvertime: hoursDuration(t, 1),
		},
		"GivenMultipleBlocks_WhenLessThanDailyMax_ThenReturnUndertime": {
			givenBlocks: []AttendanceBlock{
				{Start: hours(t, weekday, "09:00"), End: hours(t, weekday, "12:00")},
				{Start: hours(t, weekday, "13:00"), End: hours(t, weekday, "17:00")},
			},
			expectedOvertime: hoursDuration(t, -1),
		},
		"GivenSickLeaveBlocks_WhenSickLeaveIsFilling_ThenReturnZero": {
			givenBlocks: []AttendanceBlock{
				{Start: hours(t, weekday, "09:00"), End: hours(t, weekday, "12:00")},
				{Start: hours(t, weekday, "13:00"), End: hours(t, weekday, "18:00"), Reason: ReasonSickLeave},
			},
			expectedOvertime: hoursDuration(t, 0),
		},
		"GivenSickLeaveBlocks_WhenSickLeaveIsLessThanDailyMax_ThenReturnUndertime": {
			givenBlocks: []AttendanceBlock{
				{Start: hours(t, weekday, "09:00"), End: hours(t, weekday, "12:00")},
				{Start: hours(t, weekday, "13:00"), End: hours(t, weekday, "17:00"), Reason: ReasonSickLeave},
			},
			expectedOvertime: hoursDuration(t, -1),
		},
		"GivenSickLeaveBlocks_WhenCombinedHoursExceedDailyMax_ThenCapOvertime": {
			givenBlocks: []AttendanceBlock{
				{Start: hours(t, weekday, "09:00"), End: hours(t, weekday, "12:00")},
				{Start: hours(t, weekday, "13:00"), End: hours(t, weekday, "18:30"), Reason: ReasonSickLeave},
			},
			expectedOvertime: hoursDuration(t, 0),
		},
		"GivenSickLeaveBlocks_WhenWorkingHoursIsMoreThanDailyMax_ThenIgnoreSickLeaveCompletely": {
			givenBlocks: []AttendanceBlock{
				{Start: hours(t, weekday, "09:00"), End: hours(t, weekday, "18:00")},
				{Start: hours(t, weekday, "19:00"), End: hours(t, weekday, "20:00"), Reason: ReasonSickLeave},
			},
			expectedOvertime: hoursDuration(t, 1),
		},
		"GivenSickLeaveAndOutsideOfficeHoursBlocks_WhenWorkingHoursIsMoreThanDailyMax_ThenIgnoreSickLeaveCompletely": {
			givenBlocks: []AttendanceBlock{
				{Start: hours(t, weekday, "09:00"), End: hours(t, weekday, "18:00")},                                   // 1h overtime
				{Start: hours(t, weekday, "19:00"), End: hours(t, weekday, "20:00"), Reason: ReasonSickLeave},          // no overtime
				{Start: hours(t, weekday, "20:00"), End: hours(t, weekday, "22:00"), Reason: ReasonOutsideOfficeHours}, // 3h overtime
			},
			expectedOvertime: hoursDuration(t, 4),
		},
		"GivenNoBlocks_WhenNoLeavesEither_ThenReturnOneDayUndertime": {
			givenBlocks:      []AttendanceBlock{},
			expectedOvertime: hoursDuration(t, -8),
		},
		"GivenDateInWeekend_WhenNoWorkingHours_ThenReturnNoOvertime": {
			givenBlocks:      []AttendanceBlock{},
			givenDate:        date(t, weekendDay),
			expectedOvertime: hoursDuration(t, 0),
		},
		"GivenDateInWeekend_WhenWorkingHoursLogged_ThenReturnOvertime": {
			givenBlocks: []AttendanceBlock{
				{Start: hours(t, weekday, "09:00"), End: hours(t, weekday, "10:00")}, // 1h overtime
			},
			givenDate:        date(t, weekendDay),
			expectedOvertime: hoursDuration(t, 1),
		},
		"GivenDateInWeekend_WhenExcusedHoursLogged_ThenIgnoreExcusedHours": {
			givenBlocks: []AttendanceBlock{
				{Start: hours(t, weekendDay, "09:00"), End: hours(t, weekendDay, "10:00")},                              // 1h overtime
				{Start: hours(t, weekendDay, "09:00"), End: hours(t, weekendDay, "10:00"), Reason: ReasonPublicService}, // no overtime
			},
			givenDate:        date(t, weekendDay),
			expectedOvertime: hoursDuration(t, 1),
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			d := tt.givenDate
			if d == nil {
				d = date(t, weekday)
			}
			s := DailySummary{
				Date:     *d,
				Blocks:   tt.givenBlocks,
				FTERatio: 1,
			}
			result := s.CalculateOvertime()
			assert.Equal(t, tt.expectedOvertime, result)
		})
	}
}

func TestDailySummary_CalculateDailyMaxHours(t *testing.T) {
	tests := map[string]struct {
		givenDate     *time.Time
		givenFteRatio float64
		givenAbsences []AbsenceBlock
		expectedHours float64
	}{
		"GivenWeekDay_WhenIn2021_ThenReturn8Hours": {
			givenDate:     date(t, "2021-02-03"),
			givenFteRatio: float64(1),
			expectedHours: 8,
		},
		"GivenWeekDay_WhenIn2020_ThenReturn8.5Hours": {
			givenDate:     date(t, "2020-02-03"),
			givenFteRatio: float64(1),
			expectedHours: 8.5,
		},
		"GivenWeekendDay_ThenReturn0Hours": {
			givenDate:     date(t, "2021-02-06"),
			givenFteRatio: float64(1),
			expectedHours: 0,
		},
		"GivenAbsences_WhenTypeIsHoliday_ThenReturn0Hours": {
			givenDate:     date(t, "2021-02-06"),
			givenFteRatio: float64(1),
			givenAbsences: []AbsenceBlock{
				{Reason: TypePublicHoliday},
			},
			expectedHours: 0,
		},
		"GivenAbsences_WhenTypeIsUnpaid_ThenReturnNormalHours": {
			givenDate:     date(t, "2021-02-03"),
			givenFteRatio: float64(1),
			givenAbsences: []AbsenceBlock{
				{Reason: TypeUnpaid},
			},
			expectedHours: 8,
		},
		"GivenAbsencesWithSpecialFte_WhenTypeIsUnpaid_ThenReturnFteAdjustedHours": {
			givenDate:     date(t, "2021-02-03"),
			givenFteRatio: 0.6,
			givenAbsences: []AbsenceBlock{
				{Reason: TypeUnpaid},
			},
			expectedHours: 4.8,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			s := DailySummary{
				Date:     *tt.givenDate,
				FTERatio: tt.givenFteRatio,
				Absences: tt.givenAbsences,
			}
			result := s.CalculateDailyMaxHours()
			assert.Equal(t, tt.expectedHours, result)
		})
	}
}
