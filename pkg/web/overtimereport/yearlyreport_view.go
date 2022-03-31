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
		formatted = append(formatted, v.formatMonthlySummaryForYearlyReport(month, report.Year))
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

func (v *reportView) formatMonthlySummaryForYearlyReport(s timesheet.Report, year int) controller.Values {
	val := controller.Values{
		"OvertimeHours":  v.FormatDurationInHours(s.Summary.TotalOvertime),
		"LeaveDays":      v.FormatFloat(s.Summary.TotalLeave, 1),
		"ExcusedHours":   v.FormatDurationInHours(s.Summary.TotalExcusedTime),
		"WorkedHours":    v.FormatDurationInHours(s.Summary.TotalWorkedTime),
		"DetailViewLink": fmt.Sprintf("/report/%d/%d/%d", s.Employee.ID, year, s.From.Month()),
		"Name":           fmt.Sprintf("%s %d", s.From.Month(), year),
	}
	return val
}

func (v *reportView) formatYearlySummary(summary timesheet.YearlySummary) controller.Values {
	return controller.Values{
		"TotalExcused":  v.FormatDurationInHours(summary.TotalExcused),
		"TotalWorked":   v.FormatDurationInHours(summary.TotalWorked),
		"TotalOvertime": v.FormatDurationInHours(summary.TotalOvertime),
		"TotalLeaves":   v.FormatFloat(summary.TotalLeaves, 1),
	}
}
