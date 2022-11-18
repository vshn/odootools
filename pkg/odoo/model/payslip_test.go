package model

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vshn/odootools/pkg/odoo"
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
			p := Payslip{XOvertime: tt.givenOvertime}
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

func TestPayslipList_FilterInMonth(t *testing.T) {
	tests := map[string]struct {
		givenDate       time.Time
		givenList       PayslipList
		expectedPayslip *Payslip
	}{
		"EmptyList": {
			givenDate:       time.Date(2021, 02, 03, 0, 0, 0, 0, time.UTC),
			givenList:       PayslipList{},
			expectedPayslip: nil,
		},
		"DateNotCovered": {
			givenDate: time.Date(2021, 02, 03, 0, 0, 0, 0, time.UTC),
			givenList: PayslipList{
				Items: []Payslip{
					{DateFrom: odoo.NewDate(2021, 01, 01, 0, 0, 0, time.UTC), DateTo: odoo.NewDate(2021, 01, 31, 0, 0, 0, time.UTC)},
				},
			},
			expectedPayslip: nil,
		},
		"MultipleEntries_DateCovered": {
			givenDate: time.Date(2021, 02, 03, 0, 0, 0, 0, time.UTC),
			givenList: PayslipList{
				Items: []Payslip{
					{DateFrom: odoo.NewDate(2021, 01, 01, 0, 0, 0, time.UTC), DateTo: odoo.NewDate(2021, 01, 31, 0, 0, 0, time.UTC)},
					{DateFrom: odoo.NewDate(2021, 02, 01, 0, 0, 0, time.UTC), DateTo: odoo.NewDate(2021, 02, 28, 0, 0, 0, time.UTC)},
				},
			},
			expectedPayslip: &Payslip{
				DateFrom: odoo.NewDate(2021, 02, 01, 0, 0, 0, time.UTC), DateTo: odoo.NewDate(2021, 02, 28, 0, 0, 0, time.UTC),
			},
		},
		"DateCovered": {
			givenDate: time.Date(2021, 02, 03, 0, 0, 0, 0, time.UTC),
			givenList: PayslipList{
				Items: []Payslip{
					{DateFrom: odoo.NewDate(2021, 02, 01, 0, 0, 0, time.UTC), DateTo: odoo.NewDate(2021, 02, 28, 0, 0, 0, time.UTC)},
				},
			},
			expectedPayslip: &Payslip{
				DateFrom: odoo.NewDate(2021, 02, 01, 0, 0, 0, time.UTC), DateTo: odoo.NewDate(2021, 02, 28, 0, 0, 0, time.UTC),
			},
		},
		"DateCoveredInTimezone": {
			givenDate: time.Date(2021, 02, 01, 0, 0, 0, 0, zurichTZ),
			givenList: PayslipList{
				Items: []Payslip{
					{DateFrom: odoo.NewDate(2021, 02, 01, 0, 0, 0, time.UTC), DateTo: odoo.NewDate(2021, 02, 28, 0, 0, 0, time.UTC)},
				},
			},
			expectedPayslip: &Payslip{
				DateFrom: odoo.NewDate(2021, 02, 01, 0, 0, 0, time.UTC), DateTo: odoo.NewDate(2021, 02, 28, 0, 0, 0, time.UTC),
			},
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := tc.givenList.FilterInMonth(tc.givenDate)
			assert.Equal(t, tc.expectedPayslip, result)
		})
	}
}

func newDuration(t *testing.T, s string) time.Duration {
	d, err := time.ParseDuration(s)
	require.NoError(t, err)
	return d
}
