package employeereport

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vshn/odootools/pkg/odoo/model"
)

func TestReportView_getButtonText(t *testing.T) {
	tests := map[string]struct {
		givenNextPayslip   *model.Payslip
		expectedButtonText string
	}{
		"GivenNoPayslip_ThenExpectSaveNew": {
			givenNextPayslip:   nil,
			expectedButtonText: "Save (New)",
		},
		"GivenPayslip_WhenNoOvertimeSaved_ThenExpectSaveNew": {
			givenNextPayslip:   &model.Payslip{},
			expectedButtonText: "Save (New)",
		},
		"GivenPayslip_WhenOvertimeSaved_ThenExpectSaveUpdate": {
			givenNextPayslip:   &model.Payslip{Overtime: "2:00:00"},
			expectedButtonText: "Save (Update)",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			v := reportView{}
			result := v.getButtonText(tc.givenNextPayslip)
			assert.Equal(t, tc.expectedButtonText, result, "button text")
		})
	}
}

func TestReportView_getPreviousBalance(t *testing.T) {
	tests := map[string]struct {
		givenPreviousPayslip *model.Payslip
		expectedCell         string
		expectedBalance      time.Duration
	}{
		"GivenNoPayslip_ThenExpectNoPayslipFound": {
			givenPreviousPayslip: nil,
			expectedCell:         "<no payslip found>",
			expectedBalance:      0,
		},
		"GivenPayslip_WhenNoOvertimeSaved_ThenExpectNoOvertimeSaved": {
			givenPreviousPayslip: &model.Payslip{},
			expectedCell:         "<no overtime saved>",
			expectedBalance:      0,
		},
		"GivenPayslip_WhenOvertimeSaved_ThenExpectParsedValue": {
			givenPreviousPayslip: &model.Payslip{Overtime: "2:00:00 incl holidays"},
			expectedCell:         "2:00:00 incl holidays",
			expectedBalance:      mustParseDuration(t, "2h"),
		},
		"GivenPayslip_WhenOvertimeCannotParse_ThenExpectErrorText": {
			givenPreviousPayslip: &model.Payslip{Overtime: "2 hours"},
			expectedCell:         "<format not parseable: 2 hours>",
			expectedBalance:      0,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			v := reportView{}
			cellText, balance := v.getPreviousBalance(tc.givenPreviousPayslip)
			assert.Equal(t, tc.expectedCell, cellText, "cell text")
			assert.Equal(t, tc.expectedBalance, balance, "previous balance")
		})
	}
}

func TestReportView_getNextBalance(t *testing.T) {
	tests := map[string]struct {
		givenNextPayslip     *model.Payslip
		givenProposedBalance time.Duration
		expectedCell         string
		expectedBalance      time.Duration
	}{
		"GivenNoPayslip_ThenExpectNoPayslipFound": {
			givenNextPayslip:     nil,
			givenProposedBalance: 0,
			expectedCell:         "<no payslip found>",
			expectedBalance:      0,
		},
		"GivenPayslip_WhenNoOvertimeSaved_ThenExpectEmptyCell": {
			givenNextPayslip:     &model.Payslip{},
			givenProposedBalance: mustParseDuration(t, "2h"),
			expectedCell:         "",
			expectedBalance:      mustParseDuration(t, "2h"),
		},
		"GivenPayslip_WhenOvertimeSaved_ThenExpectCellValueVerbatim": {
			givenNextPayslip:     &model.Payslip{Overtime: "2:00:00 with holidays"},
			givenProposedBalance: mustParseDuration(t, "4h"),
			expectedCell:         "2:00:00 with holidays",
			expectedBalance:      mustParseDuration(t, "2h"),
		},
		"GivenPayslip_WhenOvertimeCannotParse_ThenExpectErrorTextAndUnchangedProposedBalance": {
			givenNextPayslip:     &model.Payslip{Overtime: "2 hours"},
			givenProposedBalance: mustParseDuration(t, "4h"),
			expectedCell:         "<format not parseable: 2 hours>",
			expectedBalance:      mustParseDuration(t, "4h"),
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			v := reportView{}
			cellText, balance := v.getNextBalance(tc.givenProposedBalance, tc.givenNextPayslip)
			assert.Equal(t, tc.expectedCell, cellText, "cell text")
			assert.Equal(t, tc.expectedBalance, balance, "next balance")
		})
	}
}

func TestReportView_getOvertimeBalanceEditPreview(t *testing.T) {
	tests := map[string]struct {
		givenNextPayslip                   *model.Payslip
		givenProposedBalance               time.Duration
		expectedOvertimeBalanceEditPreview string
	}{
		"GivenNoPayslip_WhenProposedBalanceZero_ThenExpectZero": {
			givenNextPayslip:                   nil,
			givenProposedBalance:               0,
			expectedOvertimeBalanceEditPreview: "0:00:00",
		},
		"GivenNoPayslip_WhenProposedBalanceNotZero_ThenExpectProposedBalance": {
			givenNextPayslip:                   nil,
			givenProposedBalance:               mustParseDuration(t, "2h"),
			expectedOvertimeBalanceEditPreview: "2:00:00",
		},
		"GivenNoPayslip_WhenNoOvertimeSaved_ThenExpectProposedBalance": {
			givenNextPayslip:                   &model.Payslip{},
			givenProposedBalance:               mustParseDuration(t, "2h"),
			expectedOvertimeBalanceEditPreview: "2:00:00",
		},
		"GivenNoPayslip_WhenOvertimeExists_ThenUseExistingValue": {
			givenNextPayslip:                   &model.Payslip{Overtime: "2:00:00"},
			givenProposedBalance:               mustParseDuration(t, "4h"),
			expectedOvertimeBalanceEditPreview: "2:00:00",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			v := reportView{}
			result := v.getOvertimeBalanceEditPreview(tc.givenNextPayslip, tc.givenProposedBalance)
			assert.Equal(t, tc.expectedOvertimeBalanceEditPreview, result, "overtime balance edit preview")
		})
	}
}

func mustParseDuration(t *testing.T, fmt string) time.Duration {
	duration, err := time.ParseDuration(fmt)
	require.NoError(t, err)
	return duration
}
