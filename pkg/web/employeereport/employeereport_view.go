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
		reportValues[i] = v.getValuesForReport(report.Result, report.Payslip)
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
		"Reports": reportValues,
		"Warning": v.formatErrorForFailedEmployeeReports(failedEmployees),
		"Year":    v.year,
		"Month":   time.Month(v.month).String(),
	}
}

func (v *reportView) getValuesForReport(report timesheet.MonthlyReport, payslip *model.Payslip) controller.Values {
	overtimeBalance := ""
	if payslip == nil {
		overtimeBalance = "no payslip found"
	} else {
		balance, err := payslip.ParseOvertime()
		if err != nil {
			overtimeBalance = err.Error()
		} else {
			overtimeBalance = v.FormatDurationInHours(balance + report.Summary.TotalOvertime)
		}
	}
	return controller.Values{
		"Name":             report.Employee.Name,
		"ReportDirectLink": fmt.Sprintf("/report/%d/%d/%02d", report.Employee.ID, v.year, v.month),
		"Workload":         v.FormatFloat(report.Summary.AverageWorkload*100, 0),
		"Leaves":           report.Summary.TotalLeave,
		"ExcusedHours":     v.FormatDurationInHours(report.Summary.TotalExcusedTime),
		"WorkedHours":      v.FormatDurationInHours(report.Summary.TotalWorkedTime),
		"OvertimeHours":    v.FormatDurationInHours(report.Summary.TotalOvertime),
		"OvertimeBalance":  overtimeBalance,
	}
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
