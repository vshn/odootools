package overtimereport

import "github.com/vshn/odootools/pkg/web/controller"

const configViewTemplate = "createreport"

type ConfigView struct {
	roles []string
}

func (v *ConfigView) GetConfigurationValues() controller.Values {
	vals := controller.Values{
		"Nav": controller.Values{
			"LoggedIn":   true,
			"ActiveView": configViewTemplate,
		},
		"Roles": controller.Values{},
	}
	if len(v.roles) > 0 {
		for _, role := range v.roles {
			vals["Roles"].(controller.Values)[role] = true
		}
	}
	return vals
}
