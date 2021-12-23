package overtimereport

import (
	"html"
	"time"

	"github.com/labstack/echo/v4"
)

type ReportRequest struct {
	Year              int    `param:"year" form:"year"`
	Month             int    `param:"month" form:"month"`
	SearchUser        string `form:"username"`
	SearchUserEnabled bool
	EmployeeID        int `param:"employee"`
}

func (i *ReportRequest) FromRequest(e echo.Context) error {
	if err := e.Bind(i); err != nil {
		return err
	}
	i.SearchUserEnabled = e.FormValue("userscope") == "user-foreign-radio"
	i.SearchUser = html.EscapeString(i.SearchUser)

	if i.Month == 0 && i.Year == 0 {
		i.Month = int(time.Now().Month())
	}
	if i.Year == 0 {
		i.Year = time.Now().Year()
	}
	return nil
}

func (i ReportRequest) getFirstDay() time.Time {
	month := i.Month
	if month == 0 {
		month = 1
	}
	firstDay := time.Date(i.Year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	// Let's get attendances within a month with - 1 day to respect localized dates and filter them later.
	begin := firstDay.AddDate(0, 0, -1)
	return begin
}

func (i ReportRequest) getLastDay() time.Time {
	month := i.Month
	if month == 0 {
		month = 12
	}
	firstDay := time.Date(i.Year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	// Let's get attendances within a month with + 1 day to respect localized dates and filter them later.
	end := firstDay.AddDate(0, 1, 0)
	return end
}
