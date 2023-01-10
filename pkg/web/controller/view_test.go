package controller

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReportView_formatDurationHumanFriendly(t *testing.T) {
	tests := map[string]struct {
		givenDuration   time.Duration
		expectedOutcome string
	}{
		"GivenNoDuration_ThenReturnZero": {
			givenDuration:   time.Duration(0),
			expectedOutcome: "0:00:00",
		},
		"GivenPositiveDuration_WhenDurationMoreThan0.5s_ThenRoundUp": {
			givenDuration:   parseDuration(t, "2h34m56s500ms"),
			expectedOutcome: "2:34:57",
		},
		"GivenPositiveDuration_WhenDurationMoreThan0.5s_ThenRoundUpEdgeCase": {
			givenDuration:   parseDuration(t, "1s999ms"),
			expectedOutcome: "0:00:02",
		},
		"GivenPositiveDuration_WhenDurationLessThan0.5s_ThenRoundDown": {
			givenDuration:   parseDuration(t, "499ms"),
			expectedOutcome: "0:00:00",
		},
		"GivenPositiveDuration_WhenUnder1Hour_ThenReturnMinutesOnly": {
			givenDuration:   parseDuration(t, "38m"),
			expectedOutcome: "0:38:00",
		},
		"GivenPositiveDuration_WhenOver1Hour_ThenReturnHoursAndMinutes": {
			givenDuration:   parseDuration(t, "1h38m"),
			expectedOutcome: "1:38:00",
		},
		"GivenPositiveDuration_WhenOver10Hour_ThenReturnHoursAndMinutes": {
			givenDuration:   parseDuration(t, "10h38m"),
			expectedOutcome: "10:38:00",
		},
		"GivenPositiveDuration_WhenOver100Hour_ThenReturnHoursAndMinutes": {
			givenDuration:   parseDuration(t, "100h38m"),
			expectedOutcome: "100:38:00",
		},
		"GivenNegativeDuration_WhenDurationMoreThan0.5s_ThenRoundUp": {
			givenDuration:   parseDuration(t, "-500ms"),
			expectedOutcome: "-0:00:01",
		},
		"GivenNegativeDuration_WhenDurationLessThan0.5s_ThenRoundDown": {
			givenDuration:   parseDuration(t, "-499ms"),
			expectedOutcome: "-0:00:00",
		},
		"GivenNegativeDuration_WhenUnder1Hour_ThenReturnMinutesOnly": {
			givenDuration:   parseDuration(t, "-38m"),
			expectedOutcome: "-0:38:00",
		},
		"GivenNegativeDuration_WhenOver1Hour_ThenReturnHoursAndMinutes": {
			givenDuration:   parseDuration(t, "-1h38m"),
			expectedOutcome: "-1:38:00",
		},
		"GivenNegativeDuration_WhenOver10Hour_ThenReturnHoursAndMinutes": {
			givenDuration:   parseDuration(t, "-10h38m"),
			expectedOutcome: "-10:38:00",
		},
		"GivenNegativeDuration_WhenOver100Hour_ThenReturnHoursAndMinutes": {
			givenDuration:   parseDuration(t, "-100h38m"),
			expectedOutcome: "-100:38:00",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			result := BaseView{}.FormatDurationInHours(tt.givenDuration)
			assert.Equal(t, tt.expectedOutcome, result)
		})
	}
}

func parseDuration(t *testing.T, format string) time.Duration {
	d, err := time.ParseDuration(format)
	require.NoError(t, err)
	return d
}

func TestBaseView_GetNextMonth(t *testing.T) {
	tests := map[string]struct {
		givenYear     int
		givenMonth    int
		expectedYear  int
		expectedMonth int
	}{
		"WithinYear": {
			givenYear: 2022, givenMonth: 1,
			expectedYear: 2022, expectedMonth: 2,
		},
		"OverlappingYear": {
			givenYear: 2022, givenMonth: 12,
			expectedYear: 2023, expectedMonth: 1,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			resultYear, resultMonth := BaseView{}.GetNextMonth(tc.givenYear, tc.givenMonth)
			assert.Equal(t, tc.expectedYear, resultYear)
			assert.Equal(t, tc.expectedMonth, resultMonth)
		})
	}
}

func TestBaseView_GetPreviousMonth(t *testing.T) {
	tests := map[string]struct {
		givenYear     int
		givenMonth    int
		expectedYear  int
		expectedMonth int
	}{
		"WithinYear": {
			givenYear: 2022, givenMonth: 2,
			expectedYear: 2022, expectedMonth: 1,
		},
		"OverlappingYear": {
			givenYear: 2022, givenMonth: 1,
			expectedYear: 2021, expectedMonth: 12,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			resultYear, resultMonth := BaseView{}.GetPreviousMonth(tc.givenYear, tc.givenMonth)
			assert.Equal(t, tc.expectedYear, resultYear)
			assert.Equal(t, tc.expectedMonth, resultMonth)
		})
	}
}

func TestBaseView_OvertimeClassnameThreshold(t *testing.T) {
	overtimeClassName := "Overtime"
	undertimeClassName := "Undertime"
	tests := map[string]struct {
		givenDuration   time.Duration
		givenDailyMax   time.Duration
		expectedOutcome string
	}{
		"MyFirstTestCase": {
			givenDuration:   parseDuration(t, "0h"),
			givenDailyMax:   parseDuration(t, "8h"),
			expectedOutcome: "",
		},
		"3PercentUndertime": {
			givenDuration:   parseDuration(t, "-15m"),
			givenDailyMax:   parseDuration(t, "8h"),
			expectedOutcome: undertimeClassName,
		},
		"3PercentOvertime": {
			givenDuration:   parseDuration(t, "15m"),
			givenDailyMax:   parseDuration(t, "8h"),
			expectedOutcome: overtimeClassName,
		},
		"14MinOvertime": {
			givenDuration:   parseDuration(t, "14m"),
			givenDailyMax:   parseDuration(t, "8h"),
			expectedOutcome: "",
		},
		"14MinUndertime": {
			givenDuration:   parseDuration(t, "-14m"),
			givenDailyMax:   parseDuration(t, "8h"),
			expectedOutcome: "",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			actual := BaseView{}.OvertimeClassnameThreshold(tc.givenDuration, tc.givenDailyMax)
			assert.Equal(t, tc.expectedOutcome, actual)
		})
	}
}
