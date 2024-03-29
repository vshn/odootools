package employeereport

import (
	"context"
	"fmt"
	"net/http"

	pipeline "github.com/ccremer/go-command-pipeline"
	"github.com/vshn/odootools/pkg/odoo/model"
	"github.com/vshn/odootools/pkg/web/controller"
)

type UpdatePayslipController struct {
	controller.BaseController
	Input       UpdateRequest
	NextPayslip *model.Payslip
	Employee    *model.Employee
}

func NewUpdatePayslipController(ctx *controller.BaseController) *UpdatePayslipController {
	return &UpdatePayslipController{
		BaseController: *ctx,
	}
}

// UpdatePayslipOfEmployee POST /report/employee/:employee/:year/:month
func (c *UpdatePayslipController) UpdatePayslipOfEmployee() error {
	root := pipeline.NewPipeline[context.Context]()
	root.WithSteps(
		root.NewStep("parse user input", c.parseInput).WithErrorHandler(c.serverError(http.StatusBadRequest)),
		root.NewStep("fetch employee", c.fetchEmployeeByID).WithErrorHandler(c.serverError(http.StatusBadRequest)),
		root.NewStep("fetch current month's payslip", c.fetchNextPayslip).WithErrorHandler(c.serverError(http.StatusBadRequest)),
		root.NewStep("save payslip", c.savePayslip).WithErrorHandler(c.serverError(http.StatusInternalServerError)),
	)
	err := root.RunWithContext(c.RequestContext)
	return err
}

func (c *UpdatePayslipController) parseInput(_ context.Context) error {
	input := UpdateRequest{}
	err := input.FromRequest(c.Echo)
	c.Input = input
	return err
}

func (c *UpdatePayslipController) fetchEmployeeByID(ctx context.Context) error {
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

func (c *UpdatePayslipController) fetchNextPayslip(ctx context.Context) error {
	thisMonth := c.Input.BaseReportRequest.GetFirstDayOfMonth()
	payslip, err := c.OdooClient.FetchPayslipInMonth(ctx, c.Input.EmployeeID, thisMonth)
	if err != nil {
		return err
	}
	if payslip == nil {
		return fmt.Errorf("attempting to update a payslip that doesn't exist in %s %d for employee %q",
			thisMonth.Month().String(), thisMonth.Year(), c.Employee.Name)
	}
	c.NextPayslip = payslip
	return nil
}

func (c *UpdatePayslipController) savePayslip(ctx context.Context) error {
	payslip := c.NextPayslip
	payslip.XOvertime = c.Input.Overtime
	err := c.OdooClient.UpdatePayslip(ctx, payslip)
	if err != nil {
		return err
	}
	return c.Echo.JSON(http.StatusOK, UpdateResponse{
		Overtime: c.Input.Overtime,
		Employee: c.Employee,
	})
}

func (c *UpdatePayslipController) serverError(httpStatusError int) func(_ context.Context, err error) error {
	return func(_ context.Context, err error) error {
		if err != nil {
			jsonErr := c.Echo.JSON(httpStatusError, UpdateResponse{ErrorMessage: err.Error()})
			if jsonErr != nil {
				return jsonErr
			}
		}
		return err
	}
}
