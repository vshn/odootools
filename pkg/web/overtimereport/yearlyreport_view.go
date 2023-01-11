package overtimereport

import (
	"fmt"
	"time"

	"github.com/vshn/odootools/pkg/timesheet"
	"github.com/vshn/odootools/pkg/web/controller"
)

type yearlyReportView struct {
	controller.BaseView
}

const yearlyReportTemplateName string = "overtimereport-yearly"

func (v *yearlyReportView) GetValuesForYearlyReport(report timesheet.YearlyReport) controller.Values {
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

func (v *yearlyReportView) formatMonthlySummaryForYearlyReport(s timesheet.BalanceReport, year int) controller.Values {
	defBalance := ""
	if s.DefinitiveBalance != nil {
		defBalance = v.FormatDurationInHours(*s.DefinitiveBalance)
	}
	validationErrorList := &timesheet.ValidationErrorList{}
	for _, summary := range s.Report.DailySummaries {
		timesheet.AppendValidationError(validationErrorList, summary.ValidateTimesheetEntries())
	}
	val := controller.Values{
		"OvertimeHours":     v.FormatDurationInHours(s.Report.Summary.TotalOvertime),
		"LeaveDays":         v.FormatFloat(s.Report.Summary.TotalLeave, 1),
		"ExcusedHours":      v.FormatDurationInHours(s.Report.Summary.TotalExcusedTime),
		"WorkedHours":       v.FormatDurationInHours(s.Report.Summary.TotalWorkedTime),
		"DefinitiveBalance": defBalance,
		"DetailViewLink":    fmt.Sprintf("/report/%d/%d/%d", s.Report.Employee.ID, year, s.Report.From.Month()),
		"Name":              fmt.Sprintf("%s %d", s.Report.From.Month(), year),
		"ValidationError":   validationErrorList.Error(),
		"OvertimeClassname": v.OvertimeClassname(s.Report.Summary.TotalOvertime),
	}
	return val
}

func (v *yearlyReportView) formatYearlySummary(summary timesheet.YearlySummary) controller.Values {
	val := controller.Values{
		"TotalExcused":      v.FormatDurationInHours(summary.TotalExcused),
		"TotalWorked":       v.FormatDurationInHours(summary.TotalWorked),
		"TotalOvertime":     v.FormatDurationInHours(summary.TotalOvertime),
		"TotalLeaves":       v.FormatFloat(summary.TotalLeaves, 1),
		"OvertimeClassname": v.OvertimeClassname(summary.TotalOvertime),
	}
	return val
}
