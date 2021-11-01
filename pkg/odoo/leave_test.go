package odoo

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLeave_SplitByDay(t *testing.T) {
	t.Run("GivenLeaveWithSingleDate_ThenExpectSameLeave", func(t *testing.T) {
		givenLeave := Leave{
			ID:       1,
			DateFrom: newDate(t, "2021-02-03 07:00"),
			DateTo:   newDate(t, "2021-02-03 19:00"),
			Type:     &LeaveType{ID: 1, Name: "SomeType"},
			State:    "validated",
		}
		result := givenLeave.SplitByDay()
		require.Len(t, result, 1)
		assert.Equal(t, givenLeave, result[0])
	})

	tests := map[string]struct {
		givenLeave     Leave
		expectedLeaves []Leave
	}{
		"GivenLeave_WhenDurationGoesIntoNextDay_ThenExpectSplit": {
			givenLeave: Leave{
				DateFrom: newDate(t, "2021-02-03 07:00"), DateTo: newDate(t, "2021-02-04 19:00"),
			},
			expectedLeaves: []Leave{
				{DateFrom: newDate(t, "2021-02-03 07:00"), DateTo: newDate(t, "2021-02-03 15:00")},
				{DateFrom: newDate(t, "2021-02-04 07:00"), DateTo: newDate(t, "2021-02-04 15:00")},
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			result := tt.givenLeave.SplitByDay()
			require.Len(t, result, len(tt.expectedLeaves))
			for i := 0; i < len(result); i++ {
				actual := result[i]
				expected := tt.expectedLeaves[i]
				assert.Equal(t, expected.DateFrom.ToTime(), actual.DateFrom.ToTime())
				assert.Equal(t, expected.DateTo.ToTime(), actual.DateTo.ToTime())
				assert.Equal(t, expected.State, actual.State)
				assert.Zero(t, actual.ID)
			}
		})
	}
}
