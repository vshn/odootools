package html

import (
	"net/http"
	"strconv"

	"github.com/mhutter/vshn-ftb/pkg/odoo"
	"github.com/mhutter/vshn-ftb/pkg/timesheet"
)

type DashboardView struct {
	renderer *Renderer
	template string
}

func NewDashboardView(renderer *Renderer) *DashboardView {
	return &DashboardView{
		renderer: renderer,
		template: "dashboard",
	}
}

func (v *DashboardView) formatEntry(daily timesheet.DailySummary) Values {
	basic := Values{
		"Weekday":       daily.Date.Weekday(),
		"Date":          daily.Date.Format(odoo.AttendanceDateFormat),
		"OvertimeHours": strconv.FormatFloat(daily.CalculateOvertime(), 'f', 2, 64),
	}
	return basic
}

func (v *DashboardView) formatSummary(s timesheet.Summary) Values {
	return Values{
		"TotalOvertime": s.TotalWorkedHours,
	}
}

func (v *DashboardView) ShowAttendanceReport(w http.ResponseWriter, report timesheet.Report) {
	w.WriteHeader(http.StatusOK)
	v.renderer.Render(w, v.template, v.prepareValues(report))
}

func (v *DashboardView) ShowError(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	v.renderer.Render(w, v.template, Values{
		"Error": err.Error(),
	})
}

func (v *DashboardView) prepareValues(report timesheet.Report) Values {
	formatted := make([]Values, len(report.DailySummaries))
	for i := range report.DailySummaries {
		formatted[i] = v.formatEntry(report.DailySummaries[i])
	}
	return Values{
		"Attendances": formatted,
		"Summary":     v.formatSummary(report.Summary),
	}
}
