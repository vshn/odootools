package web

import (
	"fmt"
	"html"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
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

	vars := mux.Vars(r)

	year := time.Now().Year()
	month := int(time.Now().Month())

	if yearFromURL, hasYear := vars["year"]; hasYear {
		year = parseIntOrDefault(yearFromURL, year)
	}
	if monthFromURL, found := vars["month"]; found {
		month = parseIntOrDefault(monthFromURL, month)
	}

	i.Year = parseIntOrDefault(r.FormValue("year"), year)
	i.Month = parseIntOrDefault(r.FormValue("month"), month)
}

func (s Server) RedirectReport() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		session := s.sessionFrom(req)
		if session == nil {
			// User is unauthenticated
			http.Redirect(w, req, "/login", http.StatusTemporaryRedirect)
			return
		}
		input := OvertimeInput{}
		input.FromForm(req)

		view := views.NewErrorView(s.html)
		employee := s.searchEmployee(w, input, session, view)
		if employee == nil {
			// error already shown in view
			return
		}

		http.Redirect(w, req, fmt.Sprintf("/report/%d/%d/%02d", employee.ID, input.Year, input.Month), http.StatusFound)
	})
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
		reportView := views.NewOvertimeReportView(s.html)
		errorView := views.NewErrorView(s.html)

		input := OvertimeInput{}
		input.FromForm(r)

		employee := s.getEmployeeOrAbort(w, r, errorView, session)
		if employee == nil {
			return
		}

		firstDay := time.Date(input.Year, time.Month(input.Month), 1, 0, 0, 0, 0, time.UTC)
		// Let's get attendances within a month with +- 1 day to respect localized dates and filter them later.
		begin := firstDay.AddDate(0, 0, -1)
		end := firstDay.AddDate(0, 1, 0)

		contracts, err := s.odoo.FetchAllContracts(session.ID, employee.ID)
		if err != nil {
			errorView.ShowError(w, err)
			return
		}

		attendances, err := s.odoo.FetchAttendancesBetweenDates(session.ID, employee.ID, begin, end)
		if err != nil {
			errorView.ShowError(w, err)
			return
		}

		leaves, err := s.odoo.FetchLeavesBetweenDates(session.ID, employee.ID, begin, end)
		if err != nil {
			errorView.ShowError(w, err)
			return
		}

		reporter := timesheet.NewReporter(attendances, leaves, employee, contracts).
			SetMonth(input.Year, input.Month).
			SetTimeZone("Europe/Zurich") // hardcoded for now
		report := reporter.CalculateReport()
		reportView.ShowAttendanceReport(w, report)
	})
}

func (s Server) getEmployeeOrAbort(w http.ResponseWriter, r *http.Request, errorView *views.ErrorView, session *odoo.Session) *odoo.Employee {
	employeeID := getEmployeeIDFromURL(w, r, errorView)
	if employeeID == 0 {
		return nil
	}

	employee, err := s.odoo.FetchEmployeeByID(session.ID, employeeID)
	if err != nil {
		errorView.ShowError(w, err)
		return nil
	}
	if employee == nil {
		errorView.ShowError(w, fmt.Errorf("no employee found with given ID: %d", employeeID))
		return nil
	}
	return employee
}

func getEmployeeIDFromURL(w http.ResponseWriter, r *http.Request, errorView *views.ErrorView) int {
	vars := mux.Vars(r)
	v, found := vars["employee"]
	if !found {
		errorView.ShowError(w, fmt.Errorf("no employee ID provided in URL"))
		return 0
	}
	employeeID, err := strconv.Atoi(v)
	if err != nil {
		errorView.ShowError(w, err)
		return 0
	}
	return employeeID
}

func (s Server) searchEmployee(w http.ResponseWriter, input OvertimeInput, session *odoo.Session, view *views.ErrorView) *odoo.Employee {
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
		return e
	}
	e, err := s.odoo.FetchEmployeeBySession(session)
	if err != nil {
		view.ShowError(w, err)
		return nil
	}
	return e
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
