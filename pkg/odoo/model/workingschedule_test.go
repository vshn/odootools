package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkingSchedule_GetFTERatio(t *testing.T) {
	tests := map[string]struct {
		givenNamePattern string
		expectedRatio    float64
		expectErr        bool
	}{
		"GivenWorkloadPrefix_When100%_ThenExpect_1": {
			givenNamePattern: "Workload: 100% (40:00)",
			expectedRatio:    1,
		},
		"GivenWorkWeekSuffix_When80%_ThenExpect_0.8": {
			givenNamePattern: "80% Work Week",
			expectedRatio:    0.8,
		},
		"GivenStandardWorkload_ThenExpect_1": {
			givenNamePattern: "Standard 100% Work Week",
			expectedRatio:    1,
		},
		"GivenEmptyName_ThenExpectErr": {
			givenNamePattern: "",
			expectErr:        true,
		},
		"GivenNameWithoutPercentage_ThenExpectErr": {
			givenNamePattern: "Some Work Week",
			expectErr:        true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			subject := WorkingSchedule{
				Name: tt.givenNamePattern,
			}
			result, err := subject.GetFTERatio()
			if tt.expectErr {
				require.Error(t, err)
				return
			}
			assert.Equal(t, tt.expectedRatio, result)
		})
	}
}
