package odoo

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTimeZone_UnmarshalJSON(t *testing.T) {
	tests := map[string]struct {
		givenInput       string
		expectedLocation *time.Location
	}{
		"UTC":           {givenInput: "UTC", expectedLocation: time.UTC},
		"Local":         {givenInput: "Local", expectedLocation: time.Local},
		"EuropeZurich":  {givenInput: `"Europe/Zurich"`, expectedLocation: mustLoadLocation("Europe/Zurich")},
		"CanadaPacific": {givenInput: "Canada/Pacific", expectedLocation: mustLoadLocation("Canada/Pacific")},
		"PST8PDT":       {givenInput: "PST8PDT", expectedLocation: mustLoadLocation("PST8PDT")},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			subject := &TimeZone{}
			err := subject.UnmarshalJSON([]byte(tt.givenInput))
			require.NoError(t, err)
			if tt.expectedLocation == nil {
				assert.Nil(t, subject.Location())
				return
			}
			assert.Equal(t, tt.expectedLocation, subject.Location())
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
