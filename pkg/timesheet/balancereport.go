package timesheet

import (
	"fmt"
	"time"

	"github.com/vshn/odootools/pkg/odoo/model"
)

type BalanceReport struct {
	Report Report
	// PreviousBalance is the definitive balance from the previous month's payslip, if given.
	PreviousBalance time.Duration
	// CalculatedBalance is the sum of TotalOvertime with the previous month's payslip, if given.
	CalculatedBalance time.Duration
	// DefinitiveBalance contains the value of the current month's payslip, if given.
	DefinitiveBalance *time.Duration
}

type BalanceReportBuilder struct {
	report   Report
	payslips model.PayslipList
}

func NewBalanceReportBuilder(report Report, payslips model.PayslipList) *BalanceReportBuilder {
	return &BalanceReportBuilder{
		report:   report,
		payslips: payslips,
	}
}

func (b *BalanceReportBuilder) CalculateBalanceReport() (BalanceReport, error) {
	r := BalanceReport{Report: b.report}

	// calculated balance
	previousMonth := b.payslips.FilterInMonth(r.Report.From.AddDate(0, -1, 0))
	if previousMonth != nil {
		parsed, err := previousMonth.ParseOvertime()
		if err != nil {
			return r, fmt.Errorf("cannot parse overtime of payslip '%s': %w", previousMonth.Name, err)
		}
		r.PreviousBalance = parsed
	}
	r.CalculatedBalance = r.PreviousBalance + r.Report.Summary.TotalOvertime

	// definitive balance
	currentMonth := b.payslips.FilterInMonth(r.Report.From.AddDate(0, 0, 1)) // add a day to cover timezone offsets
	if currentMonth != nil && currentMonth.Overtime() != "" {
		parsed, err := currentMonth.ParseOvertime()
		if err != nil {
			return r, fmt.Errorf("cannot parse overtime of payslip '%s': %w", currentMonth.Name, err)
		}
		r.DefinitiveBalance = &parsed
	}
	return r, nil
}
