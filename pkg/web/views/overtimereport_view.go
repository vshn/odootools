package views

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/vshn/odootools/pkg/odoo"
	"github.com/vshn/odootools/pkg/timesheet"
)

type OvertimeReportView struct {
	renderer *Renderer
	template string
}

func NewOvertimeReportView(renderer *Renderer) *OvertimeReportView {
	return &OvertimeReportView{
		renderer: renderer,
		template: "overtimereport",
	}
}

func (v *OvertimeReportView) formatDailySummary(daily *timesheet.DailySummary) Values {
	basic := Values{
		"Weekday":       daily.Date.Weekday(),
		"Date":          daily.Date.Format(odoo.DateFormat),
		"Workload":      daily.FTERatio * 100,
		"ExcusedHours":  formatDurationInHours(timesheet.ToDuration(daily.CalculateExcusedHours())),
		"WorkedHours":   formatDurationInHours(timesheet.ToDuration(daily.CalculateWorkingHours())),
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

func (v *OvertimeReportView) formatSummary(s timesheet.Summary) Values {
	return Values{
		"TotalOvertime": formatDurationInHours(s.TotalOvertime.Truncate(time.Minute)),
		// TODO: Might not be accurate for days before 2021
		"TotalLeaves": fmt.Sprintf("%sd", strconv.FormatFloat(s.TotalLeaveDays.Hours()/8, 'f', 0, 64)),
	}
}

func (v *OvertimeReportView) ShowAttendanceReport(w http.ResponseWriter, report timesheet.Report) {
	w.WriteHeader(http.StatusOK)
	v.renderer.Render(w, v.template, v.prepareValues(report))
}

func (v *OvertimeReportView) prepareValues(report timesheet.Report) Values {
	formatted := make([]Values, 0)
	for _, summary := range report.DailySummaries {
		if summary.IsWeekend() && summary.CalculateWorkingHours() == 0 {
			continue
		}
		formatted = append(formatted, v.formatDailySummary(summary))
	}
	return Values{
		"Attendances": formatted,
		"Summary":     v.formatSummary(report.Summary),
		"Nav": Values{
			"LoggedIn":   true,
			"ActiveView": v.template,
		},
		"Username": report.Employee.Name,
	}
}
