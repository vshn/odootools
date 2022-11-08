package overtimereport

import (
	"fmt"
	"time"

	"github.com/vshn/odootools/pkg/odoo/model"
	"github.com/vshn/odootools/pkg/timesheet"
	"github.com/vshn/odootools/pkg/web/controller"
)

const monthlyReportTemplateName string = "overtimereport-monthly"

type reportView struct {
	controller.BaseView
}

func (v *reportView) formatMonthlySummary(s timesheet.Summary, previousPayslip *model.Payslip, currentPayslip *model.Payslip) controller.Values {
	val := controller.Values{
		"TotalOvertime": v.FormatDurationInHours(s.TotalOvertime),
		"TotalLeaves":   fmt.Sprintf("%sd", v.FormatFloat(s.TotalLeave, 1)),
		"TotalWorked":   v.FormatDurationInHours(s.TotalWorkedTime),
		"TotalExcused":  v.FormatDurationInHours(s.TotalExcusedTime),
	}
	if previousPayslip == nil {
		val["PayslipError"] = "No matching payslip found"
	} else {
		lastMonthBalance, err := previousPayslip.ParseOvertime()
		if err != nil {
			val["PayslipError"] = err.Error()
		}
		if lastMonthBalance == 0 {
			val["PayslipError"] = "No overtime saved in payslip"
		} else {
			val["NewOvertimeBalance"] = v.FormatDurationInHours(lastMonthBalance + s.TotalOvertime)
		}
	}
	if currentPayslip != nil {
		if currentBalance, err := currentPayslip.ParseOvertime(); err == nil {
			if currentBalance == 0 {
				val["CurrentPayslipBalance"] = ""
			} else {
				val["CurrentPayslipBalance"] = v.FormatDurationInHours(currentBalance)
			}
		}
	}
	return val
}

func (v *reportView) GetValuesForMonthlyReport(report timesheet.Report, previousPayslip, nextPayslip *model.Payslip) controller.Values {
	formatted := make([]controller.Values, 0)
	for _, summary := range report.DailySummaries {
		if summary.IsWeekend() && summary.CalculateOvertimeSummary().WorkingTime() == 0 {
			continue
		}
		formatted = append(formatted, v.FormatDailySummary(summary))
	}
	nextYear, nextMonth := v.GetNextMonth(report.From.Year(), int(report.From.Month()))
	prevYear, prevMonth := v.GetPreviousMonth(report.From.Year(), int(report.From.Month()))
	linkFormat := "/report/%d/%d/%02d"
	return controller.Values{
		"Attendances": formatted,
		"Summary":     v.formatMonthlySummary(report.Summary, previousPayslip, nextPayslip),
		"Nav": controller.Values{
			"LoggedIn":          true,
			"ActiveView":        monthlyReportTemplateName,
			"CurrentMonthLink":  fmt.Sprintf(linkFormat, report.Employee.ID, time.Now().Year(), time.Now().Month()),
			"NextMonthLink":     fmt.Sprintf(linkFormat, report.Employee.ID, nextYear, nextMonth),
			"PreviousMonthLink": fmt.Sprintf(linkFormat, report.Employee.ID, prevYear, prevMonth),
		},
		"Username":            report.Employee.Name,
		"MonthDisplayName":    fmt.Sprintf("%s %d", report.From.Month(), report.From.Year()),
		"TimezoneDisplayName": report.From.Location().String(),
	}
}
