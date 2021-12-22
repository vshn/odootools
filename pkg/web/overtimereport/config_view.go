package overtimereport

import "github.com/vshn/odootools/pkg/web/controller"

const configViewTemplate = "createreport"

type ConfigView struct {
}

func (v *ConfigView) GetConfigurationValues() controller.Values {
	return controller.Values{
		"Nav": controller.Values{
			"LoggedIn":   true,
			"ActiveView": configViewTemplate,
		},
	}
}
