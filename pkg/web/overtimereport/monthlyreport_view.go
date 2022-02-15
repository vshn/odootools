package overtimereport

import (
	"fmt"
	"time"

	"github.com/vshn/odootools/pkg/odoo"
	"github.com/vshn/odootools/pkg/odoo/model"
	"github.com/vshn/odootools/pkg/timesheet"
	"github.com/vshn/odootools/pkg/web/controller"
)

const monthlyReportTemplateName string = "overtimereport-monthly"

type reportView struct {
	controller.BaseView
}

func (v *reportView) formatDailySummary(daily *timesheet.DailySummary) controller.Values {
	basic := controller.Values{
		"Weekday":       daily.Date.Weekday(),
		"Date":          daily.Date.Format(odoo.DateFormat),
		"Workload":      daily.FTERatio * 100,
		"ExcusedHours":  v.FormatDurationInHours(daily.CalculateExcusedTime()),
		"WorkedHours":   v.FormatDurationInHours(daily.CalculateWorkingTime()),
		"OvertimeHours": v.FormatDurationInHours(daily.CalculateOvertime()),
		"LeaveType":     "",
	}
	if daily.HasAbsences() {
		basic["LeaveType"] = daily.Absences[0].Reason
	}
	return basic
}

func (v *reportView) formatMonthlySummary(s timesheet.Summary, payslip *model.Payslip) controller.Values {
	val := controller.Values{
		"TotalOvertime": v.FormatDurationInHours(s.TotalOvertime),
		"TotalLeaves":   fmt.Sprintf("%sd", v.FormatFloat(s.TotalLeave, 1)),
	}
	if payslip == nil {
		val["PayslipError"] = "No matching payslip found"
	} else {
		lastMonthBalance, err := payslip.ParseOvertime()
		if err != nil {
			val["PayslipError"] = err.Error()
		}
		if lastMonthBalance == 0 {
			val["PayslipError"] = "No overtime saved in payslip"
		} else {
			val["NewOvertimeBalance"] = v.FormatDurationInHours(lastMonthBalance + s.TotalOvertime)
		}
	}
	return val
}

func (v *reportView) GetValuesForMonthlyReport(report timesheet.MonthlyReport, payslip *model.Payslip) controller.Values {
	formatted := make([]controller.Values, 0)
	for _, summary := range report.DailySummaries {
		if summary.IsWeekend() && summary.CalculateWorkingTime() == 0 {
			continue
		}
		formatted = append(formatted, v.formatDailySummary(summary))
	}
	nextYear, nextMonth := v.GetNextMonth(report.Year, report.Month)
	prevYear, prevMonth := v.GetPreviousMonth(report.Year, report.Month)
	linkFormat := "/report/%d/%d/%02d"
	return controller.Values{
		"Attendances": formatted,
		"Summary":     v.formatMonthlySummary(report.Summary, payslip),
		"Nav": controller.Values{
			"LoggedIn":          true,
			"ActiveView":        monthlyReportTemplateName,
			"CurrentMonthLink":  fmt.Sprintf(linkFormat, report.Employee.ID, time.Now().Year(), time.Now().Month()),
			"NextMonthLink":     fmt.Sprintf(linkFormat, report.Employee.ID, nextYear, nextMonth),
			"PreviousMonthLink": fmt.Sprintf(linkFormat, report.Employee.ID, prevYear, prevMonth),
		},
		"Username": report.Employee.Name,
	}
}
