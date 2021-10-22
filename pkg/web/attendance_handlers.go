package web

import (
	"net/http"
	"time"

	"github.com/mhutter/vshn-ftb/pkg/odoo"
	"github.com/mhutter/vshn-ftb/pkg/timesheet"
	"github.com/mhutter/vshn-ftb/pkg/web/html"
)

// Dashboard GET /
func (s Server) Dashboard() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session := s.sessionFrom(r)
		if session == nil {
			// User is unauthenticated
			http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
			return
		}
		dashboard := html.NewDashboardView(s.html)

		attendances, err := s.odoo.ReadAllAttendances(session.ID, session.UID)
		if err != nil {
			dashboard.ShowError(w, err)
			return
		}

		reporter := timesheet.NewReport()
		reporter.SetAttendances(attendances)

		report := reporter.CalculateReportForMonth(2021, 6)
		dashboard.ShowAttendanceReport(w, report)
	})
}

func filterAttendancesToCurrentMonth(attendances []odoo.Attendance) []odoo.Attendance {
	year, _, _ := time.Now().Date()
	currentMonth := time.Date(year, 2, 1, 0, 0, 1, 0, time.Now().Location())
	nextMonth := currentMonth.AddDate(0, 1, 0)
	filtered := make([]odoo.Attendance, 0)
	for _, a := range attendances {
		if a.DateTime.ToTime().After(currentMonth) && a.DateTime.ToTime().Before(nextMonth) {
			filtered = append(filtered, a)
		}
	}
	return filtered
}
