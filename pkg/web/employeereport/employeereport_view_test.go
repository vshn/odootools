package employeereport

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vshn/odootools/pkg/odoo/model"
	"github.com/vshn/odootools/pkg/timesheet"
)

func TestReportView_getOvertimeBalancePreview(t *testing.T) {
	tests := map[string]struct {
		givenReport                        timesheet.MonthlyReport
		givenPreviousPayslip               *model.Payslip
		givenNextPayslip                   *model.Payslip
		expectedOvertimeBalance            string
		expectedOvertimeBalanceEditPreview string
		expectedButtonText                 string
	}{
		"GivenNewEmployee_WhenNoNextPayslipFound_ThenExpectError": {
			givenReport:                        dummyReport(t),
			givenPreviousPayslip:               nil,
			givenNextPayslip:                   nil,
			expectedOvertimeBalance:            "1:02",
			expectedOvertimeBalanceEditPreview: "1:02",
			expectedButtonText:                 "Save (New)",
		},
		"GivenNewEmployee_WhenNextPayslipExists_ThenUsePreviewValue": {
			givenReport:                        dummyReport(t),
			givenPreviousPayslip:               nil,
			givenNextPayslip:                   &model.Payslip{},
			expectedOvertimeBalance:            "1:02:03",
			expectedOvertimeBalanceEditPreview: "1:02:03",
			expectedButtonText:                 "Save (New)",
		},
		"GivenNewEmployee_WhenNextPayslipExistsWithExistingValue_ThenUseExistingValue": {
			givenReport:                        dummyReport(t),
			givenPreviousPayslip:               nil,
			givenNextPayslip:                   &model.Payslip{Overtime: "2:00:00"},
			expectedOvertimeBalance:            "1:02",
			expectedOvertimeBalanceEditPreview: "2:00:00",
			expectedButtonText:                 "Save (Update)",
		},
		"GivenExistingEmployeeWithoutOvertime_WhenNextPayslipExistsWithExistingValue_ThenUseExistingValue": {
			givenReport:                        dummyReport(t),
			givenPreviousPayslip:               &model.Payslip{},
			givenNextPayslip:                   &model.Payslip{Overtime: "2:00:00"},
			expectedOvertimeBalance:            "1:02:03",
			expectedOvertimeBalanceEditPreview: "2:00:00",
			expectedButtonText:                 "Save (Update)",
		},
		"GivenExistingEmployeeWithOvertime_WhenNextPayslipExistsWithExistingValue_ThenUseExistingValue": {
			givenReport:                        dummyReport(t),
			givenPreviousPayslip:               &model.Payslip{Overtime: "1:00:00"},
			givenNextPayslip:                   &model.Payslip{Overtime: "-5:00:00"},
			expectedOvertimeBalance:            "2:02:03",
			expectedOvertimeBalanceEditPreview: "-5:00:00",
			expectedButtonText:                 "Save (Update)",
		},
		"GivenExistingEmployeeWithOvertime_WhenNextPayslipExistsWithNoValue_ThenUsePredictedValue": {
			givenReport:                        dummyReport(t),
			givenPreviousPayslip:               &model.Payslip{Overtime: "1:00:00"},
			givenNextPayslip:                   &model.Payslip{},
			expectedOvertimeBalance:            "2:02:03",
			expectedOvertimeBalanceEditPreview: "2:02:03",
			expectedButtonText:                 "Save (New)",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			v := reportView{}
			actualOvertimeBalance, actualOvertimeBalanceEditPreview := v.getOvertimeBalancePreview(tc.givenReport, tc.givenPreviousPayslip, tc.givenNextPayslip)
			assert.Equal(t, tc.expectedOvertimeBalance, actualOvertimeBalance, "overtime balance")
			assert.Equal(t, tc.expectedOvertimeBalanceEditPreview, actualOvertimeBalanceEditPreview, "overtime balance edit preview")
		})
	}
}

func dummyReport(t *testing.T) timesheet.MonthlyReport {
	duration, err := time.ParseDuration("1h2m3s")
	require.NoError(t, err)
	return timesheet.MonthlyReport{
		Summary: timesheet.Summary{
			TotalOvertime: duration,
		},
	}
}
