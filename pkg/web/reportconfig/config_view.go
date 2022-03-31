package reportconfig

import (
	"github.com/vshn/odootools/pkg/timesheet"
	"github.com/vshn/odootools/pkg/web/controller"
)

const configViewTemplate = "createreport"

type ConfigView struct {
	controller.BaseView
	roles      []string
	isSignedIn bool
}

func (v *ConfigView) GetConfigurationValues(report timesheet.Report) controller.Values {
	formatted := make([]controller.Values, 0)
	for _, summary := range report.DailySummaries {
		if summary.IsWeekend() && summary.CalculateWorkingTime() == 0 {
			continue
		}
		formatted = append(formatted, v.FormatDailySummary(summary))
	}
	summary := controller.Values{
		"TotalOvertime": v.FormatDurationInHours(report.Summary.TotalOvertime),
		"TotalWorked":   v.FormatDurationInHours(report.Summary.TotalWorkedTime),
	}
	vals := controller.Values{
		"Nav": controller.Values{
			"LoggedIn":   true,
			"ActiveView": configViewTemplate,
		},
		"Roles":       controller.Values{},
		"IsSignedIn":  v.isSignedIn,
		"Attendances": formatted,
		"Summary":     summary,
	}
	if len(v.roles) > 0 {
		for _, role := range v.roles {
			vals["Roles"].(controller.Values)[role] = true
		}
	}
	return vals
}
