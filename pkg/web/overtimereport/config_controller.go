package overtimereport

import (
	"net/http"

	"github.com/vshn/odootools/pkg/web/controller"
)

type ConfigController struct {
	controller.Context
	view *ConfigView
}

func NewConfigController(ctx *controller.Context) *ConfigController {
	return &ConfigController{
		Context: *ctx,
		view:    &ConfigView{},
	}
}

func (c *ConfigController) ShowConfigurationForm() error {
	c.view.roles = c.SessionData.Roles
	return c.Echo.Render(http.StatusOK, configViewTemplate, c.view.GetConfigurationValues())
}
