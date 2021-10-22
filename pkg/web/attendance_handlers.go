package web

import (
	"net/http"
	"strconv"
	"time"

	"github.com/mhutter/vshn-ftb/pkg/timesheet"
	"github.com/mhutter/vshn-ftb/pkg/web/html"
)

// OvertimeReport GET /report
func (s Server) OvertimeReport() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session := s.sessionFrom(r)
		if session == nil {
			// User is unauthenticated
			http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
			return
		}
		view := html.NewOvertimeReportView(s.html)

		attendances, err := s.odoo.ReadAllAttendances(session.ID, session.UID)
		if err != nil {
			view.ShowError(w, err)
			return
		}

		reporter := timesheet.NewReport()
		reporter.SetAttendances(attendances)

		year := parseIntOrDefault(r.FormValue("year"), time.Now().Year())
		month := parseIntOrDefault(r.FormValue("month"), int(time.Now().Month()))
		fte := parseFloatOrDefault(r.FormValue("ftepercentage"), 100)

		report := reporter.CalculateReportForMonth(year, month, fte/100)
		view.ShowAttendanceReport(w, report)
	})
}

func parseIntOrDefault(toParse string, def int) int {
	if toParse == "" {
		return def
	}
	if v, err := strconv.Atoi(toParse); err == nil {
		return v
	}
	return def
}

func parseFloatOrDefault(toParse string, def float64) float64 {
	if toParse == "" {
		return def
	}
	if v, err := strconv.ParseFloat(toParse, 64); err == nil {
		return v
	}
	return def
}

func (s Server) RequestReportForm() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session := s.sessionFrom(r)
		if session == nil {
			// User is unauthenticated
			http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
			return
		}

		view := html.NewRequestReportView(s.html)
		view.ShowConfigurationForm(w)
	})
}
