package odoo

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	zurichTZ    *time.Location
	vancouverTZ *time.Location
)

func init() {
	zue, err := time.LoadLocation("Europe/Zurich")
	if err != nil {
		panic(err)
	}
	zurichTZ = zue
	van, err := time.LoadLocation("America/Vancouver")
	if err != nil {
		panic(err)
	}
	vancouverTZ = van
}

func TestTimeZone_UnmarshalJSON(t *testing.T) {
	tests := map[string]struct {
		givenInput       string
		expectedLocation *time.Location
	}{
		"false":            {givenInput: "false", expectedLocation: nil},
		"empty":            {givenInput: "", expectedLocation: nil},
		"UTC":              {givenInput: "UTC", expectedLocation: time.UTC},
		"Local":            {givenInput: "Local", expectedLocation: time.Local},
		"EuropeZurich":     {givenInput: `"Europe/Zurich"`, expectedLocation: mustLoadLocation("Europe/Zurich")},
		"AmericaVancouver": {givenInput: "America/Vancouver", expectedLocation: mustLoadLocation("America/Vancouver")},
		"PST8PDT":          {givenInput: "PST8PDT", expectedLocation: mustLoadLocation("PST8PDT")},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			subject := &TimeZone{}
			err := subject.UnmarshalJSON([]byte(tt.givenInput))
			require.NoError(t, err)
			if tt.expectedLocation == nil {
				assert.Nil(t, subject.Location)
				return
			}
			assert.Equal(t, tt.expectedLocation, subject.Location)
		})
	}
}

func TestTimeZone_MarshalJSON(t *testing.T) {
	tests := map[string]struct {
		givenLocation  *time.Location
		expectedOutput []byte
	}{
		"nil":          {givenLocation: nil, expectedOutput: []byte("null")},
		"EuropeZurich": {givenLocation: mustLoadLocation("Europe/Zurich"), expectedOutput: []byte(`"Europe/Zurich"`)},
		"Local":        {givenLocation: time.Local, expectedOutput: []byte("null")}, // "Local" isn't recognized by Odoo
		"UTC":          {givenLocation: time.UTC, expectedOutput: []byte(`"UTC"`)},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			subject := &TimeZone{Location: tt.givenLocation}
			result, err := subject.MarshalJSON()
			require.NoError(t, err)
			if tt.expectedOutput == nil {
				assert.Nil(t, result)
			} else {
				assert.Equal(t, tt.expectedOutput, result)
			}
		})
	}
}

func TestTimeZone_IsEqualTo(t *testing.T) {
	tests := map[string]struct {
		givenTimeZoneA *TimeZone
		givenTimeZoneB *TimeZone
		expectedResult bool
	}{
		"BothNil": {
			givenTimeZoneA: nil, givenTimeZoneB: nil,
			expectedResult: true,
		},
		"BothNilNested": {
			givenTimeZoneA: NewTimeZone(nil),
			givenTimeZoneB: NewTimeZone(nil),
			expectedResult: true,
		},
		"A_IsNil": {
			givenTimeZoneA: nil,
			givenTimeZoneB: NewTimeZone(vancouverTZ),
			expectedResult: false,
		},
		"B_IsNil": {
			givenTimeZoneA: NewTimeZone(vancouverTZ),
			givenTimeZoneB: nil,
			expectedResult: false,
		},
		"BothSame": {
			givenTimeZoneA: NewTimeZone(zurichTZ),
			givenTimeZoneB: NewTimeZone(zurichTZ),
			expectedResult: true,
		},
		"A_NestedNil": {
			givenTimeZoneA: NewTimeZone(nil),
			givenTimeZoneB: NewTimeZone(zurichTZ),
			expectedResult: false,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			actual := tc.givenTimeZoneA.IsEqualTo(tc.givenTimeZoneB)
			assert.Equal(t, tc.expectedResult, actual, "zone not equal: zone A: %s, zone B: %s", tc.givenTimeZoneA, tc.givenTimeZoneB)
		})
	}
}

func mustLoadLocation(name string) *time.Location {
	loc, err := time.LoadLocation(name)
	if err != nil {
		panic(err)
	}
	return loc
}
