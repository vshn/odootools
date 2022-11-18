package overtimereport

import (
	"context"
	"net/http"

	pipeline "github.com/ccremer/go-command-pipeline"
	"github.com/vshn/odootools/pkg/odoo/model"
	"github.com/vshn/odootools/pkg/timesheet"
	"github.com/vshn/odootools/pkg/web/controller"
)

type YearlyReportController struct {
	ReportController
	ReportView *yearlyReportView
	Payslips   model.PayslipList
}

func NewYearlyReportController(controller controller.BaseController) *YearlyReportController {
	return &YearlyReportController{
		ReportController: ReportController{
			BaseController: controller,
		},
		ReportView: &yearlyReportView{},
	}
}

// DisplayYearlyOvertimeReport GET /report/:id/:year
func (c *YearlyReportController) DisplayYearlyOvertimeReport() error {
	root := pipeline.NewPipeline[context.Context]()
	root.WithSteps(
		root.NewStep("parse user input", c.parseInput),
		root.NewStep("fetch employee", c.fetchEmployeeByID),
		root.NewStep("fetch payslips", c.fetchPayslips),
		root.NewStep("fetch contracts", c.fetchContracts),
		root.NewStep("fetch attendances", c.fetchAttendances),
		root.NewStep("fetch leaves", c.fetchLeaves),
		root.NewStep("calculate monthly report", c.calculateYearlyReport),
	)
	err := root.RunWithContext(c.RequestContext)
	return err
}

func (c *YearlyReportController) calculateYearlyReport(_ context.Context) error {
	reporter := timesheet.NewYearlyReporter(c.Attendances, c.Leaves, c.Employee, c.Contracts, c.Payslips).
		SetYear(c.Input.Year)
	report, err := reporter.CalculateYearlyReport()
	if err != nil {
		return err
	}
	values := c.ReportView.GetValuesForYearlyReport(report)
	return c.Echo.Render(http.StatusOK, yearlyReportTemplateName, values)
}

func (c *YearlyReportController) fetchPayslips(ctx context.Context) error {
	start := c.Input.GetFirstDayOfYear().AddDate(0, 0, -1)
	end := start.AddDate(1, 0, 2)
	payslips, err := c.OdooClient.FetchPayslipBetween(ctx, c.Employee.ID, start, end)
	c.Payslips = payslips
	return err
}
