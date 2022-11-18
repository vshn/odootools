package odoo

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDate_UnmarshalJSON(t *testing.T) {
	tests := map[string]struct {
		givenInput   string
		expectedDate Date
	}{
		"GivenFalse_ThenExpectZeroDate": {
			givenInput:   "false",
			expectedDate: Date{},
		},
		"GivenValidInput_WhenFormatIsDate_ThenExpectDate": {
			givenInput:   "2021-02-03",
			expectedDate: MustParseDate("2021-02-03"),
		},
		"GivenValidInput_WhenFormatIsDateTime_ThenExpectDateTime": {
			givenInput:   "2021-02-03 15:34:00",
			expectedDate: MustParseDateTime("2021-02-03 15:34:00"),
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			subject := Date{}
			err := subject.UnmarshalJSON([]byte(tt.givenInput))
			require.NoError(t, err)
			if tt.expectedDate.IsZero() {
				assert.True(t, subject.IsZero())
				return
			}
			assert.Equal(t, tt.expectedDate, subject)
		})
	}
}

func TestDate_MarshalJSON(t *testing.T) {
	tests := map[string]struct {
		givenDate      Date
		expectedOutput string
		expectedError  string
	}{
		"GivenZero_ThenReturnFalse": {
			givenDate:      Date{},
			expectedOutput: "false",
		},
		"GivenTime_ThenReturnFormatted": {
			givenDate:      Date{Time: time.Date(2021, 02, 03, 4, 5, 6, 0, time.UTC)},
			expectedOutput: "2021-02-03 04:05:06",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result, err := tc.givenDate.MarshalJSON()
			if tc.expectedError != "" {
				assert.EqualError(t, err, tc.expectedError)
			} else {
				assert.Equal(t, tc.expectedOutput, string(result))
			}
		})
	}
}
