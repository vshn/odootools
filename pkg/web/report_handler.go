package web

import (
	"github.com/labstack/echo/v4"
	"github.com/vshn/odootools/pkg/web/employeereport"
	"github.com/vshn/odootools/pkg/web/overtimereport"
	"github.com/vshn/odootools/pkg/web/reportconfig"
)

// MonthlyOvertimeReport GET /report/:id/:year/:month
func (s *Server) MonthlyOvertimeReport(e echo.Context) error {
	ctrl := overtimereport.NewMonthlyReportController(*s.newControllerContext(e))
	if err := ctrl.DisplayMonthlyOvertimeReport(); err != nil {
		return s.ShowError(e, err)
	}
	return nil
}

// YearlyOvertimeReport GET /report/:id/:year
func (s *Server) YearlyOvertimeReport(e echo.Context) error {
	ctrl := overtimereport.NewYearlyReportController(*s.newControllerContext(e))
	if err := ctrl.DisplayYearlyOvertimeReport(); err != nil {
		return s.ShowError(e, err)
	}
	return nil
}

// RequestReportForm GET /report
func (s *Server) RequestReportForm(e echo.Context) error {
	return reportconfig.NewConfigController(s.newControllerContext(e)).ShowConfigurationFormAndWeeklyReport()
}

// ProcessReportInput POST /report
func (s *Server) ProcessReportInput(e echo.Context) error {
	ctrl := reportconfig.NewConfigController(s.newControllerContext(e))

	if err := ctrl.ProcessInput(); err != nil {
		return s.ShowError(e, err)
	}
	return nil
}

// EmployeeReport GET /report/employees/:year/:month
func (s *Server) EmployeeReport(e echo.Context) error {
	ctrl := employeereport.NewEmployeeReportController(s.newControllerContext(e))
	if err := ctrl.DisplayEmployeeReport(); err != nil {
		return s.ShowError(e, err)
	}
	return nil
}

// EmployeeReportUpdate POST /report/employee/:employee/:year/:month.
// Updates the payslip with the overtime value of the given month.
func (s *Server) EmployeeReportUpdate(e echo.Context) error {
	ctrl := employeereport.NewUpdatePayslipController(s.newControllerContext(e))
	return ctrl.UpdatePayslipOfEmployee()
}
