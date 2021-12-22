package web

import (
	"fmt"
	"html"
	"net/http"
	"time"

	pipeline "github.com/ccremer/go-command-pipeline"
	"github.com/gorilla/mux"
	"github.com/vshn/odootools/pkg/odoo"
	"github.com/vshn/odootools/pkg/timesheet"
	"github.com/vshn/odootools/pkg/web/views"
)

type ReportInput struct {
	Year              int
	Month             int
	SearchUser        string
	SearchUserEnabled bool
	EmployeeID        int
}

func (i *ReportInput) FromForm(r *http.Request) {
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

	if v, found := vars["employee"]; found {
		i.EmployeeID = parseIntOrDefault(v, 0)
	}
}

func (i ReportInput) getFirstDayOfMonth() time.Time {
	firstDay := time.Date(i.Year, time.Month(i.Month), 1, 0, 0, 0, 0, time.UTC)
	// Let's get attendances within a month with - 1 day to respect localized dates and filter them later.
	begin := firstDay.AddDate(0, 0, -1)
	return begin
}

func (i ReportInput) getLastDayOfMonth() time.Time {
	firstDay := time.Date(i.Year, time.Month(i.Month), 1, 0, 0, 0, 0, time.UTC)
	// Let's get attendances within a month with + 1 day to respect localized dates and filter them later.
	end := firstDay.AddDate(0, 1, 0)
	return end
}

type OverviewReportContext struct {
	*CommonContext
	Input       ReportInput
	Employee    *odoo.Employee
	ReportView  *views.OvertimeReportView
	Contracts   odoo.ContractList
	Attendances []odoo.Attendance
	Leaves      []odoo.Leave
}

func (c *OverviewReportContext) GetCommonContext() *CommonContext {
	return c.CommonContext
}

// OvertimeReport GET /report/{id}/{year}/{month}
func (s Server) OvertimeReport() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		ctx := &OverviewReportContext{
			CommonContext: &CommonContext{Request: r, Response: w, ErrorView: views.NewErrorView(s.html), OdooClient: s.odoo, Session: s.sessionFrom(r)},
			ReportView:    views.NewOvertimeReportView(s.html),
		}
		root := pipeline.NewPipelineWithContext(ctx).
			WithSteps(
				pipeline.NewStepFromFunc("check login", ctx.checkLogin),
				pipeline.NewStepFromFunc("parse user input", ctx.parseInput),
				pipeline.NewStepFromFunc("fetch employee", ctx.fetchEmployeeByID),
				pipeline.NewStepFromFunc("fetch contracts", ctx.fetchContracts),
				pipeline.NewStepFromFunc("fetch attendances", ctx.fetchAttendances),
				pipeline.NewStepFromFunc("fetch leaves", ctx.fetchLeaves),
				pipeline.NewStepFromFunc("calculate report", ctx.calculateReport),
			)
		result := root.Run()
		if result.IsFailed() {
			ctx.ErrorView.ShowError(w, result.Err)
		}
	})
}

func (OverviewReportContext) parseInput(ctx pipeline.Context) error {
	c := ctx.(*OverviewReportContext)
	input := ReportInput{}
	input.FromForm(c.Request)
	c.Input = input
	return nil
}

func (OverviewReportContext) fetchEmployeeByID(ctx pipeline.Context) error {
	c := ctx.(*OverviewReportContext)
	employeeID := c.Input.EmployeeID
	employee, err := c.OdooClient.FetchEmployeeByID(c.Session.ID, employeeID)
	if employee == nil {
		return fmt.Errorf("no employee found with given ID: %d", employeeID)
	}
	c.Employee = employee
	return err
}

func (OverviewReportContext) fetchContracts(ctx pipeline.Context) error {
	c := ctx.(*OverviewReportContext)
	contracts, err := c.OdooClient.FetchAllContracts(c.Session.ID, c.Employee.ID)
	c.Contracts = contracts
	return err
}

func (OverviewReportContext) fetchAttendances(ctx pipeline.Context) error {
	c := ctx.(*OverviewReportContext)

	attendances, err := c.OdooClient.FetchAttendancesBetweenDates(c.Session.ID, c.Employee.ID, c.Input.getFirstDayOfMonth(), c.Input.getLastDayOfMonth())
	c.Attendances = attendances
	return err
}

func (OverviewReportContext) fetchLeaves(ctx pipeline.Context) error {
	c := ctx.(*OverviewReportContext)

	leaves, err := c.OdooClient.FetchLeavesBetweenDates(c.Session.ID, c.Employee.ID, c.Input.getFirstDayOfMonth(), c.Input.getLastDayOfMonth())
	c.Leaves = leaves
	return err
}

func (OverviewReportContext) calculateReport(ctx pipeline.Context) error {
	c := ctx.(*OverviewReportContext)

	reporter := timesheet.NewReporter(c.Attendances, c.Leaves, c.Employee, c.Contracts).
		SetMonth(c.Input.Year, c.Input.Month).
		SetTimeZone("Europe/Zurich") // hardcoded for now
	report := reporter.CalculateReport()
	c.ReportView.ShowAttendanceReport(c.Response, report)
	return nil
}

func (OverviewReportContext) searchEmployee(ctx pipeline.Context) error {
	c := ctx.(*OverviewReportContext)

	if c.Input.SearchUserEnabled {
		e, err := c.OdooClient.SearchEmployee(c.Input.SearchUser, c.Session.ID)
		if e == nil {
			return fmt.Errorf("no user matching '%s' found", c.Input.SearchUser)
		}
		c.Employee = e
		return err
	}
	e, err := c.OdooClient.FetchEmployeeBySession(c.Session)
	c.Employee = e
	return err
}

// RequestReportForm GET /report
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

// RedirectReport POST /report
func (s Server) RedirectReport() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		ctx := &OverviewReportContext{
			CommonContext: &CommonContext{Request: req, Response: w, ErrorView: views.NewErrorView(s.html), OdooClient: s.odoo, Session: s.sessionFrom(req)},
		}
		root := pipeline.NewPipelineWithContext(ctx).
			WithSteps(
				pipeline.NewStepFromFunc("check login", ctx.checkLogin),
				pipeline.NewStepFromFunc("parse user input", ctx.parseInput),
				pipeline.NewStepFromFunc("search employee", ctx.searchEmployee),
				pipeline.NewStepFromFunc("redirect to report", func(ctx pipeline.Context) error {
					c := ctx.(*OverviewReportContext)
					http.Redirect(w, req, fmt.Sprintf("/report/%d/%d/%02d", c.Employee.ID, c.Input.Year, c.Input.Month), http.StatusFound)
					return nil
				}),
			)
		result := root.Run()
		if result.IsFailed() {
			ctx.ErrorView.ShowError(w, result.Err)
		}
	})
}
