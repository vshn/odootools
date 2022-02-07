package employeereport

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/vshn/odootools/pkg/odoo/model"
	"github.com/vshn/odootools/pkg/timesheet"
	"github.com/vshn/odootools/pkg/web/controller"
	"github.com/vshn/odootools/pkg/web/overtimereport"
)

const employeeReportTemplateName = "employeereport"

type reportView struct {
	year  int
	month int
}

func (v *reportView) GetValuesForReports(reports []timesheet.MonthlyReport, failedEmployees []*model.Employee) controller.Values {
	reportValues := make([]controller.Values, len(reports))
	for i, report := range reports {
		reportValues[i] = v.getValuesForReport(report)
	}
	return controller.Values{
		"Nav": controller.Values{
			"LoggedIn":   true,
			"ActiveView": employeeReportTemplateName,
		},
		"Reports": reportValues,
		"Warning": v.formatErrorForFailedEmployeeReports(failedEmployees),
		"Year":    v.year,
		"Month":   time.Month(v.month).String(),
	}
}

func (v *reportView) getValuesForReport(report timesheet.MonthlyReport) controller.Values {
	return controller.Values{
		"Name":             report.Employee.Name,
		"ReportDirectLink": fmt.Sprintf("/report/%d/%d/%d", report.Employee.ID, v.year, v.month),
		"Workload":         formatFloat(report.Summary.AverageWorkload * 100),
		"Leaves":           report.Summary.TotalLeave,
		"ExcusedHours":     overtimereport.FormatDurationInHours(report.Summary.TotalExcusedTime),
		"WorkedHours":      overtimereport.FormatDurationInHours(report.Summary.TotalWorkedTime),
		"OvertimeHours":    overtimereport.FormatDurationInHours(report.Summary.TotalOvertime),
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

func formatFloat(v float64) string {
	return strconv.FormatFloat(v, 'f', 0, 64)
}
