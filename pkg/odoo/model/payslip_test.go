package model

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPayslip_ParseOvertime(t *testing.T) {
	tests := map[string]struct {
		givenOvertime    string
		expectedOvertime time.Duration
		expectedError    string
	}{
		"GivenEmptyField_ThenExpectZero": {
			givenOvertime: "",
		},
		"GivenField_WhenNoFormatRecognized_ThenExpectError": {
			givenOvertime: "not-properly-formatted",
			expectedError: "format not parseable: not-properly-formatted",
		},
		"GivenField_WhenColonFormatWithoutSeconds_ThenExpectParsedHours": {
			givenOvertime:    "Currently 143:34 (including holidays)\n",
			expectedOvertime: newDuration(t, "143h34m"),
		},
		"GivenField_WhenColonFormatWithSeconds_ThenExpectParsedHours": {
			givenOvertime:    "Currently 143:34:43 (including holidays)\n",
			expectedOvertime: newDuration(t, "143h34m43s"),
		},
		"GivenField_WhenColonFormatNegativeNumber_ThenExpectNegativeHours": {
			givenOvertime:    "-143:34:43 (including holidays)\n",
			expectedOvertime: newDuration(t, "-143h34m43s"),
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			p := Payslip{Overtime: tt.givenOvertime}
			result, err := p.ParseOvertime()
			if tt.expectedError != "" {
				assert.EqualError(t, err, tt.expectedError)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedOvertime, result)
		})
	}
}

func newDuration(t *testing.T, s string) time.Duration {
	d, err := time.ParseDuration(s)
	require.NoError(t, err)
	return d
}
