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
		givenShifts      []AttendanceShift
		givenDate        *time.Time
		expectedOvertime time.Duration
	}{
		"GivenSingleShift_WhenMoreThanDailyMax_ThenReturnOvertime": {
			givenShifts: []AttendanceShift{
				{Start: hours(t, weekday, "09:00"), End: hours(t, weekday, "18:00")},
			},
			expectedOvertime: hoursDuration(t, 1),
		},
		"GivenMultipleShifts_WhenMoreThanDailyMax_ThenReturnOvertime": {
			givenShifts: []AttendanceShift{
				{Start: hours(t, weekday, "09:00"), End: hours(t, weekday, "12:00")},
				{Start: hours(t, weekday, "13:00"), End: hours(t, weekday, "19:00")},
			},
			expectedOvertime: hoursDuration(t, 1),
		},
		"GivenMultipleShifts_WhenLessThanDailyMax_ThenReturnUndertime": {
			givenShifts: []AttendanceShift{
				{Start: hours(t, weekday, "09:00"), End: hours(t, weekday, "12:00")},
				{Start: hours(t, weekday, "13:00"), End: hours(t, weekday, "17:00")},
			},
			expectedOvertime: hoursDuration(t, -1),
		},
		"GivenSickLeaveShifts_WhenSickLeaveIsFilling_ThenReturnZero": {
			givenShifts: []AttendanceShift{
				{Start: hours(t, weekday, "09:00"), End: hours(t, weekday, "12:00")},
				{Start: hours(t, weekday, "13:00"), End: hours(t, weekday, "18:00"), Reason: ReasonSickLeave},
			},
			expectedOvertime: hoursDuration(t, 0),
		},
		"GivenSickLeaveShifts_WhenSickLeaveIsLessThanDailyMax_ThenReturnUndertime": {
			givenShifts: []AttendanceShift{
				{Start: hours(t, weekday, "09:00"), End: hours(t, weekday, "12:00")},
				{Start: hours(t, weekday, "13:00"), End: hours(t, weekday, "17:00"), Reason: ReasonSickLeave},
			},
			expectedOvertime: hoursDuration(t, -1),
		},
		"GivenSickLeaveShifts_WhenCombinedHoursExceedDailyMax_ThenCapOvertime": {
			givenShifts: []AttendanceShift{
				{Start: hours(t, weekday, "09:00"), End: hours(t, weekday, "12:00")},
				{Start: hours(t, weekday, "13:00"), End: hours(t, weekday, "18:30"), Reason: ReasonSickLeave},
			},
			expectedOvertime: hoursDuration(t, 0),
		},
		"GivenSickLeaveShifts_WhenWorkingHoursIsMoreThanDailyMax_ThenIgnoreSickLeaveCompletely": {
			givenShifts: []AttendanceShift{
				{Start: hours(t, weekday, "09:00"), End: hours(t, weekday, "18:00")},
				{Start: hours(t, weekday, "19:00"), End: hours(t, weekday, "20:00"), Reason: ReasonSickLeave},
			},
			expectedOvertime: hoursDuration(t, 1),
		},
		"GivenSickLeaveAndOutsideOfficeHoursShifts_WhenWorkingHoursIsMoreThanDailyMax_ThenIgnoreSickLeaveCompletely": {
			givenShifts: []AttendanceShift{
				{Start: hours(t, weekday, "09:00"), End: hours(t, weekday, "18:00")},                                   // 1h overtime
				{Start: hours(t, weekday, "19:00"), End: hours(t, weekday, "20:00"), Reason: ReasonSickLeave},          // no overtime
				{Start: hours(t, weekday, "20:00"), End: hours(t, weekday, "22:00"), Reason: ReasonOutsideOfficeHours}, // 3h overtime
			},
			expectedOvertime: hoursDuration(t, 4),
		},
		"GivenNoShifts_WhenNoLeavesEither_ThenReturnOneDayUndertime": {
			givenShifts:      []AttendanceShift{},
			expectedOvertime: hoursDuration(t, -8),
		},
		"GivenDateInWeekend_WhenNoWorkingHours_ThenReturnNoOvertime": {
			givenShifts:      []AttendanceShift{},
			givenDate:        date(t, weekendDay),
			expectedOvertime: hoursDuration(t, 0),
		},
		"GivenDateInWeekend_WhenWorkingHoursLogged_ThenReturnOvertime": {
			givenShifts: []AttendanceShift{
				{Start: hours(t, weekday, "09:00"), End: hours(t, weekday, "10:00")}, // 1h overtime
			},
			givenDate:        date(t, weekendDay),
			expectedOvertime: hoursDuration(t, 1),
		},
		"GivenDateInWeekend_WhenExcusedHoursLogged_ThenIgnoreExcusedHours": {
			givenShifts: []AttendanceShift{
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
				Shifts:   tt.givenShifts,
				FTERatio: 1,
			}
			result := s.CalculateOvertimeSummary().Overtime()
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
			result := s.calculateDailyMax()
			assert.Equal(t, time.Duration(tt.expectedHours*float64(time.Hour)), result)
		})
	}
}

