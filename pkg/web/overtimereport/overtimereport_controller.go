package overtimereport

import (
	"fmt"
	"net/http"

	pipeline "github.com/ccremer/go-command-pipeline"
	"github.com/ccremer/go-command-pipeline/predicate"
	"github.com/vshn/odootools/pkg/odoo"
	"github.com/vshn/odootools/pkg/timesheet"
	"github.com/vshn/odootools/pkg/web/controller"
)

type ReportController struct {
	controller.Context
	Input       ReportRequest
	Employee    *odoo.Employee
	ReportView  *reportView
	Contracts   odoo.ContractList
	Attendances []odoo.Attendance
	Leaves      []odoo.Leave
	view        *reportView
	Payslip     *odoo.Payslip
}

func NewReportController(ctx *controller.Context) *ReportController {
	return &ReportController{
		Context: *ctx,
		view:    &reportView{},
	}
}

// DisplayOvertimeReport GET /report/:id/:year/:month
func (c *ReportController) DisplayOvertimeReport() error {
	root := pipeline.NewPipelineWithContext(c).
		WithSteps(
			pipeline.NewStepFromFunc("parse user input", c.parseInput),
			pipeline.NewStepFromFunc("fetch employee", c.fetchEmployeeByID),
			pipeline.NewStepFromFunc("fetch contracts", c.fetchContracts),
			pipeline.NewStepFromFunc("fetch attendances", c.fetchAttendances),
			pipeline.NewStepFromFunc("fetch leaves", c.fetchLeaves),
			pipeline.NewStepFromFunc("fetch last issued payslip", c.fetchPayslip),
			predicate.If(predicate.Not(c.noMonthGiven), pipeline.NewStepFromFunc("calculate monthly report", c.calculateMonthlyReport)),
			predicate.If(c.noMonthGiven, pipeline.NewStepFromFunc("calculate yearly report", c.calculateYearlyReport)),
		)
	result := root.Run()
	return result.Err
}

func (c *ReportController) ProcessInput() error {
	root := pipeline.NewPipelineWithContext(c).
		WithSteps(
			pipeline.NewStepFromFunc("parse user input", c.parseInput),
			pipeline.NewStepFromFunc("search employee", c.searchEmployee),
			pipeline.NewStepFromFunc("redirect to report", func(_ pipeline.Context) error {
				return c.Echo.Redirect(http.StatusFound, fmt.Sprintf("/report/%d/%d/%02d", c.Employee.ID, c.Input.Year, c.Input.Month))
			}),
		)
	result := root.Run()
	return result.Err
}

func (c *ReportController) parseInput(_ pipeline.Context) error {
	input := ReportRequest{}
	err := input.FromRequest(c.Echo)
	c.Input = input
	return err
}

func (c *ReportController) fetchEmployeeByID(_ pipeline.Context) error {
	employeeID := c.Input.EmployeeID
	employee, err := c.OdooClient.FetchEmployeeByID(c.OdooSession.ID, employeeID)
	if employee == nil {
		return fmt.Errorf("no employee found with given ID: %d", employeeID)
	}
	c.Employee = employee
	return err
}

func (c *ReportController) fetchContracts(_ pipeline.Context) error {
	contracts, err := c.OdooClient.FetchAllContracts(c.OdooSession.ID, c.Employee.ID)
	c.Contracts = contracts
	return err
}

func (c *ReportController) fetchAttendances(_ pipeline.Context) error {
	attendances, err := c.OdooClient.FetchAttendancesBetweenDates(c.OdooSession.ID, c.Employee.ID, c.Input.getFirstDay(), c.Input.getLastDay())
	c.Attendances = attendances
	return err
}

func (c *ReportController) fetchLeaves(_ pipeline.Context) error {
	leaves, err := c.OdooClient.FetchLeavesBetweenDates(c.OdooSession.ID, c.Employee.ID, c.Input.getFirstDay(), c.Input.getLastDay())
	c.Leaves = leaves
	return err
}

func (c *ReportController) calculateMonthlyReport(_ pipeline.Context) error {
	reporter := timesheet.NewReporter(c.Attendances, c.Leaves, c.Employee, c.Contracts).
		SetMonth(c.Input.Year, c.Input.Month).
		SetTimeZone("Europe/Zurich") // hardcoded for now
	report := reporter.CalculateMonthlyReport()
	values := c.ReportView.GetValuesForMonthlyReport(report, c.Payslip)
	return c.Echo.Render(http.StatusOK, monthlyReportTemplateName, values)
}

func (c *ReportController) calculateYearlyReport(_ pipeline.Context) error {
	reporter := timesheet.NewReporter(c.Attendances, c.Leaves, c.Employee, c.Contracts).
		SetMonth(c.Input.Year, 1).
		SetTimeZone("Europe/Zurich") // hardcoded for now
	report := reporter.CalculateYearlyReport()
	values := c.ReportView.GetValuesForYearlyReport(report)
	return c.Echo.Render(http.StatusOK, yearlyReportTemplateName, values)
}

func (c *ReportController) searchEmployee(_ pipeline.Context) error {
	if c.Input.SearchUserEnabled {
		e, err := c.OdooClient.SearchEmployee(c.Input.SearchUser, c.OdooSession.ID)
		if e == nil {
			return fmt.Errorf("no user matching '%s' found", c.Input.SearchUser)
		}
		c.Employee = e
		return err
	}
	e, err := c.OdooClient.FetchEmployeeBySession(c.OdooSession)
	c.Employee = e
	return err
}

func (c *ReportController) fetchPayslip(_ pipeline.Context) error {
	lastMonth := c.Input.getLastDay().AddDate(0, -1, 0)
	payslip, err := c.OdooClient.FetchPayslipOfLastMonth(c.OdooSession.ID, c.Employee.ID, lastMonth)
	c.Payslip = payslip
	return err
}

func (c *ReportController) noMonthGiven(_ pipeline.Context, _ pipeline.Step) bool {
	return c.Input.Month == 0
}
