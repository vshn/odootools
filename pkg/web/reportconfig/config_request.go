package reportconfig

import (
	"html"
	"time"

	"github.com/labstack/echo/v4"
)

// ReportRequest contains all relevant input data for generating reports.
type ReportRequest struct {
	Year                  int    `param:"year" form:"year"`
	Month                 int    `param:"month" form:"month"`
	SearchUser            string `form:"username"`
	SearchUserEnabled     bool
	EmployeeReportEnabled bool
	EmployeeID            int `param:"employee"`
}

// FromRequest parses the properties based on the given request echo.Context.
func (i *ReportRequest) FromRequest(e echo.Context) error {
	if err := e.Bind(i); err != nil {
		return err
	}
	i.SearchUserEnabled = e.FormValue("userscope") == "user-foreign-radio"
	i.EmployeeReportEnabled = e.FormValue("employeeReport") == "true"
	i.SearchUser = html.EscapeString(i.SearchUser)

	if i.Month == 0 && i.Year == 0 {
		// this is kinda invalid input. Maybe created via curl or so.
		i.Month = int(time.Now().Month())
	}
	if i.Year == 0 {
		// The HTML view doesn't leave this empty. Maybe the request is foreign, so we give a sane default.
		i.Year = time.Now().Year()
	}
	if e.FormValue("yearlyReport") == "true" {
		// this way we configure the pipeline to do a yearly report.
		i.Month = 0
	}
	return nil
}

// GetFirstDayOfMonth returns the first day of the ReportRequest.Month in time.UTC at midnight.
func (i ReportRequest) GetFirstDayOfMonth() time.Time {
	month := i.Month
	if month <= 0 {
		month = 1
	}
	firstDay := time.Date(i.Year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	return firstDay
}

// GetLastDayOfMonth returns the last day of the ReportRequest.Month in time.UTC at midnight.
func (i ReportRequest) GetLastDayOfMonth() time.Time {
	return i.GetFirstDayOfNextMonth().AddDate(0, 0, -1)
}

// GetLastDayFromPreviousMonth returns GetFirstDayOfMonth subtracted by 1 day.
func (i ReportRequest) GetLastDayFromPreviousMonth() time.Time {
	return i.GetFirstDayOfMonth().AddDate(0, 0, -1)
}

// GetFirstDayOfNextMonth returns the last day of the ReportRequest.Month in time.UTC.
func (i ReportRequest) GetFirstDayOfNextMonth() time.Time {
	return i.GetFirstDayOfMonth().AddDate(0, 1, 0)
}
