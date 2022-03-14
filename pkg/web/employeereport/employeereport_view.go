package employeereport

import (
	"fmt"
	"strings"
	"time"

	"github.com/vshn/odootools/pkg/odoo/model"
	"github.com/vshn/odootools/pkg/timesheet"
	"github.com/vshn/odootools/pkg/web/controller"
)

const employeeReportTemplateName = "employeereport"

type reportView struct {
	controller.BaseView
	year  int
	month int
}

func (v *reportView) GetValuesForReports(reports []*EmployeeReport, failedEmployees []*model.Employee) controller.Values {
	reportValues := make([]controller.Values, len(reports))
	for i, report := range reports {
		reportValues[i] = v.getValuesForReport(report.Result, report.PreviousPayslip, report.NextPayslip)
	}
	nextYear, nextMonth := v.GetNextMonth(v.year, v.month)
	prevYear, prevMonth := v.GetPreviousMonth(v.year, v.month)
	linkFormat := "/report/employees/%d/%02d"
	return controller.Values{
		"Nav": controller.Values{
			"LoggedIn":          true,
			"ActiveView":        employeeReportTemplateName,
			"PreviousMonthLink": fmt.Sprintf(linkFormat, prevYear, prevMonth),
			"NextMonthLink":     fmt.Sprintf(linkFormat, nextYear, nextMonth),
			"CurrentMonthLink":  fmt.Sprintf(linkFormat, time.Now().Year(), time.Now().Month()),
		},
		"Reports":   reportValues,
		"Warning":   v.formatErrorForFailedEmployeeReports(failedEmployees),
		"Year":      v.year,
		"Month":     time.Month(v.month).String(),
		"LastMonth": time.Month(prevMonth).String(),
	}
}

func (v *reportView) getValuesForReport(report timesheet.MonthlyReport, previousPayslip, nextPayslip *model.Payslip) controller.Values {
	overtimeBalance, overtimeBalanceEditPreview := v.getOvertimeBalancePreview(report, previousPayslip, nextPayslip)
	return controller.Values{
		"Name":                            report.Employee.Name,
		"EmployeeID":                      report.Employee.ID,
		"ReportDirectLink":                fmt.Sprintf("/report/%d/%d/%02d", report.Employee.ID, v.year, v.month),
		"OvertimeBalanceEditEnabled":      nextPayslip != nil,
		"OvertimeBalanceEditPreviewValue": overtimeBalanceEditPreview,
		"ButtonText":                      v.getButtonText(previousPayslip, nextPayslip),
		"Workload":                        v.FormatFloat(report.Summary.AverageWorkload*100, 0),
		"Leaves":                          report.Summary.TotalLeave,
		"ExcusedHours":                    v.FormatDurationInHours(report.Summary.TotalExcusedTime),
		"WorkedHours":                     v.FormatDurationInHours(report.Summary.TotalWorkedTime),
		"OvertimeHours":                   v.FormatDurationInHours(report.Summary.TotalOvertime),
		"PreviousBalance":                 v.getPreviousBalance(previousPayslip),
		"NextBalance":                     v.getNextBalance(nextPayslip),
		"PredictedBalance":                overtimeBalance,
	}
}

func (v *reportView) getOvertimeBalancePreview(report timesheet.MonthlyReport, previousPayslip *model.Payslip, nextPayslip *model.Payslip) (overtimeBalance string, overtimeBalanceEditPreview string) {
	if previousPayslip == nil {
		// new VSHNeer?
		overtimeBalance = v.FormatDurationInHours(report.Summary.TotalOvertime)
		overtimeBalanceEditPreview = v.FormatDurationInHours(report.Summary.TotalOvertime)
		return
	}
	balance, err := previousPayslip.ParseOvertime()
	if err != nil {
		// error case: show error + just this month's total overtime preview
		overtimeBalance = err.Error()
		overtimeBalanceEditPreview = v.FormatDurationInHours(report.Summary.TotalOvertime)
		return
	}
	// we have new valid overtime balance
	overtimeBalance = v.FormatDurationInHours(balance + report.Summary.TotalOvertime)
	// edit preview to be determined next
	overtimeBalanceEditPreview = "create payslip first"
	if nextPayslip != nil {
		// next payslip exists
		if existingValue := nextPayslip.GetOvertime(); existingValue == "" {
			// new overtime balance proposal
			overtimeBalanceEditPreview = overtimeBalance
		} else {
			// payslip may have been updated already
			overtimeBalanceEditPreview = existingValue
		}
	}
	return
}

func (v *reportView) getButtonText(previousPayslip *model.Payslip, nextPayslip *model.Payslip) string {
	if previousPayslip == nil || nextPayslip.GetOvertime() == "" {
		return "Save (New)"
	}
	return "Save (Update)"
}

func (v *reportView) formatErrorForFailedEmployeeReports(employees []*model.Employee) string {
	if len(employees) == 0 {
		return ""
	}
	names := make([]string, len(employees))
	for i, report := range employees {
		names[i] = report.Name
	}
	list := strings.Join(names, ", ")
	return fmt.Sprintf("reports failed for following employees: %v. Most probably due to missing contracts.", list)
}

func (v *reportView) getPreviousBalance(payslip *model.Payslip) string {
	if payslip == nil {
		return "no payslip found"
	}
	return payslip.GetOvertime()
}

func (v *reportView) getNextBalance(payslip *model.Payslip) string {
	if payslip == nil {
		return "create payslip first"
	}
	return payslip.GetOvertime()
}
