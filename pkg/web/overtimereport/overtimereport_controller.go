package overtimereport

import (
	"context"
	"fmt"
	"net/http"

	pipeline "github.com/ccremer/go-command-pipeline"
	"github.com/vshn/odootools/pkg/odoo/model"
	"github.com/vshn/odootools/pkg/timesheet"
	"github.com/vshn/odootools/pkg/web/controller"
	"github.com/vshn/odootools/pkg/web/reportconfig"
)

type ReportController struct {
	controller.BaseController
	Input       reportconfig.ReportRequest
	Employee    *model.Employee
	ReportView  *reportView
	Contracts   model.ContractList
	Attendances model.AttendanceList
	Leaves      model.LeaveList
	Payslip     *model.Payslip
}

func NewReportController(ctx *controller.BaseController) *ReportController {
	return &ReportController{
		BaseController: *ctx,
		ReportView:     &reportView{},
	}
}

// DisplayOvertimeReport GET /report/:id/:year/:month
func (c *ReportController) DisplayOvertimeReport() error {
	root := pipeline.NewPipeline().
		WithSteps(
			pipeline.NewStepFromFunc("parse user input", c.parseInput),
			pipeline.NewStepFromFunc("fetch employee", c.fetchEmployeeByID),
			pipeline.NewStepFromFunc("fetch contracts", c.fetchContracts),
			pipeline.NewStepFromFunc("fetch attendances", c.fetchAttendances),
			pipeline.NewStepFromFunc("fetch leaves", c.fetchLeaves),
			pipeline.NewStepFromFunc("fetch last issued payslip", c.fetchPayslip),
			pipeline.If(pipeline.Not(c.noMonthGiven), pipeline.NewStepFromFunc("calculate monthly report", c.calculateMonthlyReport)),
			pipeline.If(c.noMonthGiven, pipeline.NewStepFromFunc("calculate yearly report", c.calculateYearlyReport)),
		)
	result := root.RunWithContext(c.Echo.Request().Context())
	return result.Err()
}

func (c *ReportController) parseInput(_ context.Context) error {
	input := reportconfig.ReportRequest{}
	err := input.FromRequest(c.Echo)
	c.Input = input
	return err
}

func (c *ReportController) fetchEmployeeByID(_ context.Context) error {
	employeeID := c.Input.EmployeeID
	if c.SessionData.Employee != nil && c.SessionData.Employee.ID == employeeID {
		c.Employee = c.SessionData.Employee
		return nil
	}

	employee, err := c.OdooClient.FetchEmployeeByID(employeeID)
	if employee == nil {
		return fmt.Errorf("no employee found with given ID: %d", employeeID)
	}
	c.Employee = employee
	return err
}

func (c *ReportController) fetchContracts(_ context.Context) error {
	contracts, err := c.OdooClient.FetchAllContracts(c.Employee.ID)
	c.Contracts = contracts
	return err
}

func (c *ReportController) fetchAttendances(_ context.Context) error {
	begin, end := c.Input.GetDateRange()
	attendances, err := c.OdooClient.FetchAttendancesBetweenDates(c.Employee.ID, begin, end)
	c.Attendances = attendances
	return err
}

func (c *ReportController) fetchLeaves(_ context.Context) error {
	begin, end := c.Input.GetDateRange()
	leaves, err := c.OdooClient.FetchLeavesBetweenDates(c.Employee.ID, begin, end)
	c.Leaves = leaves
	return err
}

func (c *ReportController) calculateMonthlyReport(_ context.Context) error {
	reporter := timesheet.NewReporter(c.Attendances, c.Leaves, c.Employee, c.Contracts).
		SetMonth(c.Input.Year, c.Input.Month).
		SetTimeZone("Europe/Zurich") // hardcoded for now
	report, err := reporter.CalculateMonthlyReport()
	if err != nil {
		return err
	}
	values := c.ReportView.GetValuesForMonthlyReport(report, c.Payslip)
	return c.Echo.Render(http.StatusOK, monthlyReportTemplateName, values)
}

func (c *ReportController) calculateYearlyReport(_ context.Context) error {
	reporter := timesheet.NewReporter(c.Attendances, c.Leaves, c.Employee, c.Contracts).
		SetMonth(c.Input.Year, 1).
		SetTimeZone("Europe/Zurich") // hardcoded for now
	report, err := reporter.CalculateYearlyReport()
	if err != nil {
		return err
	}
	values := c.ReportView.GetValuesForYearlyReport(report)
	return c.Echo.Render(http.StatusOK, yearlyReportTemplateName, values)
}

func (c *ReportController) searchEmployee(_ context.Context) error {
	if c.Input.SearchUserEnabled {
		e, err := c.OdooClient.SearchEmployee(c.Input.SearchUser)
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

func (c *ReportController) fetchPayslip(_ context.Context) error {
	lastMonth := c.Input.GetFirstDayOfNextMonth().AddDate(0, -1, 0)
	payslip, err := c.OdooClient.FetchPayslipOfLastMonth(c.Employee.ID, lastMonth)
	c.Payslip = payslip
	return err
}

func (c *ReportController) noMonthGiven(_ context.Context) bool {
	return c.Input.Month == 0
}
