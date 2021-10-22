package html

import (
	"net/http"
	"strconv"
	"time"

	"github.com/mhutter/vshn-ftb/pkg/odoo"
	"github.com/mhutter/vshn-ftb/pkg/timesheet"
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
		"Date":          daily.Date.Format(odoo.AttendanceDateFormat),
		"OvertimeHours": strconv.FormatFloat(daily.Overtime.Hours(), 'f', 2, 64),
	}
	return basic
}

func (v *OvertimeReportView) formatSummary(s timesheet.Summary) Values {
	return Values{
		"TotalOvertime": s.TotalWorkedHours.Truncate(time.Minute),
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
	formatted := make([]Values, len(report.DailySummaries))
	for i := range report.DailySummaries {
		formatted[i] = v.formatDailySummary(report.DailySummaries[i])
	}
	return Values{
		"Attendances": formatted,
		"Summary":     v.formatSummary(report.Summary),
	}
}
