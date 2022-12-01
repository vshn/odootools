package timesheet

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/vshn/odootools/pkg/odoo"
	"github.com/vshn/odootools/pkg/odoo/model"
)

func TestDailySummary_CalculateOvertime(t *testing.T) {
	weekday := "2021-02-03"
	weekendDay := "2021-02-06"
	tests := map[string]struct {
		givenShifts         []AttendanceShift
		givenDate           time.Time
		expectedOvertime    time.Duration
		expectedExcusedTime time.Duration
	}{
		"GivenSingleShift_WhenMoreThanDailyMax_ThenReturnOvertime": {
			givenShifts: []AttendanceShift{
				newAttendanceShift(hours(t, weekday, "09:00"), hours(t, weekday, "18:00"), ""),
			},
			expectedOvertime: hoursDuration(t, 1),
		},
		"GivenMultipleShifts_WhenMoreThanDailyMax_ThenReturnOvertime": {
			givenShifts: []AttendanceShift{
				newAttendanceShift(hours(t, weekday, "09:00"), hours(t, weekday, "12:00"), ""),
				newAttendanceShift(hours(t, weekday, "13:00"), hours(t, weekday, "19:00"), ""),
			},
			expectedOvertime: hoursDuration(t, 1),
		},
		"GivenMultipleShifts_WhenLessThanDailyMax_ThenReturnUndertime": {
			givenShifts: []AttendanceShift{
				newAttendanceShift(hours(t, weekday, "09:00"), hours(t, weekday, "12:00"), ""),
				newAttendanceShift(hours(t, weekday, "13:00"), hours(t, weekday, "17:00"), ""),
			},
			expectedOvertime: hoursDuration(t, -1),
		},
		"GivenSickLeaveShifts_WhenSickLeaveIsFilling_ThenReturnZero": {
			givenShifts: []AttendanceShift{
				newAttendanceShift(hours(t, weekday, "09:00"), hours(t, weekday, "12:00"), ""),
				newAttendanceShift(hours(t, weekday, "13:00"), hours(t, weekday, "18:00"), ReasonSickLeave),
			},
			expectedOvertime:    hoursDuration(t, 0),
			expectedExcusedTime: hoursDuration(t, 5),
		},
		"GivenSickLeaveShifts_WhenSickLeaveIsLessThanDailyMax_ThenReturnUndertime": {
			givenShifts: []AttendanceShift{
				newAttendanceShift(hours(t, weekday, "09:00"), hours(t, weekday, "12:00"), ""),
				newAttendanceShift(hours(t, weekday, "13:00"), hours(t, weekday, "17:00"), ReasonSickLeave),
			},
			expectedOvertime:    hoursDuration(t, -1),
			expectedExcusedTime: hoursDuration(t, 4),
		},
		"GivenSickLeaveShifts_WhenCombinedHoursExceedDailyMax_ThenCapOvertime": {
			givenShifts: []AttendanceShift{
				newAttendanceShift(hours(t, weekday, "09:00"), hours(t, weekday, "12:00"), ""),
				newAttendanceShift(hours(t, weekday, "13:00"), hours(t, weekday, "18:30"), ReasonSickLeave),
			},
			expectedOvertime:    hoursDuration(t, 0),
			expectedExcusedTime: hoursDuration(t, 5.5),
		},
		"GivenSickLeaveShifts_WhenExcusedHoursExceedDailyMax_ThenCapExcusedTime": {
			givenShifts: []AttendanceShift{
				{Start: hours(t, weekday, "09:00"), End: hours(t, weekday, "19:00"), Reason: ReasonSickLeave},
			},
			expectedOvertime:    hoursDuration(t, 0),
			expectedExcusedTime: hoursDuration(t, 8),
		},
		"GivenSickLeaveShifts_WhenWorkingHoursIsMoreThanDailyMax_ThenIgnoreSickLeaveCompletely": {
			givenShifts: []AttendanceShift{
				newAttendanceShift(hours(t, weekday, "09:00"), hours(t, weekday, "18:00"), ""),
				newAttendanceShift(hours(t, weekday, "19:00"), hours(t, weekday, "20:00"), ReasonSickLeave),
			},
			expectedOvertime:    hoursDuration(t, 1),
			expectedExcusedTime: hoursDuration(t, 1),
		},
		"GivenSickLeaveAndOutsideOfficeHoursShifts_WhenWorkingHoursIsMoreThanDailyMax_ThenIgnoreSickLeaveCompletely": {
			givenShifts: []AttendanceShift{
				newAttendanceShift(hours(t, weekday, "09:00"), hours(t, weekday, "18:00"), ""),                       // 1h overtime
				newAttendanceShift(hours(t, weekday, "19:00"), hours(t, weekday, "20:00"), ReasonSickLeave),          // no overtime
				newAttendanceShift(hours(t, weekday, "20:00"), hours(t, weekday, "22:00"), ReasonOutsideOfficeHours), // 3h overtime
			},
			expectedOvertime:    hoursDuration(t, 4),
			expectedExcusedTime: hoursDuration(t, 1),
		},
		"GivenNoShifts_WhenNoLeavesEither_ThenReturnOneDayUndertime": {
			givenShifts:      []AttendanceShift{},
			expectedOvertime: hoursDuration(t, -8),
		},
		"GivenDateInWeekend_WhenNoWorkingHours_ThenReturnNoOvertime": {
			givenShifts:      []AttendanceShift{},
			givenDate:        odoo.MustParseDate(weekendDay).Time,
			expectedOvertime: hoursDuration(t, 0),
		},
		"GivenDateInWeekend_WhenWorkingHoursLogged_ThenReturnOvertime": {
			givenShifts: []AttendanceShift{
				newAttendanceShift(hours(t, weekday, "09:00"), hours(t, weekday, "10:00"), ""), // 1h overtime
			},
			givenDate:        odoo.MustParseDate(weekendDay).Time,
			expectedOvertime: hoursDuration(t, 1),
		},
		"GivenDateInWeekend_WhenExcusedHoursLogged_ThenIgnoreExcusedHours": {
			givenShifts: []AttendanceShift{
				newAttendanceShift(hours(t, weekendDay, "09:00"), hours(t, weekendDay, "10:00"), ""),                  // 1h overtime
				newAttendanceShift(hours(t, weekendDay, "09:00"), hours(t, weekendDay, "10:00"), ReasonPublicService), // no overtime
			},
			givenDate:        odoo.MustParseDate(weekendDay).Time,
			expectedOvertime: hoursDuration(t, 1),
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			d := tt.givenDate
			if tt.givenDate.IsZero() {
				d = odoo.MustParseDate(weekday).Time
			}
			s := DailySummary{
				Date:     d,
				Shifts:   tt.givenShifts,
				FTERatio: 1,
			}
			result := s.CalculateOvertimeSummary()
			assert.Equal(t, tt.expectedOvertime, result.Overtime(), "overtime")
			assert.Equal(t, tt.expectedExcusedTime, result.ExcusedTime(), "excused time")
		})
	}
}

