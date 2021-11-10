package web

import (
	"fmt"
	"html"
	"net/http"
	"strconv"
	"time"

	"github.com/vshn/odootools/pkg/odoo"
	"github.com/vshn/odootools/pkg/timesheet"
	"github.com/vshn/odootools/pkg/web/views"
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
		view := views.NewOvertimeReportView(s.html)

		forAnotherUser := r.FormValue("userscope") == "user-foreign-radio"
		searchUser := html.EscapeString(r.FormValue("username"))

		year := parseIntOrDefault(r.FormValue("year"), time.Now().Year())
		month := parseIntOrDefault(r.FormValue("month"), int(time.Now().Month()))
		fte := parseFloatOrDefault(r.FormValue("ftepercentage"), 100)

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
			e, err := s.odoo.FetchEmployee(session.ID, session.UID)
			if err != nil {
				view.ShowError(w, err)
				return
			}
			employee = e
		}

		begin := odoo.Date(time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC))
		end := odoo.Date(begin.ToTime().AddDate(0, 1, 0))

		attendances, err := s.odoo.FetchAttendancesBetweenDates(session.ID, employee.ID, begin, end)
		if err != nil {
			view.ShowError(w, err)
			return
		}

		leaves, err := s.odoo.FetchLeavesBetweenDates(session.ID, employee.ID, begin, end)
		if err != nil {
			view.ShowError(w, err)
			return
		}

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

		view := views.NewRequestReportView(s.html)
		view.ShowConfigurationForm(w)
	})
}
