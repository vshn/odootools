package model

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vshn/odootools/pkg/odoo"
)

func TestLeave_SplitByDay(t *testing.T) {
	t.Run("GivenLeaveWithSingleDate_ThenExpectSameLeave", func(t *testing.T) {
		givenLeave := Leave{
			ID:       1,
			DateFrom: odoo.MustParseDateTime("2021-02-03 07:00:00"),
			DateTo:   odoo.MustParseDateTime("2021-02-03 19:00:00"),
			Type:     &LeaveType{ID: 1, Name: "SomeType"},
			State:    "validated",
		}
		expectedLeave := Leave{
			ID:       1,
			DateFrom: odoo.NewDate(2021, 02, 03, 0, 0, 0, time.UTC),
			DateTo:   odoo.NewDate(2021, 02, 03, 23, 59, 59, time.UTC),
			Type:     &LeaveType{ID: 1, Name: "SomeType"},
			State:    "validated",
		}
		result := givenLeave.SplitByDay()
		require.Len(t, result, 1)
		assert.Equal(t, expectedLeave, result[0])
	})

	tests := map[string]struct {
		givenLeave     Leave
		expectedLeaves []Leave
	}{
		"GivenLeave_WhenDurationGoesIntoNextDay_ThenExpectSplit": {
			givenLeave: Leave{
				DateFrom: odoo.MustParseDateTime("2021-02-03 07:00:00"), DateTo: odoo.MustParseDateTime("2021-02-04 19:00:00"),
			},
			expectedLeaves: []Leave{
				{DateFrom: odoo.MustParseDateTime("2021-02-03 00:00:00"), DateTo: odoo.MustParseDateTime("2021-02-03 23:59:59")},
				{DateFrom: odoo.MustParseDateTime("2021-02-04 00:00:00"), DateTo: odoo.MustParseDateTime("2021-02-04 23:59:59")},
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
				assert.Equal(t, expected.DateFrom, actual.DateFrom)
				assert.Equal(t, expected.DateTo, actual.DateTo)
				assert.Equal(t, expected.State, actual.State)
				assert.Zero(t, actual.ID)
			}
		})
	}
}
