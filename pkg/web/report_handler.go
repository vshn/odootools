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

type OvertimeInput struct {
	Year              int
	Month             int
	SearchUser        string
	SearchUserEnabled bool
}

func (i *OvertimeInput) FromForm(r *http.Request) {
	i.SearchUserEnabled = r.FormValue("userscope") == "user-foreign-radio"
	i.SearchUser = html.EscapeString(r.FormValue("username"))

	i.Year = parseIntOrDefault(r.FormValue("year"), time.Now().Year())
	i.Month = parseIntOrDefault(r.FormValue("month"), int(time.Now().Month()))
}

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

		input := OvertimeInput{}
		input.FromForm(r)

		employee := getEmployee(w, input, s, session, view)
		if employee == nil {
			// error already shown in view
			return
		}

		begin := odoo.Date(time.Date(input.Year, time.Month(input.Month), 1, 0, 0, 0, 0, time.UTC))
		end := odoo.Date(begin.ToTime().AddDate(0, 1, 0))

		contracts, err := s.odoo.FetchAllContracts(session.ID, employee.ID)
		if err != nil {
			view.ShowError(w, err)
			return
		}

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

		reporter := timesheet.NewReporter(attendances, leaves, employee, contracts).SetMonth(input.Year, input.Month)
		report := reporter.CalculateReport()
		view.ShowAttendanceReport(w, report)
	})
}

func getEmployee(w http.ResponseWriter, input OvertimeInput, s Server, session *odoo.Session, view *views.OvertimeReportView) *odoo.Employee {
	var employee *odoo.Employee
	if input.SearchUserEnabled {
		e, err := s.odoo.SearchEmployee(input.SearchUser, session.ID)
		if err != nil {
			view.ShowError(w, err)
			return nil
		}
		if e == nil {
			view.ShowError(w, fmt.Errorf("no user matching '%s' found", input.SearchUser))
			return nil
		}
		employee = e
		return employee
	}
	e, err := s.odoo.FetchEmployee(session.ID, session.UID)
	if err != nil {
		view.ShowError(w, err)
		return nil
	}
	employee = e
	return employee
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
