package web

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/vshn/odootools/pkg/odoo"
	"github.com/vshn/odootools/pkg/timesheet"
	"github.com/vshn/odootools/pkg/web/html"
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

		forAnotherUser := r.FormValue("userscope") == "user-foreign-radio"
		searchUser := r.FormValue("username")

		userId := session.UID
		var employee *odoo.Employee
		if forAnotherUser {
			e, err := s.odoo.SearchEmployee(searchUser, session.ID)
			if err != nil {
				view.ShowError(w, err)
				return
			}
			if e == nil {
				view.ShowError(w, fmt.Errorf("no user matching '%s' found", searchUser))
				return
			}
			employee = e
		} else {
			e, err := s.odoo.FetchEmployee(userId, session.ID)
			if err != nil {
				view.ShowError(w, err)
				return
			}
			employee = e
		}

		attendances, err := s.odoo.ReadAllAttendances(session.ID, userId)
		if err != nil {
			view.ShowError(w, err)
			return
		}

		leaves, err := s.odoo.ReadAllLeaves(session.ID, userId)
		if err != nil {
			view.ShowError(w, err)
			return
		}

		year := parseIntOrDefault(r.FormValue("year"), time.Now().Year())
		month := parseIntOrDefault(r.FormValue("month"), int(time.Now().Month()))
		fte := parseFloatOrDefault(r.FormValue("ftepercentage"), 100)

		reporter := timesheet.NewReporter(attendances, leaves, employee).SetFteRatio(fte/100).SetMonth(year, month)
		report := reporter.CalculateReport()
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
