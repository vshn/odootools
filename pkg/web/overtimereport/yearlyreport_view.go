package overtimereport

import (
	"fmt"
	"time"

	"github.com/vshn/odootools/pkg/timesheet"
	"github.com/vshn/odootools/pkg/web/controller"
)

const yearlyReportTemplateName string = "overtimereport-yearly"

func (v *reportView) GetValuesForYearlyReport(report timesheet.YearlyReport) controller.Values {
	formatted := make([]controller.Values, 0)
	for _, month := range report.MonthlyReports {
		formatted = append(formatted, v.formatMonthlySummaryForYearlyReport(month))
	}
	nextYear := report.Year + 1
	prevYear := report.Year - 1
	linkFormat := "/report/%d/%d"
	return controller.Values{
		"MonthlyReports": formatted,
		"Summary":        v.formatYearlySummary(report.Summary),
		"Nav": controller.Values{
			"LoggedIn":         true,
			"ActiveView":       yearlyReportTemplateName,
			"CurrentYearLink":  fmt.Sprintf(linkFormat, report.Employee.ID, time.Now().Year()),
			"NextYearLink":     fmt.Sprintf(linkFormat, report.Employee.ID, nextYear),
			"PreviousYearLink": fmt.Sprintf(linkFormat, report.Employee.ID, prevYear),
		},
		"Username": report.Employee.Name,
	}
}

func (v *reportView) formatMonthlySummaryForYearlyReport(s timesheet.MonthlyReport) controller.Values {
	val := controller.Values{
		"OvertimeHours":  formatDurationInHours(s.Summary.TotalOvertime),
		"LeaveDays":      formatFloat(s.Summary.TotalLeave),
		"ExcusedHours":   formatDurationInHours(s.Summary.TotalExcusedTime),
		"WorkedHours":    formatDurationInHours(s.Summary.TotalWorkedTime),
		"DetailViewLink": fmt.Sprintf("/report/%d/%d/%d", s.Employee.ID, s.Year, s.Month),
		"Name":           time.Month(s.Month).String(),
	}
	return val
}

func (v *reportView) formatYearlySummary(summary timesheet.YearlySummary) controller.Values {
	return controller.Values{
		"TotalExcused":  formatDurationInHours(summary.TotalExcused),
		"TotalWorked":   formatDurationInHours(summary.TotalWorked),
		"TotalOvertime": formatDurationInHours(summary.TotalOvertime),
		"TotalLeaves":   formatFloat(summary.TotalLeaves),
	}
}
