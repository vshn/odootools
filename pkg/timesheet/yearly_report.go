package timesheet

import (
	"time"

	"github.com/vshn/odootools/pkg/odoo"
	"github.com/vshn/odootools/pkg/odoo/model"
)

type YearlyReport struct {
	MonthlyReports []Report
	Employee       *model.Employee
	Year           int
	Summary        YearlySummary
}

type YearlySummary struct {
	TotalOvertime time.Duration
	TotalExcused  time.Duration
	TotalWorked   time.Duration
	TotalLeaves   float64
}

type YearlyReportBuilder struct {
	ReportBuilder
	year int
}

func NewYearlyReporter(attendances odoo.List[model.Attendance], leaves odoo.List[model.Leave], employee *model.Employee, contracts model.ContractList) *YearlyReportBuilder {
	return &YearlyReportBuilder{
		ReportBuilder: ReportBuilder{
			attendances: attendances,
			leaves:      leaves,
			employee:    employee,
			contracts:   contracts,
			timezone:    time.Local,
		},
	}
}

func (r *YearlyReportBuilder) CalculateYearlyReport() (YearlyReport, error) {
	reports := make([]Report, 0)

	max := 12
	if r.year >= now().Year() {
		max = int(now().Month())
	}
	min := 1
	if startDate, found := r.getEarliestStartContractDate(); found && startDate.Year() == now().Year() && r.year == now().Year() {
		min = int(startDate.Month())
	}

	for _, month := range makeRange(min, max) {
		start := time.Date(r.year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
		end := start.AddDate(0, 1, 0)
		r.SetRange(start, end)
		monthlyReport, err := r.CalculateReport()
		if err != nil {
			return YearlyReport{}, err
		}
		reports = append(reports, monthlyReport)
	}
	yearlyReport := YearlyReport{
		MonthlyReports: reports,
		Year:           r.year,
		Employee:       r.employee,
	}
	summary := YearlySummary{}
	for _, month := range reports {
		summary.TotalOvertime += month.Summary.TotalOvertime
		summary.TotalExcused += month.Summary.TotalExcusedTime
		summary.TotalWorked += month.Summary.TotalWorkedTime
		summary.TotalLeaves += month.Summary.TotalLeave
	}
	yearlyReport.Summary = summary
	return yearlyReport, nil
}

func makeRange(min, max int) []int {
	a := make([]int, max-min+1)
	for i := range a {
		a[i] = min + i
	}
	return a
}

func (r *YearlyReportBuilder) getEarliestStartContractDate() (time.Time, bool) {
	n := now()
	start := n
	for _, contract := range r.contracts.Items {
		if contract.Start.ToTime().Before(start) {
			start = contract.Start.ToTime()
		}
	}
	return start, start != n
}

func (r *YearlyReportBuilder) SetYear(year int) *YearlyReportBuilder {
	r.year = year
	return r
}