func TestDailySummary_CalculateDailyMaxHours(t *testing.T) {
	tests := map[string]struct {
		givenDate     time.Time
		givenFteRatio float64
		givenAbsences []AbsenceBlock
		expectedHours float64
	}{
		"GivenWeekDay_WhenIn2021_ThenReturn8Hours": {
			givenDate:     odoo.MustParseDate("2021-02-03").Time,
			givenFteRatio: float64(1),
			expectedHours: 8,
		},
		"GivenWeekDay_WhenIn2020_ThenReturn8.5Hours": {
			givenDate:     odoo.MustParseDate("2020-02-03").Time,
			givenFteRatio: float64(1),
			expectedHours: 8.5,
		},
		"GivenWeekendDay_ThenReturn0Hours": {
			givenDate:     odoo.MustParseDate("2021-02-06").Time,
			givenFteRatio: float64(1),
			expectedHours: 0,
		},
		"GivenAbsences_WhenTypeIsHoliday_ThenReturn0Hours": {
			givenDate:     odoo.MustParseDate("2021-02-06").Time,
			givenFteRatio: float64(1),
			givenAbsences: []AbsenceBlock{
				{Reason: TypePublicHoliday},
			},
			expectedHours: 0,
		},
		"GivenAbsences_WhenTypeIsUnpaid_ThenReturnNormalHours": {
			givenDate:     odoo.MustParseDate("2021-02-03").Time,
			givenFteRatio: float64(1),
			givenAbsences: []AbsenceBlock{
				{Reason: TypeUnpaid},
			},
			expectedHours: 8,
		},
		"GivenAbsencesWithSpecialFte_WhenTypeIsUnpaid_ThenReturnFteAdjustedHours": {
			givenDate:     odoo.MustParseDate("2021-02-03").Time,
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
				Date:     tt.givenDate,
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
				NewDailySummary(1, odoo.MustParseDate("2021-02-03").In(zurichTZ)),
			},
			givenDate:       odoo.MustParseDate("2021-02-03").In(zurichTZ),
			expectedSummary: NewDailySummary(1, odoo.MustParseDate("2021-02-03").In(zurichTZ)),
		},
		"GivenDailies_WhenDateMatchesInUTC_ThenReturnDaily": {
			givenDailies: []*DailySummary{
				NewDailySummary(1, odoo.MustParseDate("2021-02-04").UTC()),
				NewDailySummary(1, odoo.MustParseDate("2021-02-03").UTC()),
			},
			givenDate:       odoo.MustParseDateTime("2021-02-03 23:30:00").Time,
			expectedSummary: NewDailySummary(1, odoo.MustParseDate("2021-02-03").UTC()),
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
			givenDay:        &DailySummary{Date: odoo.MustParseDate("2021-02-04").Time},
			expectedHoliday: false,
		},
		"GivenDailyWithAbsence_WhenPublicHoliday_ThenReturnFalse": {
			givenDay: &DailySummary{
				Date:     odoo.MustParseDate("2021-02-04").Time,
				Absences: []AbsenceBlock{{Reason: TypePublicHoliday}},
			},
			expectedHoliday: false,
		},
		"GivenDailyWithAbsence_WhenPublicHolidayOnWeekend_ThenReturnFalse": {
			givenDay: &DailySummary{
				Date:     odoo.MustParseDate("2021-02-06").Time,
				Absences: []AbsenceBlock{{Reason: TypePublicHoliday}},
			},
			expectedHoliday: false,
		},
		"GivenDailyWithAbsence_WhenUnpaid_ThenReturnFalse": {
			givenDay: &DailySummary{
				Date:     odoo.MustParseDate("2021-02-04").Time,
				Absences: []AbsenceBlock{{Reason: TypeUnpaid}},
			},
			expectedHoliday: false,
		},
		"GivenDailyWithAbsence_WhenTypeLegalLeaves_ThenReturnTrue": {
			givenDay: &DailySummary{
				Date:     odoo.MustParseDate("2021-02-04").Time,
				Absences: []AbsenceBlock{{Reason: TypeLegalLeavesPrefix}},
			},
			expectedHoliday: true,
		},
		"GivenDailyWithAbsence_WhenTypeMilitary_ThenReturnTrue": {
			givenDay: &DailySummary{
				Date:     odoo.MustParseDate("2021-02-04").Time,
				Absences: []AbsenceBlock{{Reason: TypeMilitaryService}},
			},
			expectedHoliday: true,
		},
		"GivenDailyWithAbsence_WhenTypeSpecialOccasions_ThenReturnTrue": {
			givenDay: &DailySummary{
				Date:     odoo.MustParseDate("2021-02-04").Time,
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

func TestDailySummary_ValidateTimesheetEntries(t *testing.T) {
	tests := map[string]struct {
		givenShifts   []AttendanceShift
		expectedError string
	}{
		"EmptyList": {givenShifts: []AttendanceShift{}, expectedError: ""},
		"NilList":   {givenShifts: nil, expectedError: ""},
		"Single_SignIn_Error": {
			givenShifts: []AttendanceShift{
				{Start: model.Attendance{DateTime: odoo.NewDate(2021, 01, 02, 8, 0, 0, time.UTC), Action: model.ActionSignIn}},
			},
			expectedError: "no sign_out detected for 2021-01-02 after 2021-01-02 08:00:00 +0000 UTC",
		},
		"Single_SignOut_Error": {
			givenShifts: []AttendanceShift{
				{End: model.Attendance{DateTime: odoo.NewDate(2021, 01, 02, 8, 0, 0, time.UTC), Action: model.ActionSignOut}},
			},
			expectedError: "no sign_in detected for 2021-01-02 before 08:00:00",
		},
		"SameStartAndEnd_Error": {
			givenShifts: []AttendanceShift{
				{
					Start: model.Attendance{DateTime: odoo.NewDate(2021, 01, 02, 8, 0, 0, time.UTC), Action: model.ActionSignIn},
					End:   model.Attendance{DateTime: odoo.NewDate(2021, 01, 02, 8, 0, 0, time.UTC), Action: model.ActionSignOut},
				},
			},
			expectedError: "shift start and end times cannot be the same for 2021-01-02: 2021-01-02 08:00:00 +0000 UTC",
		},
		"Multiple_SignOutMissing_Error": {
			givenShifts: []AttendanceShift{
				{
					Start: model.Attendance{DateTime: odoo.NewDate(2021, 01, 02, 8, 0, 0, time.UTC), Action: model.ActionSignIn},
					End:   model.Attendance{DateTime: odoo.NewDate(2021, 01, 02, 10, 0, 0, time.UTC), Action: model.ActionSignOut},
				},
				{
					Start: model.Attendance{DateTime: odoo.NewDate(2021, 01, 02, 11, 0, 0, time.UTC), Action: model.ActionSignIn},
				},
			},
			expectedError: "no sign_out detected for 2021-01-02 after 2021-01-02 11:00:00 +0000 UTC",
		},
		"Multiple_TotalDurationExceeds24h": {
			givenShifts: []AttendanceShift{
				{
					Start: model.Attendance{DateTime: odoo.NewDate(2021, 01, 02, 0, 0, 0, zurichTZ), Action: model.ActionSignIn},
					End:   model.Attendance{DateTime: odoo.NewDate(2021, 01, 02, 12, 0, 0, zurichTZ), Action: model.ActionSignOut},
				},
				{
					Start: model.Attendance{DateTime: odoo.NewDate(2021, 01, 02, 12, 0, 0, zurichTZ), Action: model.ActionSignIn},
					End:   model.Attendance{DateTime: odoo.NewDate(2021, 01, 03, 0, 0, 1, zurichTZ), Action: model.ActionSignOut},
				},
			},
			expectedError: "duration of all shifts for 2021-01-02 cannot exceed 24h: 24h0m1s",
		},
		"DifferentReasonsInShift": {
			givenShifts: []AttendanceShift{
				{
					Start: model.Attendance{DateTime: odoo.NewDate(2021, 01, 02, 8, 0, 0, time.UTC), Action: model.ActionSignIn},
					End:   model.Attendance{DateTime: odoo.NewDate(2021, 01, 02, 10, 0, 0, time.UTC), Action: model.ActionSignOut, Reason: &model.ActionReason{Name: ReasonSickLeave}},
				},
			},
			expectedError: "the reasons for shift sign_in and sign_out have to be equal: start 2021-01-02 08:00:00 +0000 UTC (), end 2021-01-02 10:00:00 +0000 UTC (Sick / Medical Consultation)",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			s := &DailySummary{Shifts: tc.givenShifts, Date: time.Date(2021, 01, 02, 0, 0, 0, 0, time.UTC)}
			err := s.ValidateTimesheetEntries()
			if tc.expectedError != "" {
				assert.EqualError(t, err, tc.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func newAttendanceShift(start, end odoo.Date, reason string) AttendanceShift {
	return AttendanceShift{
		Start: model.Attendance{DateTime: start, Action: model.ActionSignIn, Reason: &model.ActionReason{Name: reason}},
		End:   model.Attendance{DateTime: end, Action: model.ActionSignOut, Reason: &model.ActionReason{Name: reason}},
	}
}
