package timesheet

import (
	"time"

	"github.com/vshn/odootools/pkg/odoo"
	"github.com/vshn/odootools/pkg/odoo/model"
)

type YearlyReport struct {
	MonthlyReports []BalanceReport
	Employee       model.Employee
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
	year        int
	payslips    model.PayslipList
	attendances model.AttendanceList
	leaves      odoo.List[model.Leave]
	employee    model.Employee
	contracts   model.ContractList
	clock       func() time.Time
}

func NewYearlyReporter(attendances model.AttendanceList, leaves odoo.List[model.Leave], employee model.Employee, contracts model.ContractList, payslips model.PayslipList) *YearlyReportBuilder {
	return &YearlyReportBuilder{
		payslips:    payslips,
		attendances: attendances,
		leaves:      leaves,
		employee:    employee,
		contracts:   contracts,
		clock:       time.Now,
	}
}

func (r *YearlyReportBuilder) CalculateYearlyReport() (YearlyReport, error) {
	reports := make([]BalanceReport, 0)
	now := r.clock()
	max := 12
	if r.year >= now.Year() {
		max = int(now.Month())
	}
	min := 1
	if startDate, found := r.getEarliestStartContractDate(); found && startDate.Year() == now.Year() && r.year == now.Year() {
		min = int(startDate.Month())
	}

	for _, month := range makeRange(min, max) {
		tz := DefaultTimeZone
		payslip := r.payslips.FilterInMonth(time.Date(r.year, time.Month(month), 5, 0, 0, 0, 0, time.UTC))
		if payslip != nil {
			tz = payslip.TimeZone.LocationOrDefault(tz)
		}
		start := time.Date(r.year, time.Month(month), 1, 0, 0, 0, 0, tz)
		end := start.AddDate(0, 1, 0)
		monthlyReportBuilder := NewReporter(r.attendances, r.leaves, r.employee, r.contracts)
		monthlyReportBuilder.clock = r.clock
		monthlyReport, err := monthlyReportBuilder.CalculateReport(start, end)
		if err != nil {
			return YearlyReport{}, err
		}
		balanceReportBuilder := NewBalanceReportBuilder(monthlyReport, r.payslips)
		balanceReport, err := balanceReportBuilder.CalculateBalanceReport()
		if err != nil {
			return YearlyReport{}, err
		}
		reports = append(reports, balanceReport)
	}
	yearlyReport := YearlyReport{
		MonthlyReports: reports,
		Year:           r.year,
		Employee:       r.employee,
	}
	summary := YearlySummary{}
	for _, month := range reports {
		summary.TotalOvertime += month.Report.Summary.TotalOvertime
		summary.TotalExcused += month.Report.Summary.TotalExcusedTime
		summary.TotalWorked += month.Report.Summary.TotalWorkedTime
		summary.TotalLeaves += month.Report.Summary.TotalLeave
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
	now := r.clock()
	start := now
	for _, contract := range r.contracts.Items {
		if contract.Start.Before(start) {
			start = contract.Start.Time
		}
	}
	return start, start != now
}

func (r *YearlyReportBuilder) SetYear(year int) *YearlyReportBuilder {
	r.year = year
	return r
}
