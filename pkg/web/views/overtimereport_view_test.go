package views

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOvertimeReportView_formatDurationHumanFriendly(t *testing.T) {
	tests := map[string]struct {
		givenDuration   time.Duration
		expectedOutcome string
	}{
		"GivenNoDuration_ThenReturnZero": {
			givenDuration:   time.Duration(0),
			expectedOutcome: "0:00",
		},
		"GivenPositiveDuration_WhenDurationMoreThan30s_ThenRoundUp": {
			givenDuration:   parseDuration(t, "30s"),
			expectedOutcome: "0:01",
		},
		"GivenPositiveDuration_WhenDurationMoreThan30s_ThenRoundUpEdgeCase": {
			givenDuration:   parseDuration(t, "1m59s"),
			expectedOutcome: "0:02",
		},
		"GivenPositiveDuration_WhenDurationLessThan30s_ThenRoundDown": {
			givenDuration:   parseDuration(t, "29s"),
			expectedOutcome: "0:00",
		},
		"GivenPositiveDuration_WhenUnder1Hour_ThenReturnMinutesOnly": {
			givenDuration:   parseDuration(t, "38m"),
			expectedOutcome: "0:38",
		},
		"GivenPositiveDuration_WhenOver1Hour_ThenReturnHoursAndMinutes": {
			givenDuration:   parseDuration(t, "1h38m"),
			expectedOutcome: "1:38",
		},
		"GivenPositiveDuration_WhenOver10Hour_ThenReturnHoursAndMinutes": {
			givenDuration:   parseDuration(t, "10h38m"),
			expectedOutcome: "10:38",
		},
		"GivenPositiveDuration_WhenOver100Hour_ThenReturnHoursAndMinutes": {
			givenDuration:   parseDuration(t, "100h38m"),
			expectedOutcome: "100:38",
		},
		"GivenNegativeDuration_WhenDurationMoreThan30s_ThenRoundUp": {
			givenDuration:   parseDuration(t, "-30s"),
			expectedOutcome: "-0:01",
		},
		"GivenNegativeDuration_WhenDurationLessThan30s_ThenRoundDown": {
			givenDuration:   parseDuration(t, "-29s"),
			expectedOutcome: "-0:00",
		},
		"GivenNegativeDuration_WhenUnder1Hour_ThenReturnMinutesOnly": {
			givenDuration:   parseDuration(t, "-38m"),
			expectedOutcome: "-0:38",
		},
		"GivenNegativeDuration_WhenOver1Hour_ThenReturnHoursAndMinutes": {
			givenDuration:   parseDuration(t, "-1h38m"),
			expectedOutcome: "-1:38",
		},
		"GivenNegativeDuration_WhenOver10Hour_ThenReturnHoursAndMinutes": {
			givenDuration:   parseDuration(t, "-10h38m"),
			expectedOutcome: "-10:38",
		},
		"GivenNegativeDuration_WhenOver100Hour_ThenReturnHoursAndMinutes": {
			givenDuration:   parseDuration(t, "-100h38m"),
			expectedOutcome: "-100:38",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			result := formatDurationInHours(tt.givenDuration)
			assert.Equal(t, tt.expectedOutcome, result)
		})
	}
}

func parseDuration(t *testing.T, format string) time.Duration {
	d, err := time.ParseDuration(format)
	require.NoError(t, err)
	return d
}
