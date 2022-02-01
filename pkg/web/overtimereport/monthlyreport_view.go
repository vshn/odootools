package overtimereport

import (
	"fmt"
	"strconv"
	"time"

	"github.com/vshn/odootools/pkg/odoo"
	"github.com/vshn/odootools/pkg/odoo/model"
	"github.com/vshn/odootools/pkg/timesheet"
	"github.com/vshn/odootools/pkg/web/controller"
)

const monthlyReportTemplateName string = "overtimereport-monthly"

type reportView struct {
}

func (v *reportView) formatDailySummary(daily *timesheet.DailySummary) controller.Values {
	basic := controller.Values{
		"Weekday":       daily.Date.Weekday(),
		"Date":          daily.Date.Format(odoo.DateFormat),
		"Workload":      daily.FTERatio * 100,
		"ExcusedHours":  formatDurationInHours(daily.CalculateExcusedTime()),
		"WorkedHours":   formatDurationInHours(daily.CalculateWorkingTime()),
		"OvertimeHours": formatDurationInHours(daily.CalculateOvertime()),
		"LeaveType":     "",
	}
	if daily.HasAbsences() {
		basic["LeaveType"] = daily.Absences[0].Reason
	}
	return basic
}

// formatDurationInHours returns a human friendly "0:00"-formatted duration.
// Seconds within a minute are rounded up or down to the nearest full minute.
// A sign ("-") is prefixed if duration is negative.
func formatDurationInHours(d time.Duration) string {
	sign := ""
	if d.Seconds() < 0 {
		sign = "-"
		d = time.Duration(d.Nanoseconds() * -1)
	}
	d = d.Round(time.Minute)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	return fmt.Sprintf("%s%d:%02d", sign, h, m)
}

func formatFloat(v float64) string {
	return strconv.FormatFloat(v, 'f', 1, 64)
}

func (v *reportView) formatMonthlySummary(s timesheet.Summary, payslip *model.Payslip) controller.Values {
	val := controller.Values{
		"TotalOvertime": formatDurationInHours(s.TotalOvertime),
		"TotalLeaves":   fmt.Sprintf("%sd", formatFloat(s.TotalLeave)),
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
			val["NewOvertimeBalance"] = formatDurationInHours(lastMonthBalance + s.TotalOvertime)
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
	nextYear, nextMonth := getNextMonth(report)
	prevYear, prevMonth := getPreviousMonth(report)
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

func getNextMonth(r timesheet.MonthlyReport) (int, int) {
	if r.Month >= 12 {
		return r.Year + 1, 1
	}
	return r.Year, r.Month + 1
}

func getPreviousMonth(r timesheet.MonthlyReport) (int, int) {
	if r.Month <= 1 {
		return r.Year - 1, 12
	}
	return r.Year, r.Month - 1
}
