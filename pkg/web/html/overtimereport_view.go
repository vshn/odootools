package html

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
		"OvertimeHours": strconv.FormatFloat(daily.CalculateOvertime().Hours(), 'f', 2, 64),
		"LeaveType":     "",
	}
	if daily.HasAbsences() {
		basic["LeaveType"] = daily.Absences[0].Reason
	}
	return basic
}

func (v *OvertimeReportView) formatSummary(s timesheet.Summary) Values {
	return Values{
		"TotalOvertime": s.TotalOvertime.Truncate(time.Minute),
		// TODO: Might not be accurate for days before 2021
		"TotalLeaves": fmt.Sprintf("%sd", strconv.FormatFloat(s.TotalLeaveDays.Hours()/8, 'f', 0, 64)),
	}
}

func (v *OvertimeReportView) ShowAttendanceReport(w http.ResponseWriter, report timesheet.Report) {
	w.WriteHeader(http.StatusOK)
	v.renderer.Render(w, v.template, v.prepareValues(report))
}

func (v *OvertimeReportView) ShowError(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	v.renderer.Render(w, v.template, Values{
		"Error": err.Error(),
	})
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
	}
}
