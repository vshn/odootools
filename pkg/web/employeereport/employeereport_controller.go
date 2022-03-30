package employeereport

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	pipeline "github.com/ccremer/go-command-pipeline"
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
	employees odoo.List[model.Employee]
	reports   []*EmployeeReport
	view      *reportView
}

type EmployeeReport struct {
	controller.BaseController
	// Start is the first day of the month
	Start time.Time
	// Stop is the last day of the month
	Stop            time.Time
	Employee        model.Employee
	Contracts       model.ContractList
	Attendances     odoo.List[model.Attendance]
	Leaves          odoo.List[model.Leave]
	PreviousPayslip *model.Payslip
	NextPayslip     *model.Payslip

	Result timesheet.Report
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
		pipeline.NewWorkerPoolStep("generate reports for each employee", 4, c.createPipelinesForEachEmployee, c.collectReports),
		pipeline.NewStepFromFunc("render report", c.renderReport),
	)
	result := root.RunWithContext(c.RequestContext)
	return result.Err()
}

func (c *ReportController) createPipelinesForEachEmployee(ctx context.Context, pipelines chan *pipeline.Pipeline) {
	defer close(pipelines)
	c.reports = make([]*EmployeeReport, c.employees.Len())
	for i, employee := range c.employees.Items {
		select {
		case <-ctx.Done():
			return
		default:
			report := &EmployeeReport{
				BaseController: c.BaseController,
				Employee:       employee,
				Start:          c.Input.GetFirstDayOfMonth(),
				Stop:           c.Input.GetLastDayOfMonth(),
			}
			c.reports[i] = report
			pipe := report.createPipeline()
			pipelines <- pipe
		}
	}
}

func (c *EmployeeReport) createPipeline() *pipeline.Pipeline {
	p := pipeline.NewPipeline()
	p.AddStep(p.WithNestedSteps(fmt.Sprintf("report for %q", c.Employee.Name),
		pipeline.NewStepFromFunc("fetch contracts", c.fetchContracts),
		pipeline.NewStepFromFunc("fetch attendances", c.fetchAttendances),
		pipeline.NewStepFromFunc("fetch leaves", c.fetchLeaves),
		pipeline.NewStepFromFunc("fetch last issued payslip", c.fetchPreviousPayslip),
		pipeline.NewStepFromFunc("fetch current month's payslip", c.fetchNextPayslip),
		pipeline.NewStepFromFunc("calculate monthly report", c.calculateMonthlyReport).WithErrorHandler(c.ignoreNoContractFound),
	))
	return p
}

func (c *ReportController) collectReports(_ context.Context, results map[uint64]pipeline.Result) error {
	var combined error
	for _, result := range results {
		if result.IsFailed() {
			combined = multierror.Append(combined, result.Err())
		}
	}
	return combined
}

func (c *ReportController) parseInput(_ context.Context) error {
	input := reportconfig.ReportRequest{}
	err := input.FromRequest(c.Echo)
	c.Input = input
	return err
}

func (c *ReportController) fetchEmployees(ctx context.Context) error {
	list := odoo.List[model.Employee]{}
	err := c.OdooSession.SearchGenericModel(ctx, odoo.SearchReadModel{
		Model: "hr.employee",
		Domain: []odoo.Filter{
			[]string{"work_email", "ilike", "@vshn.ch"},
		},
		Fields: []string{"name"},
	}, &list)
	c.employees = list
	return err
}

func (c *ReportController) renderReport(_ context.Context) error {
	successfulReports := make([]*EmployeeReport, 0)
	failedReports := make([]*model.Employee, 0)
	for _, report := range c.reports {
		if report.Result.DailySummaries != nil {
			successfulReports = append(successfulReports, report)
		} else {
			failedReports = append(failedReports, report.Result.Employee)
		}
	}
	c.view.year, c.view.month = c.Input.Year, c.Input.Month
	return c.Echo.Render(http.StatusOK, employeeReportTemplateName, c.view.GetValuesForReports(successfulReports, failedReports))
}

func (c *EmployeeReport) fetchContracts(ctx context.Context) error {
	contracts, err := c.OdooClient.FetchAllContracts(ctx, c.Employee.ID)
	c.Contracts = contracts
	return err
}

func (c *EmployeeReport) fetchAttendances(ctx context.Context) error {
	// extend date range for timezone correction
	start := c.Start.AddDate(0, 0, -1)
	stop := c.Start.AddDate(0, 1, 0)
	attendances, err := c.OdooClient.FetchAttendancesBetweenDates(ctx, c.Employee.ID, start, stop)
	c.Attendances = attendances
	return err
}

func (c *EmployeeReport) fetchLeaves(ctx context.Context) error {
	// extend date range for timezone correction
	start := c.Start.AddDate(0, 0, -1)
	stop := c.Start.AddDate(0, 1, 0)
	leaves, err := c.OdooClient.FetchLeavesBetweenDates(ctx, c.Employee.ID, start, stop)
	c.Leaves = leaves
	return err
}

func (c *EmployeeReport) calculateMonthlyReport(_ context.Context) error {
	reporter := timesheet.NewReporter(c.Attendances, c.Leaves, &c.Employee, c.Contracts).
		SetRange(c.Start, c.Stop.AddDate(0, 0, 1)).
		SetTimeZone("Europe/Zurich") // hardcoded for now
	report, err := reporter.CalculateReport()
	c.Result = report
	return err
}

func (c *EmployeeReport) fetchPreviousPayslip(ctx context.Context) error {
	// TODO: verify timestamp
	firstDayOfLastMonth := c.Start.AddDate(0, -1, 0)
	payslip, err := c.OdooClient.FetchPayslipInMonth(ctx, c.Employee.ID, firstDayOfLastMonth)
	c.PreviousPayslip = payslip
	return err
}

func (c *EmployeeReport) fetchNextPayslip(ctx context.Context) error {
	thisMonth := c.Start
	payslip, err := c.OdooClient.FetchPayslipInMonth(ctx, c.Employee.ID, thisMonth)
	c.NextPayslip = payslip
	return err
}

func (c *EmployeeReport) ignoreNoContractFound(_ context.Context, err error) error {
	if strings.Contains(err.Error(), "no contract found that covers date") {
		return nil
	}
	return err
}
