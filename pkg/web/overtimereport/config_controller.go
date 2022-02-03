package overtimereport

import (
	"net/http"

	pipeline "github.com/ccremer/go-command-pipeline"
	"github.com/vshn/odootools/pkg/odoo/model"
	"github.com/vshn/odootools/pkg/web/controller"
)

type ConfigController struct {
	controller.Context
	view  *ConfigView
	group *model.Group
}

func NewConfigController(ctx *controller.Context) *ConfigController {
	return &ConfigController{
		Context: *ctx,
		view:    &ConfigView{},
	}
}

func (c *ConfigController) ShowConfigurationForm() error {
	p := pipeline.NewPipelineWithContext(c).
		WithSteps(
			pipeline.NewStepFromFunc("fetch manager group", c.fetchGroup),
			pipeline.NewStepFromFunc("check permissions", c.checkGroupMembership),
		)
	result := p.Run()
	if result.IsFailed() {
		return result.Err
	}
	return c.Echo.Render(http.StatusOK, configViewTemplate, c.view.GetConfigurationValues())
}

func (c *ConfigController) fetchGroup(_ pipeline.Context) error {
	group, err := c.OdooClient.FetchGroupByName("Human Resources", "Manager")
	c.group = group
	return err
}

func (c *ConfigController) checkGroupMembership(_ pipeline.Context) error {
	if c.group != nil {
		for _, userID := range c.group.UserIDs {
			if c.OwnUserID == userID {
				c.view.roles = append(c.view.roles, "HRManager")
			}
		}
	}
	return nil
}
