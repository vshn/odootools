package reportconfig

import (
	"context"
	"fmt"
	"net/http"
	"time"

	pipeline "github.com/ccremer/go-command-pipeline"
	"github.com/vshn/odootools/pkg/odoo"
	"github.com/vshn/odootools/pkg/odoo/model"
	"github.com/vshn/odootools/pkg/timesheet"
	"github.com/vshn/odootools/pkg/web/controller"
)

type ConfigController struct {
	controller.BaseController
	Input       ReportRequest
	view        *ConfigView
	Employee    *model.Employee
	Attendances model.AttendanceList
	StartOfWeek time.Time
	EndOfWeek   time.Time
	Leaves      odoo.List[model.Leave]
	Contracts   model.ContractList
	Report      timesheet.Report
}

func NewConfigController(ctrl *controller.BaseController) *ConfigController {
	return &ConfigController{
		BaseController: *ctrl,
		view:           &ConfigView{},
	}
}

func (c *ConfigController) ShowConfigurationFormAndWeeklyReport() error {
	c.view.roles = c.SessionData.Roles
	root := pipeline.NewPipeline[context.Context]()
	root.WithSteps(
		root.NewStep("parse user input", c.parseInput),
		root.WithNestedSteps("weekly report", pipeline.Bool[context.Context](true),
			root.NewStep("fetch attendances", c.fetchAttendanceOfCurrentWeek),
			root.NewStep("fetch contracts", c.fetchContracts),
			root.NewStep("fetch leaves", c.fetchLeaves),
			root.NewStep("calculate report", c.calculateReport),
		).WithErrorHandler(c.displayWarning),
		root.NewStep("render", c.render),
	)
	err := root.RunWithContext(c.RequestContext)
	return err
}

func (c *ConfigController) ProcessInput() error {
	root := pipeline.NewPipeline[context.Context]()
	root.WithSteps(
		root.NewStep("parse user input", c.parseInput),
		root.NewStep("search employee", c.searchEmployee),
		root.NewStep("redirect to report", c.redirectToReportView),
	)
	err := root.RunWithContext(c.RequestContext)
	return err
}

func (c *ConfigController) parseInput(_ context.Context) error {
	input := ReportRequest{}
	err := input.FromRequest(c.Echo)
	c.Input = input

	today := time.Now()
	//today := time.Date(2021, time.December, 5, 4, 5, 6, 0, time.Local)
	monday := getStartOfWeek(today)
	sunday := getEndOfWeek(today).AddDate(0, 0, 1)

	c.StartOfWeek = monday
	c.EndOfWeek = sunday
	return err
}

func (c *ConfigController) searchEmployee(ctx context.Context) error {
	if c.Input.SearchUserEnabled {
		e, err := c.OdooClient.SearchEmployee(ctx, c.Input.SearchUser)
		if e == nil {
			return fmt.Errorf("no user matching '%s' found", c.Input.SearchUser)
		}
		c.Employee = e
		return err
	}
	if c.SessionData.Employee != nil {
		c.Employee = c.SessionData.Employee
		return nil
	}
	return fmt.Errorf("no Employee found for user ID %q", c.OdooSession.UID)
}

func (c *ConfigController) redirectToReportView(_ context.Context) error {
	if c.Input.EmployeeReportEnabled {
		return c.Echo.Redirect(http.StatusFound, fmt.Sprintf("/report/employees/%d/%02d", c.Input.Year, c.Input.Month))
	}
	if c.Input.Month == 0 {
		return c.Echo.Redirect(http.StatusFound, fmt.Sprintf("/report/%d/%d", c.Employee.ID, c.Input.Year))
	}
	return c.Echo.Redirect(http.StatusFound, fmt.Sprintf("/report/%d/%d/%02d", c.Employee.ID, c.Input.Year, c.Input.Month))
}

func (c *ConfigController) render(_ context.Context) error {
	return c.Echo.Render(http.StatusOK, configViewTemplate, c.view.GetConfigurationValues(c.Report))
}

func (c *ConfigController) fetchAttendanceOfCurrentWeek(ctx context.Context) error {
	attendances, err := c.OdooClient.FetchAttendancesBetweenDates(ctx, c.SessionData.Employee.ID, c.StartOfWeek, c.EndOfWeek)
	if err != nil {
		return err
	}
	attendances.SortByDate()
	attendances = attendances.FilterAttendanceBetweenDates(c.StartOfWeek, c.EndOfWeek)
	c.Attendances = attendances
	if len(attendances.Items) > 0 {
		lastAttendance := attendances.Items[len(attendances.Items)-1]
		if lastAttendance.Action == timesheet.ActionSignIn {
			c.view.isSignedIn = true
			now := odoo.Date(time.Now())
			// fake a sign_out
			c.Attendances.Items = append(c.Attendances.Items, model.Attendance{
				DateTime: &now,
				Action:   timesheet.ActionSignOut,
				Reason:   lastAttendance.Reason,
			})
		}
	}
	return nil
}

func (c *ConfigController) fetchContracts(ctx context.Context) error {
	contracts, err := c.OdooClient.FetchAllContractsOfEmployee(ctx, c.SessionData.Employee.ID)
	c.Contracts = contracts
	return err
}

func (c *ConfigController) fetchLeaves(ctx context.Context) error {
	leaves, err := c.OdooClient.FetchLeavesBetweenDates(ctx, c.SessionData.Employee.ID, c.StartOfWeek, c.EndOfWeek)
	c.Leaves = leaves
	return err
}

func (c *ConfigController) calculateReport(_ context.Context) error {
	reporter := timesheet.NewReporter(c.Attendances, c.Leaves, c.Employee, c.Contracts).
		SetRange(c.StartOfWeek, c.EndOfWeek).
		SetTimeZone("Europe/Zurich"). // hardcoded for now
		SkipClampingToNow(true)
	report, err := reporter.CalculateReport()
	c.Report = report
	return err
}

func (c *ConfigController) displayWarning(_ context.Context, err error) error {
	if err != nil {
		c.view.warning = err.Error()
	}
	return nil
}

// getStartOfWeek returns the previously occurred Monday at midnight.
// If t is already a Monday, it will be truncated to midnight the same day.
func getStartOfWeek(t time.Time) time.Time {
	t = t.Truncate(24 * time.Hour)
	if t.Weekday() == time.Sunday { // go treats Sunday as the first day of the week
		return t.AddDate(0, 0, -6)
	}
	diff := (t.Weekday() - 1) * -1
	return t.AddDate(0, 0, int(diff))
}

// getEndOfWeek returns the next occurring Sunday at midnight.
// If t is already a Sunday, it will be truncated to midnight the same day.
func getEndOfWeek(t time.Time) time.Time {
	return getStartOfWeek(t).AddDate(0, 0, 6)
}