func Test_findDailySummaryByDate(t *testing.T) {
	tests := map[string]struct {
		givenDailies    []*DailySummary
		givenDate       time.Time
		expectedSummary *DailySummary
	}{
		"GivenDailies_WhenDateMatches_ThenReturnDaily": {
			givenDailies: []*DailySummary{
				NewDailySummary(1, *date(t, "2021-02-03")),
			},
			givenDate:       *date(t, "2021-02-03"),
			expectedSummary: NewDailySummary(1, *date(t, "2021-02-03")),
		},
		"GivenDailies_WhenDateMatchesInUTC_ThenReturnDaily": {
			givenDailies: []*DailySummary{
				NewDailySummary(1, date(t, "2021-02-04").UTC()),
				NewDailySummary(1, date(t, "2021-02-03").UTC()),
			},
			givenDate:       newDateTime(t, "2021-02-03 23:30").ToTime(),
			expectedSummary: NewDailySummary(1, date(t, "2021-02-03").UTC()),
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			result, _ := findDailySummaryByDate(tt.givenDailies, tt.givenDate)
			assert.Equal(t, tt.expectedSummary, result)
		})
	}
}

func TestDailySummary_IsHoliday(t *testing.T) {
	tests := map[string]struct {
		givenDay        *DailySummary
		expectedHoliday bool
	}{
		"GivenDailyWithoutAbsences_ThenReturnFalse": {
			givenDay:        &DailySummary{Date: *date(t, "2021-02-04")},
			expectedHoliday: false,
		},
		"GivenDailyWithAbsence_WhenPublicHoliday_ThenReturnFalse": {
			givenDay: &DailySummary{
				Date:     *date(t, "2021-02-04"),
				Absences: []AbsenceBlock{{Reason: TypePublicHoliday}},
			},
			expectedHoliday: false,
		},
		"GivenDailyWithAbsence_WhenPublicHolidayOnWeekend_ThenReturnFalse": {
			givenDay: &DailySummary{
				Date:     *date(t, "2021-02-06"),
				Absences: []AbsenceBlock{{Reason: TypePublicHoliday}},
			},
			expectedHoliday: false,
		},
		"GivenDailyWithAbsence_WhenUnpaid_ThenReturnFalse": {
			givenDay: &DailySummary{
				Date:     *date(t, "2021-02-04"),
				Absences: []AbsenceBlock{{Reason: TypeUnpaid}},
			},
			expectedHoliday: false,
		},
		"GivenDailyWithAbsence_WhenTypeLegalLeaves_ThenReturnTrue": {
			givenDay: &DailySummary{
				Date:     *date(t, "2021-02-04"),
				Absences: []AbsenceBlock{{Reason: TypeLegalLeavesPrefix}},
			},
			expectedHoliday: true,
		},
		"GivenDailyWithAbsence_WhenTypeMilitary_ThenReturnTrue": {
			givenDay: &DailySummary{
				Date:     *date(t, "2021-02-04"),
				Absences: []AbsenceBlock{{Reason: TypeMilitaryService}},
			},
			expectedHoliday: true,
		},
		"GivenDailyWithAbsence_WhenTypeSpecialOccasions_ThenReturnTrue": {
			givenDay: &DailySummary{
				Date:     *date(t, "2021-02-04"),
				Absences: []AbsenceBlock{{Reason: TypeSpecialOccasions}},
			},
			expectedHoliday: true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			result := tt.givenDay.IsHoliday()
			assert.Equal(t, tt.expectedHoliday, result)
		})
	}
}
