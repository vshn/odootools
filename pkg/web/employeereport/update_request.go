package employeereport

import (
	"fmt"
	"html"

	"github.com/labstack/echo/v4"
	"github.com/vshn/odootools/pkg/web/reportconfig"
)

type UpdateRequest struct {
	reportconfig.BaseReportRequest
	Overtime   string `form:"overtime"`
	EmployeeID int    `param:"employee"`
}

type UpdateResponse struct {
	ErrorMessage string `json:"errorMessage"`
}

// FromRequest parses the properties based on the given request echo.Context.
func (i *UpdateRequest) FromRequest(e echo.Context) error {
	if err := i.BaseReportRequest.FromRequest(e); err != nil {
		return err
	}
	if err := e.Bind(i); err != nil {
		return err
	}
	if i.Overtime == "" {
		return fmt.Errorf("overtime cannot be empty")
	} else {
		i.Overtime = html.EscapeString(i.Overtime)
	}
	return nil
}
