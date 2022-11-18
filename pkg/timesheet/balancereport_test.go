package timesheet

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/vshn/odootools/pkg/odoo"
	"github.com/vshn/odootools/pkg/odoo/model"
)

func TestBalanceReportBuilder_CalculateBalanceReport(t *testing.T) {
	tests := map[string]struct {
		givenReport   Report
		givenPayslips model.PayslipList

		expectedCalculatedBalance time.Duration
		expectedDefinitiveBalance *time.Duration
	}{
		"NoPreviousAndCurrentPayslip": {
			// this case is for new starting employees that don't have a previous and current payslip.
			givenReport: Report{
				From:    time.Date(2021, time.February, 01, 0, 0, 0, 0, time.UTC),
				Summary: Summary{TotalOvertime: hoursDuration(t, 5.5)},
			},
			givenPayslips:             model.PayslipList{},
			expectedCalculatedBalance: hoursDuration(t, 5.5),
			expectedDefinitiveBalance: nil,
		},
		"NoPreviousPayslip": {
			// this case is for new starting employees that don't have a previous, but only their first (finished) payslip.
			givenReport: Report{
				From:    time.Date(2021, time.February, 01, 0, 0, 0, 0, time.UTC),
				Summary: Summary{TotalOvertime: hoursDuration(t, 5.5)},
			},
			givenPayslips: model.PayslipList{Items: []model.Payslip{
				{
					DateFrom:  odoo.NewDate(2021, time.February, 01, 0, 0, 0, time.UTC),
					DateTo:    odoo.NewDate(2021, time.February, 28, 0, 0, 0, time.UTC),
					XOvertime: "6:30:00",
				},
			}},
			expectedCalculatedBalance: hoursDuration(t, 5.5),
			expectedDefinitiveBalance: durationPtr(hoursDuration(t, 6.5)),
		},
		"HasPreviousPayslip_NoCurrentPayslip": {
			givenReport: Report{
				From:    time.Date(2021, time.March, 01, 0, 0, 0, 0, time.UTC),
				Summary: Summary{TotalOvertime: hoursDuration(t, 5.5)},
			},
			givenPayslips: model.PayslipList{Items: []model.Payslip{
				{
					DateFrom:  odoo.NewDate(2021, time.February, 01, 0, 0, 0, time.UTC),
					DateTo:    odoo.NewDate(2021, time.February, 28, 0, 0, 0, time.UTC),
					XOvertime: "6:30:00",
				},
			}},
			expectedCalculatedBalance: hoursDuration(t, 12.0),
			expectedDefinitiveBalance: nil,
		},
		"HasUnfinishedPreviousPayslip": {
			givenReport: Report{
				From:    time.Date(2021, time.March, 01, 0, 0, 0, 0, time.UTC),
				Summary: Summary{TotalOvertime: hoursDuration(t, 5.5)},
			},
			givenPayslips: model.PayslipList{Items: []model.Payslip{
				{
					// payslip got created, but no overtime info yet
					DateFrom: odoo.NewDate(2021, time.February, 01, 0, 0, 0, time.UTC),
					DateTo:   odoo.NewDate(2021, time.February, 28, 0, 0, 0, time.UTC),
				},
			}},
			expectedCalculatedBalance: hoursDuration(t, 5.5),
			expectedDefinitiveBalance: nil,
		},
		"HasPrevious_AndUnfinishedCurrentPayslips": {
			givenReport: Report{
				From:    time.Date(2021, time.March, 01, 0, 0, 0, 0, time.UTC),
				Summary: Summary{TotalOvertime: hoursDuration(t, 5.5)},
			},
			givenPayslips: model.PayslipList{Items: []model.Payslip{
				{
					DateFrom:  odoo.NewDate(2021, time.February, 01, 0, 0, 0, time.UTC),
					DateTo:    odoo.NewDate(2021, time.February, 28, 0, 0, 0, time.UTC),
					XOvertime: "6:30:00",
				},
				{
					// payslip got created, but no overtime info yet
					DateFrom: odoo.NewDate(2021, time.March, 01, 0, 0, 0, time.UTC),
					DateTo:   odoo.NewDate(2021, time.March, 31, 0, 0, 0, time.UTC),
				},
			}},
			expectedCalculatedBalance: hoursDuration(t, 12.0),
			expectedDefinitiveBalance: nil,
		},
		"HasPrevious_AndCurrentPayslips": {
			givenReport: Report{
				From:    time.Date(2021, time.March, 01, 0, 0, 0, 0, time.UTC),
				Summary: Summary{TotalOvertime: hoursDuration(t, 5.5)},
			},
			givenPayslips: model.PayslipList{Items: []model.Payslip{
				{
					DateFrom:  odoo.NewDate(2021, time.February, 01, 0, 0, 0, time.UTC),
					DateTo:    odoo.NewDate(2021, time.February, 28, 0, 0, 0, time.UTC),
					XOvertime: "6:30:00",
				},
				{
					DateFrom:  odoo.NewDate(2021, time.March, 01, 0, 0, 0, time.UTC),
					DateTo:    odoo.NewDate(2021, time.March, 31, 0, 0, 0, time.UTC),
					XOvertime: "5:00:00",
				},
			}},
			expectedCalculatedBalance: hoursDuration(t, 12.0),
			expectedDefinitiveBalance: durationPtr(hoursDuration(t, 5.0)),
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			b := NewBalanceReportBuilder(tc.givenReport, tc.givenPayslips)
			result, err := b.CalculateBalanceReport()
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedCalculatedBalance, result.CalculatedBalance, "calculated balance")
			assert.Equal(t, tc.expectedDefinitiveBalance, result.DefinitiveBalance, "definitive balance")
		})
	}
}

func durationPtr(d time.Duration) *time.Duration {
	return &d
}
