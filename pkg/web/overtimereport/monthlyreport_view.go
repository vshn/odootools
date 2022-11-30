package overtimereport

import (
	"fmt"
	"time"

	"github.com/vshn/odootools/pkg/timesheet"
	"github.com/vshn/odootools/pkg/web/controller"
)

const monthlyReportTemplateName string = "overtimereport-monthly"

type monthlyReportView struct {
	controller.BaseView
}

func (v *monthlyReportView) GetValuesForMonthlyReport(report timesheet.BalanceReport) controller.Values {
	formatted := make([]controller.Values, 0)
	hasInvalidAttendances := ""
	for _, summary := range report.Report.DailySummaries {
		if summary.IsWeekend() && summary.CalculateOvertimeSummary().WorkingTime() == 0 {
			continue
		}
		values := v.FormatDailySummary(summary)
		if values["ValidationError"] != nil {
			hasInvalidAttendances = "Your timesheet contains errors."
		}
		formatted = append(formatted, values)
	}
	month, year := report.Report.From.Month(), report.Report.From.Year()
	nextYear, nextMonth := v.GetNextMonth(year, int(month))
	prevYear, prevMonth := v.GetPreviousMonth(year, int(month))
	linkFormat := "/report/%d/%d/%02d"
	return controller.Values{
		"Attendances": formatted,
		"Warning":     hasInvalidAttendances,
		"Summary":     v.formatMonthlySummary(report),
		"Nav": controller.Values{
			"LoggedIn":          true,
			"ActiveView":        monthlyReportTemplateName,
			"CurrentMonthLink":  fmt.Sprintf(linkFormat, report.Report.Employee.ID, time.Now().Year(), time.Now().Month()),
			"NextMonthLink":     fmt.Sprintf(linkFormat, report.Report.Employee.ID, nextYear, nextMonth),
			"PreviousMonthLink": fmt.Sprintf(linkFormat, report.Report.Employee.ID, prevYear, prevMonth),
		},
		"Username":            report.Report.Employee.Name,
		"MonthDisplayName":    fmt.Sprintf("%s %d", month, year),
		"TimezoneDisplayName": report.Report.From.Location().String(),
	}
}

func (v *monthlyReportView) formatMonthlySummary(report timesheet.BalanceReport) controller.Values {
	s := report.Report.Summary
	val := controller.Values{
		"TotalOvertime": v.FormatDurationInHours(s.TotalOvertime),
		"TotalLeaves":   fmt.Sprintf("%sd", v.FormatFloat(s.TotalLeave, 1)),
		"TotalWorked":   v.FormatDurationInHours(s.TotalWorkedTime),
		"TotalExcused":  v.FormatDurationInHours(s.TotalExcusedTime),
	}
	val["PreviousBalance"] = v.FormatDurationInHours(report.PreviousBalance)
	val["NewOvertimeBalance"] = v.FormatDurationInHours(report.CalculatedBalance)
	val["CurrentPayslipBalance"] = ""
	if report.DefinitiveBalance != nil {
		val["CurrentPayslipBalance"] = v.FormatDurationInHours(*report.DefinitiveBalance)
	}
	return val
}
