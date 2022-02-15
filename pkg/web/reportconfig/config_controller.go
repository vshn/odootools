package reportconfig

import (
	"fmt"
	"net/http"

	pipeline "github.com/ccremer/go-command-pipeline"
	"github.com/vshn/odootools/pkg/odoo/model"
	"github.com/vshn/odootools/pkg/web/controller"
)

type ConfigController struct {
	controller.BaseController
	Input    ReportRequest
	view     *ConfigView
	Employee *model.Employee
}

func NewConfigController(ctx *controller.BaseController) *ConfigController {
	return &ConfigController{
		BaseController: *ctx,
		view:           &ConfigView{},
	}
}

func (c *ConfigController) ShowConfigurationForm() error {
	c.view.roles = c.SessionData.Roles
	return c.Echo.Render(http.StatusOK, configViewTemplate, c.view.GetConfigurationValues())
}

func (c *ConfigController) ProcessInput() error {
	root := pipeline.NewPipelineWithContext(c).
		WithSteps(
			pipeline.NewStepFromFunc("parse user input", c.parseInput),
			pipeline.NewStepFromFunc("search employee", c.searchEmployee),
			pipeline.NewStepFromFunc("redirect to report", c.redirectToReportView),
		)
	result := root.Run()
	return result.Err
}

func (c *ConfigController) parseInput(_ pipeline.Context) error {
	input := ReportRequest{}
	err := input.FromRequest(c.Echo)
	c.Input = input
	return err
}

func (c *ConfigController) searchEmployee(_ pipeline.Context) error {
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

func (c *ConfigController) redirectToReportView(_ pipeline.Context) error {
	if c.Input.EmployeeReportEnabled {
		return c.Echo.Redirect(http.StatusFound, fmt.Sprintf("/report/employees/%d/%02d", c.Input.Year, c.Input.Month))
	}
	if c.Input.Month == 0 {
		return c.Echo.Redirect(http.StatusFound, fmt.Sprintf("/report/%d/%d", c.Employee.ID, c.Input.Year))
	}
	return c.Echo.Redirect(http.StatusFound, fmt.Sprintf("/report/%d/%d/%02d", c.Employee.ID, c.Input.Year, c.Input.Month))
}
