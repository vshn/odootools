package employeereport

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	pipeline "github.com/ccremer/go-command-pipeline"
	"github.com/hashicorp/go-multierror"
	"github.com/vshn/odootools/pkg/odoo"
	"github.com/vshn/odootools/pkg/odoo/model"
	"github.com/vshn/odootools/pkg/web/controller"
	"github.com/vshn/odootools/pkg/web/overtimereport"
	"github.com/vshn/odootools/pkg/web/reportconfig"
)

type ReportController struct {
	controller.BaseController
	Input     reportconfig.ReportRequest
	employees odoo.List[model.Employee]
	User      *model.User
	reports   []*EmployeeReport
	view      *reportView
}

type EmployeeReport struct {
	MonthlyReportController *overtimereport.MonthlyReportController
}

func NewEmployeeReportController(ctx *controller.BaseController) *ReportController {
	return &ReportController{
		BaseController: *ctx,
		view:           &reportView{},
	}
}

// DisplayEmployeeReport GET /report/employees/:year/:month
func (c *ReportController) DisplayEmployeeReport() error {
	root := pipeline.NewPipeline[context.Context]()
	root.WithOptions(pipeline.Options{DisableErrorWrapping: true}).
		WithSteps(
			root.NewStep("parse user input", c.parseInput),
			root.NewStep("fetch employees", c.fetchEmployees),
			pipeline.NewWorkerPoolStep("generate reports for each employee", 4, c.createPipelinesForEachEmployee, c.collectReports),
			root.NewStep("render report", c.renderReport),
		)
	err := root.RunWithContext(c.RequestContext)
	return err
}

func (c *ReportController) createPipelinesForEachEmployee(ctx context.Context, pipelines chan *pipeline.Pipeline[context.Context]) {
	defer close(pipelines)
	c.reports = make([]*EmployeeReport, c.employees.Len())
	for i, employee := range c.employees.Items {
		select {
		case <-ctx.Done():
			return
		default:
			ctrl := overtimereport.NewMonthlyReportController(c.BaseController)
			ctrl.Employee = employee
			ctrl.Input.Month = c.Input.Month
			ctrl.Input.Year = c.Input.Year
			report := &EmployeeReport{
				MonthlyReportController: ctrl,
			}
			c.reports[i] = report
			pipe := report.createPipeline()
			pipelines <- pipe
		}
	}
}

func (c *EmployeeReport) createPipeline() *pipeline.Pipeline[context.Context] {
	p := pipeline.NewPipeline[context.Context]()
	p.AddStep(p.WithNestedSteps(fmt.Sprintf("report for %q", c.MonthlyReportController.Employee.Name), nil,
		p.NewStep("fetch data", c.MonthlyReportController.FetchReportData),
		p.NewStep("calculate monthly report", c.MonthlyReportController.CalculateMonthlyReport).WithErrorHandler(c.ignoreNoContractFound),
	))
	return p
}

func (c *ReportController) collectReports(_ context.Context, results map[uint64]error) error {
	var combined error
	for _, err := range results {
		if err != nil {
			combined = multierror.Append(combined, err)
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
	failedReports := make([]model.Employee, 0)
	for _, report := range c.reports {
		if report.MonthlyReportController.BalanceReport.Report.DailySummaries != nil {
			successfulReports = append(successfulReports, report)
		} else {
			failedReports = append(failedReports, report.MonthlyReportController.BalanceReport.Report.Employee)
		}
	}
	c.view.year, c.view.month = c.Input.Year, c.Input.Month
	return c.Echo.Render(http.StatusOK, employeeReportTemplateName, c.view.GetValuesForReports(successfulReports, failedReports))
}

func (c *EmployeeReport) ignoreNoContractFound(_ context.Context, err error) error {
	var noContractErr *model.NoContractCoversDateErr
	if errors.As(err, &noContractErr) {
		return nil
	}
	return err
}
