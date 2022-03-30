package reportconfig

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_getStartOfWeek(t *testing.T) {
	tests := []struct {
		givenDay    time.Time
		expectedDay time.Time
	}{
		{
			givenDay:    time.Date(2022, time.March, 28, 4, 5, 6, 0, time.UTC), // Monday
			expectedDay: time.Date(2022, time.March, 28, 0, 0, 0, 0, time.UTC), // Monday
		},
		{
			givenDay:    time.Date(2022, time.March, 29, 4, 5, 6, 0, time.UTC), // Tuesday
			expectedDay: time.Date(2022, time.March, 28, 0, 0, 0, 0, time.UTC), // Monday
		},
		{
			givenDay:    time.Date(2022, time.April, 2, 4, 5, 6, 0, time.UTC),  // Saturday
			expectedDay: time.Date(2022, time.March, 28, 0, 0, 0, 0, time.UTC), // Monday
		},
		{
			givenDay:    time.Date(2022, time.April, 3, 4, 5, 6, 0, time.UTC),  // Sunday
			expectedDay: time.Date(2022, time.March, 28, 0, 0, 0, 0, time.UTC), // Monday
		},
	}
	for _, tc := range tests {
		t.Run(tc.givenDay.Weekday().String(), func(t *testing.T) {
			result := getStartOfWeek(tc.givenDay)
			assert.Equal(t, tc.expectedDay, result)
		})
	}
}

func Test_getEndOfWeek(t *testing.T) {
	tests := []struct {
		givenDay    time.Time
		expectedDay time.Time
	}{
		{
			givenDay:    time.Date(2022, time.March, 28, 4, 5, 6, 0, time.UTC), // Monday
			expectedDay: time.Date(2022, time.April, 3, 0, 0, 0, 0, time.UTC),  // Sunday
		},
		{
			givenDay:    time.Date(2022, time.March, 29, 4, 5, 6, 0, time.UTC), // Tuesday
			expectedDay: time.Date(2022, time.April, 3, 0, 0, 0, 0, time.UTC),  // Sunday
		},
		{
			givenDay:    time.Date(2022, time.April, 2, 4, 5, 6, 0, time.UTC), // Saturday
			expectedDay: time.Date(2022, time.April, 3, 0, 0, 0, 0, time.UTC), // Sunday
		},
		{
			givenDay:    time.Date(2022, time.April, 3, 4, 5, 6, 0, time.UTC), // Sunday
			expectedDay: time.Date(2022, time.April, 3, 0, 0, 0, 0, time.UTC), // Sunday
		},
	}
	for _, tc := range tests {
		t.Run(tc.givenDay.Weekday().String(), func(t *testing.T) {
			result := getEndOfWeek(tc.givenDay)
			assert.Equal(t, tc.expectedDay, result)
		})
	}
}
