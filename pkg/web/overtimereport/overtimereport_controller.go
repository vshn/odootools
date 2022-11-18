package overtimereport

import (
	"context"
	"fmt"

	"github.com/vshn/odootools/pkg/odoo"
	"github.com/vshn/odootools/pkg/odoo/model"
	"github.com/vshn/odootools/pkg/web/controller"
	"github.com/vshn/odootools/pkg/web/reportconfig"
)

type ReportController struct {
	controller.BaseController
	Input       reportconfig.ReportRequest
	Employee    model.Employee
	Contracts   model.ContractList
	Attendances model.AttendanceList
	Leaves      odoo.List[model.Leave]
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
		c.Employee = *c.SessionData.Employee
		return nil
	}

	employee, err := c.OdooClient.FetchEmployeeByID(ctx, employeeID)
	if employee == nil {
		return fmt.Errorf("no employee found with given ID: %d", employeeID)
	}
	c.Employee = *employee
	return err
}

func (c *ReportController) fetchContracts(ctx context.Context) error {
	contracts, err := c.OdooClient.FetchAllContractsOfEmployee(ctx, c.Employee.ID)
	c.Contracts = contracts
	return err
}

func (c *ReportController) fetchAttendances(ctx context.Context) error {
	begin, end := c.Input.GetDateRange()
	// get more entries to cover all timezones, filter out later.
	begin = begin.AddDate(0, 0, -1)
	end = end.AddDate(0, 0, 1)
	attendances, err := c.OdooClient.FetchAttendancesBetweenDates(ctx, c.Employee.ID, begin, end)
	c.Attendances = attendances
	return err
}

func (c *ReportController) fetchLeaves(ctx context.Context) error {
	begin, end := c.Input.GetDateRange()
	// get more entries to cover all timezones, filter out later.
	begin = begin.AddDate(0, 0, -1)
	end = end.AddDate(0, 0, 1)
	leaves, err := c.OdooClient.FetchLeavesBetweenDates(ctx, c.Employee.ID, begin, end)
	c.Leaves = leaves
	return err
}
