package overtimereport

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
	"github.com/vshn/odootools/pkg/web/reportconfig"
)

type ReportController struct {
	controller.BaseController
	Input           reportconfig.ReportRequest
	Employee        *model.Employee
	ReportView      *reportView
	Contracts       model.ContractList
	Attendances     model.AttendanceList
	Leaves          odoo.List[model.Leave]
	PreviousPayslip *model.Payslip
	NextPayslip     *model.Payslip
	User            *model.User
}

func NewReportController(ctx *controller.BaseController) *ReportController {
	return &ReportController{
		BaseController: *ctx,
		ReportView:     &reportView{},
	}
}

// DisplayOvertimeReport GET /report/:id/:year/:month
func (c *ReportController) DisplayOvertimeReport() error {
	root := pipeline.NewPipeline[context.Context]()
	root.WithSteps(
		root.NewStep("parse user input", c.parseInput),
		root.NewStep("fetch employee", c.fetchEmployeeByID),
		root.NewStep("fetch payslips", c.fetchPayslips),
		root.NewStep("fetch user settings", c.fetchUser),
		root.NewStep("fetch contracts", c.fetchContracts),
		root.NewStep("fetch attendances", c.fetchAttendances),
		root.NewStep("fetch leaves", c.fetchLeaves),
		root.When(pipeline.Not(c.noMonthGiven), "calculate monthly report", c.calculateMonthlyReport),
		root.When(c.noMonthGiven, "calculate yearly report", c.calculateYearlyReport),
	)
	err := root.RunWithContext(c.RequestContext)
	return err
}

func (c *ReportController) parseInput(_ context.Context) error {
	input := reportconfig.ReportRequest{}
	err := input.FromRequest(c.Echo)
	c.Input = input
	return err
}

func (c *ReportController) fetchEmployeeByID(ctx context.Context) error {
	employeeID := c.Input.EmployeeID
	if c.SessionData.Employee != nil && c.SessionData.Employee.ID == employeeID {
		c.Employee = c.SessionData.Employee
		return nil
	}

	employee, err := c.OdooClient.FetchEmployeeByID(ctx, employeeID)
	if employee == nil {
		return fmt.Errorf("no employee found with given ID: %d", employeeID)
	}
	c.Employee = employee
	return err
}

func (c *ReportController) fetchUser(ctx context.Context) error {
	user, err := c.OdooClient.FetchUserByID(ctx, c.OdooSession.UID)
	c.User = user
	return err
}

func (c *ReportController) fetchContracts(ctx context.Context) error {
	contracts, err := c.OdooClient.FetchAllContractsOfEmployee(ctx, c.Employee.ID)
	c.Contracts = contracts
	return err
}

func (c *ReportController) fetchAttendances(ctx context.Context) error {
	tz := c.getTimeZone()
	begin, end := c.Input.GetDateRange()
	attendances, err := c.OdooClient.FetchAttendancesBetweenDates(ctx, c.Employee.ID, begin.In(tz), end.In(tz))
	c.Attendances = attendances.AddCurrentTimeAsSignOut(tz)
	return err
}

func (c *ReportController) fetchLeaves(ctx context.Context) error {
	tz := c.getTimeZone()
	begin, end := c.Input.GetDateRange()
	leaves, err := c.OdooClient.FetchLeavesBetweenDates(ctx, c.Employee.ID, begin.In(tz), end.In(tz))
	c.Leaves = leaves
	return err
}

func (c *ReportController) calculateMonthlyReport(_ context.Context) error {
	tz := c.getTimeZone()
	start := time.Date(c.Input.Year, time.Month(c.Input.Month), 1, 0, 0, 0, 0, tz)
	end := start.AddDate(0, 1, 0)
	reporter := timesheet.NewReporter(c.Attendances, c.Leaves, c.Employee, c.Contracts).
		SetRange(start, end).
		SetTimeZone(tz)

	report, err := reporter.CalculateReport()
	if err != nil {
		return err
	}
	values := c.ReportView.GetValuesForMonthlyReport(report, c.PreviousPayslip, c.NextPayslip)
	return c.Echo.Render(http.StatusOK, monthlyReportTemplateName, values)
}

func (c *ReportController) getTimeZone() *time.Location {
	if c.NextPayslip != nil && !c.NextPayslip.TimeZone.IsEmpty() {
		// timezone from payslip has precedence.
		return c.NextPayslip.TimeZone.Location()
	}
	if c.User != nil && c.SessionData.Employee.ID == c.Employee.ID && time.Now().Month() == time.Month(c.Input.Month) {
		// get the timezone from user preferences only if we create a report for our own user AND we're in the current month.
		// for months long in the past we don't want to calculate based on user's current preferences.
		return c.User.TimeZone.LocationOrDefault(controller.DefaultTimeZone)
	}
	// last resort to default TZ.
	return controller.DefaultTimeZone
}

func (c *ReportController) calculateYearlyReport(_ context.Context) error {
	reporter := timesheet.NewYearlyReporter(c.Attendances, c.Leaves, c.Employee, c.Contracts).
		SetYear(c.Input.Year)
	reporter.SetTimeZone(c.User.TimeZone.Location())
	report, err := reporter.CalculateYearlyReport()
	if err != nil {
		return err
	}
	values := c.ReportView.GetValuesForYearlyReport(report)
	return c.Echo.Render(http.StatusOK, yearlyReportTemplateName, values)
}

func (c *ReportController) fetchPayslips(ctx context.Context) error {
	lastMonth := c.Input.GetFirstDayOfMonth().AddDate(0, -1, 0)
	payslips, err := c.OdooClient.FetchPayslipBetween(ctx, c.Employee.ID, lastMonth, c.Input.GetLastDayOfMonth())
	if payslips.Len() >= 1 {
		c.PreviousPayslip = &payslips.Items[0]
	}
	if payslips.Len() >= 2 {
		c.NextPayslip = &payslips.Items[1]
	}
	return err
}

func (c *ReportController) noMonthGiven(_ context.Context) bool {
	return c.Input.Month == 0
}
