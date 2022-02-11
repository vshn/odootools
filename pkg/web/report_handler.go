package web

import (
	"github.com/labstack/echo/v4"
	"github.com/vshn/odootools/pkg/web/employeereport"
	"github.com/vshn/odootools/pkg/web/overtimereport"
	"github.com/vshn/odootools/pkg/web/reportconfig"
)

// OvertimeReport GET /report/:id/:year(/:month)
func (s Server) OvertimeReport(e echo.Context) error {
	ctrl := overtimereport.NewReportController(s.newControllerContext(e))
	if err := ctrl.DisplayOvertimeReport(); err != nil {
		return s.ShowError(e, err)
	}
	return nil
}

// RequestReportForm GET /report
func (s Server) RequestReportForm(e echo.Context) error {
	return reportconfig.NewConfigController(s.newControllerContext(e)).ShowConfigurationForm()
}

// ProcessReportInput POST /report
func (s Server) ProcessReportInput(e echo.Context) error {
	ctrl := reportconfig.NewConfigController(s.newControllerContext(e))

	if err := ctrl.ProcessInput(); err != nil {
		return s.ShowError(e, err)
	}
	return nil
}

// EmployeeReport GET /report/employees/:year/:month
func (s Server) EmployeeReport(e echo.Context) error {
	ctrl := employeereport.NewEmployeeReportController(s.newControllerContext(e))
	if err := ctrl.DisplayEmployeeReport(); err != nil {
		return s.ShowError(e, err)
	}
	return nil
}
