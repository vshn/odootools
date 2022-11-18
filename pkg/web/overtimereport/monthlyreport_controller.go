package overtimereport

import (
	"context"
	"net/http"
	"time"

	pipeline "github.com/ccremer/go-command-pipeline"
	"github.com/vshn/odootools/pkg/odoo/model"
	"github.com/vshn/odootools/pkg/timesheet"
	"github.com/vshn/odootools/pkg/web/controller"
)

type MonthlyReportController struct {
	ReportController
	ReportView    *monthlyReportView
	User          *model.User
	Payslips      model.PayslipList
	BalanceReport timesheet.BalanceReport
}

func NewMonthlyReportController(ctx controller.BaseController) *MonthlyReportController {
	return &MonthlyReportController{
		ReportController: ReportController{
			BaseController: ctx,
		},
		ReportView: &monthlyReportView{},
	}
}

// DisplayMonthlyOvertimeReport GET /report/:id/:year/:month
func (c *MonthlyReportController) DisplayMonthlyOvertimeReport() error {
	root := pipeline.NewPipeline[context.Context]()
	root.WithSteps(
		root.NewStep("parse user input", c.parseInput),
		root.NewStep("fetch employee", c.fetchEmployeeByID),
		root.NewStep("fetch data", c.FetchReportData),
		root.NewStep("calculate monthly report", c.CalculateMonthlyReport),
		root.NewStep("render report", c.renderReport),
	)
	err := root.RunWithContext(c.RequestContext)
	return err
}

func (c *MonthlyReportController) FetchReportData(ctx context.Context) error {
	root := pipeline.NewPipeline[context.Context]()
	root.WithSteps(
		root.NewStep("fetch payslips", c.fetchPayslips),
		root.NewStep("fetch user settings", c.fetchUser),
		root.NewStep("fetch contracts", c.fetchContracts),
		root.NewStep("fetch attendances", c.fetchAttendances),
		root.NewStep("fetch leaves", c.fetchLeaves),
	)
	err := root.RunWithContext(ctx)
	return err
}

func (c *MonthlyReportController) CalculateMonthlyReport(_ context.Context) error {
	tz := c.getTimeZone()
	start := time.Date(c.Input.Year, time.Month(c.Input.Month), 1, 0, 0, 0, 0, tz)
	end := start.AddDate(0, 1, 0)
	reporter := timesheet.NewReporter(c.Attendances.AddCurrentTimeAsSignOut(tz), c.Leaves, c.Employee, c.Contracts)
	report, err := reporter.CalculateReport(start, end)
	c.BalanceReport.Report = report // needed so that error handler can retrieve employee name
	if err != nil {
		return err
	}
	balanceReporter := timesheet.NewBalanceReportBuilder(report, c.Payslips)
	balanceReport, err := balanceReporter.CalculateBalanceReport()
	c.BalanceReport = balanceReport
	return err
}

func (c *MonthlyReportController) renderReport(_ context.Context) error {
	values := c.ReportView.GetValuesForMonthlyReport(c.BalanceReport)
	return c.Echo.Render(http.StatusOK, monthlyReportTemplateName, values)
}

func (c *MonthlyReportController) getTimeZone() *time.Location {
	nextPayslip := c.GetNextPayslip()
	if nextPayslip != nil && !nextPayslip.TimeZone.IsEmpty() {
		// timezone from payslip has precedence.
		return nextPayslip.TimeZone.Location
	}
	if c.User != nil && time.Now().Month() == time.Month(c.Input.Month) {
		// get the timezone from user preferences only if we create a report for the current month.
		// for months long in the past we don't want to calculate based on user's current preferences.
		return c.User.TimeZone.LocationOrDefault(timesheet.DefaultTimeZone)
	}
	// last resort to default TZ.
	return timesheet.DefaultTimeZone
}

func (c *MonthlyReportController) fetchPayslips(ctx context.Context) error {
	lastMonth := c.Input.GetFirstDayOfMonth().AddDate(0, -1, -1)
	currentMonth := lastMonth.AddDate(0, 2, 1)
	payslips, err := c.OdooClient.FetchPayslipBetween(ctx, c.Employee.ID, lastMonth, currentMonth)
	c.Payslips = payslips
	return err
}

func (c *MonthlyReportController) fetchUser(ctx context.Context) error {
	user, err := c.OdooClient.FetchUserByID(ctx, c.OdooSession.UID)
	c.User = user
	return err
}

func (c *MonthlyReportController) GetNextPayslip() *model.Payslip {
	return c.Payslips.FilterInMonth(time.Date(c.Input.Year, time.Month(c.Input.Month), 2, 0, 0, 0, 0, time.UTC))
}

func (c *MonthlyReportController) GetPreviousPayslip() *model.Payslip {
	return c.Payslips.FilterInMonth(time.Date(c.Input.Year, time.Month(c.Input.Month-1), 2, 0, 0, 0, 0, time.UTC))
}
