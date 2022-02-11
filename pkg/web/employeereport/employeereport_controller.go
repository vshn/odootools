package employeereport

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	pipeline "github.com/ccremer/go-command-pipeline"
	"github.com/ccremer/go-command-pipeline/parallel"
	"github.com/hashicorp/go-multierror"
	"github.com/vshn/odootools/pkg/odoo"
	"github.com/vshn/odootools/pkg/odoo/model"
	"github.com/vshn/odootools/pkg/timesheet"
	"github.com/vshn/odootools/pkg/web/controller"
	"github.com/vshn/odootools/pkg/web/reportconfig"
)

type ReportController struct {
	controller.BaseController
	Input     reportconfig.ReportRequest
	employees model.EmployeeList
	reports   []*EmployeeReport
	view      *reportView
}

type EmployeeReport struct {
	controller.BaseController
	Start       time.Time
	Stop        time.Time
	Employee    model.Employee
	Contracts   model.ContractList
	Attendances model.AttendanceList
	Leaves      model.LeaveList
	Payslip     *model.Payslip

	Result timesheet.MonthlyReport
}

func NewEmployeeReportController(ctx *controller.BaseController) *ReportController {
	return &ReportController{
		BaseController: *ctx,
		view:           &reportView{},
	}
}

// DisplayEmployeeReport GET /report/employees/:year/:month
func (c *ReportController) DisplayEmployeeReport() error {
	root := pipeline.NewPipeline().WithOptions(pipeline.DisableErrorWrapping).WithSteps(
		pipeline.NewStepFromFunc("parse user input", c.parseInput),
		pipeline.NewStepFromFunc("fetch employees", c.fetchEmployees),
		parallel.NewWorkerPoolStep("generate reports for each employee", 4, c.createPipelinesForEachEmployee, c.collectReports),
		pipeline.NewStepFromFunc("render report", c.renderReport),
	)
	result := root.Run()
	return result.Err
}

func (c *ReportController) createPipelinesForEachEmployee(pipelines chan *pipeline.Pipeline) {
	defer close(pipelines)
	c.reports = make([]*EmployeeReport, len(c.employees.Items))
	for i, employee := range c.employees.Items {
		report := &EmployeeReport{
			BaseController: c.BaseController,
			Employee:       employee,
			Start:          c.Input.GetFirstDay(),
			Stop:           c.Input.GetLastDay(),
		}
		c.reports[i] = report
		pipe := report.createPipeline()
		pipelines <- pipe
	}
}

func (c *EmployeeReport) createPipeline() *pipeline.Pipeline {
	p := pipeline.NewPipeline()
	p.AddStep(p.WithNestedSteps(fmt.Sprintf("report for %q", c.Employee.Name),
		pipeline.NewStepFromFunc("fetch contracts", c.fetchContracts),
		pipeline.NewStepFromFunc("fetch attendances", c.fetchAttendances),
		pipeline.NewStepFromFunc("fetch leaves", c.fetchLeaves),
		pipeline.NewStepFromFunc("fetch last issued payslip", c.fetchPayslip),
		pipeline.NewStepFromFunc("calculate monthly report", c.calculateMonthlyReport).WithErrorHandler(c.ignoreNoContractFound),
	))
	return p
}

func (c *ReportController) collectReports(_ pipeline.Context, results map[uint64]pipeline.Result) pipeline.Result {
	var combined error
	for _, result := range results {
		if result.IsFailed() {
			combined = multierror.Append(combined, result.Err)
		}
	}
	return pipeline.Result{Err: combined}
}

func (c *ReportController) parseInput(_ pipeline.Context) error {
	input := reportconfig.ReportRequest{}
	err := input.FromRequest(c.Echo)
	c.Input = input
	return err
}

func (c *ReportController) fetchEmployees(_ pipeline.Context) error {
	list := model.EmployeeList{}
	err := c.OdooSession.SearchGenericModel(context.Background(), odoo.SearchReadModel{
		Model: "hr.employee",
		Domain: []odoo.Filter{
			[]string{"work_email", "ilike", "@vshn.ch"},
		},
		Fields: []string{"name"},
	}, &list)
	c.employees = list
	return err
}

func (c *ReportController) renderReport(_ pipeline.Context) error {
	successfulReports := make([]timesheet.MonthlyReport, 0)
	failedReports := make([]*model.Employee, 0)
	for _, report := range c.reports {
		if report.Result.DailySummaries != nil {
			successfulReports = append(successfulReports, report.Result)
		} else {
			failedReports = append(failedReports, report.Result.Employee)
		}
	}
	c.view.year, c.view.month = c.Input.Year, c.Input.Month
	return c.Echo.Render(http.StatusOK, employeeReportTemplateName, c.view.GetValuesForReports(successfulReports, failedReports))
}

func (c *EmployeeReport) fetchContracts(_ pipeline.Context) error {
	contracts, err := c.OdooClient.FetchAllContracts(c.Employee.ID)
	c.Contracts = contracts
	return err
}

func (c *EmployeeReport) fetchAttendances(_ pipeline.Context) error {
	attendances, err := c.OdooClient.FetchAttendancesBetweenDates(c.Employee.ID, c.Start, c.Stop)
	c.Attendances = attendances
	return err
}

func (c *EmployeeReport) fetchLeaves(_ pipeline.Context) error {
	leaves, err := c.OdooClient.FetchLeavesBetweenDates(c.Employee.ID, c.Start, c.Stop)
	c.Leaves = leaves
	return err
}

func (c *EmployeeReport) calculateMonthlyReport(_ pipeline.Context) error {
	start := c.Start.AddDate(0, 0, 1)
	reporter := timesheet.NewReporter(c.Attendances, c.Leaves, &c.Employee, c.Contracts).
		SetMonth(start.Year(), int(start.Month())).
		SetTimeZone("Europe/Zurich") // hardcoded for now
	report, err := reporter.CalculateMonthlyReport()
	c.Result = report
	return err
}

func (c *EmployeeReport) fetchPayslip(_ pipeline.Context) error {
	// TODO: verify timestamp
	lastMonth := c.Start
	payslip, err := c.OdooClient.FetchPayslipOfLastMonth(c.Employee.ID, lastMonth)
	c.Payslip = payslip
	return err
}

func (c *EmployeeReport) ignoreNoContractFound(ctx pipeline.Context, err error) error {
	if strings.Contains(err.Error(), "no contract found that covers date") {
		return nil
	}
	return err
}
