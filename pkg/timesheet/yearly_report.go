package timesheet

import (
	"time"

	"github.com/vshn/odootools/pkg/odoo/model"
)

type YearlyReport struct {
	MonthlyReports []MonthlyReport
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

func (r *ReportBuilder) CalculateYearlyReport() YearlyReport {
	reports := make([]MonthlyReport, 0)

	max := 12
	if r.year >= now().Year() {
		max = int(now().Month())
	}
	min := 1
	if startDate, found := r.getEarliestStartContractDate(); found && startDate.Year() == now().Year() && r.year == now().Year() {
		min = int(startDate.Month())
	}

	for _, month := range makeRange(min, max) {
		r.month = month
		monthlyReport := r.CalculateMonthlyReport()
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
	return yearlyReport
}

func makeRange(min, max int) []int {
	a := make([]int, max-min+1)
	for i := range a {
		a[i] = min + i
	}
	return a
}

func (r *ReportBuilder) getEarliestStartContractDate() (time.Time, bool) {
	n := now()
	start := n
	for _, contract := range r.contracts.Items {
		if contract.Start.ToTime().Before(start) {
			start = contract.Start.ToTime()
		}
	}
	return start, start != n
}
