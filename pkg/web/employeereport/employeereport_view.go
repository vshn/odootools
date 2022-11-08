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
		"Reports":       reportValues,
		"Warning":       v.formatErrorForFailedEmployeeReports(failedEmployees),
		"Year":          v.year,
		"Month":         time.Month(v.month).String(),
		"LastMonth":     time.Month(prevMonth).String(),
		"UpdateBaseUrl": fmt.Sprintf("/report/employee/:employee/%d/%02d", v.year, v.month),
	}
}

func (v *reportView) getValuesForReport(report timesheet.Report, previousPayslip, nextPayslip *model.Payslip) controller.Values {
	previousBalanceCellText, previousBalance := v.getPreviousBalance(previousPayslip)
	proposedBalanceCellText, proposedBalance := v.getProposedBalance(previousBalance, report.Summary.TotalOvertime)
	nextBalanceCellText, nextBalance := v.getNextBalance(proposedBalance, nextPayslip)
	overtimeBalanceEditPreview := v.getOvertimeBalanceEditPreview(nextPayslip, nextBalance)
	return controller.Values{
		"Name":                            report.Employee.Name,
		"EmployeeID":                      report.Employee.ID,
		"ReportDirectLink":                fmt.Sprintf("/report/%d/%d/%02d", report.Employee.ID, v.year, v.month),
		"ButtonText":                      v.getButtonText(nextPayslip),
		"Workload":                        v.FormatFloat(report.Summary.AverageWorkload*100, 0),
		"Leaves":                          report.Summary.TotalLeave,
		"ExcusedHours":                    v.FormatDurationInHours(report.Summary.TotalExcusedTime),
		"WorkedHours":                     v.FormatDurationInHours(report.Summary.TotalWorkedTime),
		"OutOfOfficeHours":                v.FormatDurationInHours(report.Summary.TotalOutOfOfficeTime),
		"OvertimeHours":                   v.FormatDurationInHours(report.Summary.TotalOvertime),
		"PreviousBalance":                 previousBalanceCellText,
		"NextBalance":                     nextBalanceCellText,
		"ProposedBalance":                 proposedBalanceCellText,
		"OvertimeBalanceEditEnabled":      nextPayslip != nil,
		"OvertimeBalanceEditPreviewValue": overtimeBalanceEditPreview,
	}
}

func (v *reportView) getOvertimeBalanceEditPreview(nextPayslip *model.Payslip, proposedBalance time.Duration) (overtimeBalanceEditPreview string) {
	overtimeBalanceEditPreview = v.FormatDurationInHours(proposedBalance)
	if nextPayslip != nil {
		// next payslip exists
		if existingValue := nextPayslip.Overtime(); existingValue != "" {
			// payslip may have been updated already
			return existingValue
		}
	}
	return
}

func (v *reportView) getButtonText(nextPayslip *model.Payslip) string {
	if nextPayslip == nil || nextPayslip.Overtime() == "" {
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

func (v *reportView) getPreviousBalance(previousPayslip *model.Payslip) (cellText string, previousOvertime time.Duration) {
	if previousPayslip == nil {
		cellText = "<no payslip found>"
		return
	}
	if previousPayslip.Overtime() == "" {
		cellText = "<no overtime saved>"
		return
	}
	previousOvertime, err := previousPayslip.ParseOvertime()
	if err != nil {
		cellText = fmt.Sprintf("<%v>", err)
		previousOvertime = 0
		return
	}
	cellText = previousPayslip.Overtime()
	return
}

func (v *reportView) getProposedBalance(previousBalance time.Duration, overtime time.Duration) (cellText string, predictedOvertime time.Duration) {
	predictedOvertime = previousBalance + overtime
	cellText = v.FormatDurationInHours(predictedOvertime)
	return
}

func (v *reportView) getNextBalance(proposedBalance time.Duration, nextPayslip *model.Payslip) (cellText string, nextOvertime time.Duration) {
	if nextPayslip == nil {
		return "<no payslip found>", proposedBalance
	}
	if existing := nextPayslip.Overtime(); existing != "" {
		cellText = existing
		parsed, err := nextPayslip.ParseOvertime()
		if err != nil {
			cellText = fmt.Sprintf("<%v>", err)
			nextOvertime = proposedBalance
			return
		}
		nextOvertime = parsed
		return
	}
	return "", proposedBalance
}
