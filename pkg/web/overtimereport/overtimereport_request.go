package overtimereport

import (
	"html"
	"time"

	"github.com/labstack/echo/v4"
)

type ReportRequest struct {
	Year              int    `param:"year"`
	Month             int    `param:"month"`
	SearchUser        string `form:"username"`
	SearchUserEnabled bool
	EmployeeID        int `param:"employee"`
}

func (i *ReportRequest) FromRequest(e echo.Context) error {
	err := e.Bind(i)
	if err != nil {
		return err
	}

	i.SearchUserEnabled = e.FormValue("userscope") == "user-foreign-radio"
	i.SearchUser = html.EscapeString(i.SearchUser)

	if i.Year == 0 {
		i.Year = time.Now().Year()
	}
	if i.Month == 0 {
		i.Month = int(time.Now().Month())
	}
	return nil
}

func (i ReportRequest) getFirstDayOfMonth() time.Time {
	firstDay := time.Date(i.Year, time.Month(i.Month), 1, 0, 0, 0, 0, time.UTC)
	// Let's get attendances within a month with - 1 day to respect localized dates and filter them later.
	begin := firstDay.AddDate(0, 0, -1)
	return begin
}

func (i ReportRequest) getLastDayOfMonth() time.Time {
	firstDay := time.Date(i.Year, time.Month(i.Month), 1, 0, 0, 0, 0, time.UTC)
	// Let's get attendances within a month with + 1 day to respect localized dates and filter them later.
	end := firstDay.AddDate(0, 1, 0)
	return end
}
