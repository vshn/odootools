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

		year := parseOrDefault(r.FormValue("year"), time.Now().Year())
		month := parseOrDefault(r.FormValue("month"), int(time.Now().Month()))

		report := reporter.CalculateReportForMonth(year, month)
		view.ShowAttendanceReport(w, report)
	})
}

func parseOrDefault(toParse string, def int) int {
	if toParse == "" {
		return def
	}
	if v, err := strconv.Atoi(toParse); err == nil {
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
