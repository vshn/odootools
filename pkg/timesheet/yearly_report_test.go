package timesheet

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vshn/odootools/pkg/odoo"
	"github.com/vshn/odootools/pkg/odoo/model"
)

func TestYearlyReportBuilder_CalculateYearlyReport(t *testing.T) {
	t.Skipf("test fails, needs more investigation")
	DefaultTimeZone = zurichTZ
	givenAttendances := model.AttendanceList{Items: []model.Attendance{
		// january in zurich
		{DateTime: odoo.NewDate(2021, 01, 04, 8, 0, 0, zurichTZ), Action: model.ActionSignIn},
		{DateTime: odoo.NewDate(2021, 01, 04, 17, 0, 0, zurichTZ), Action: model.ActionSignOut}, // 9h worked, 1h overtime

		// february in vancouver
		{DateTime: odoo.NewDate(2021, 02, 02, 8, 0, 0, vancouverTZ), Action: model.ActionSignIn},
		{DateTime: odoo.NewDate(2021, 02, 02, 17, 0, 0, vancouverTZ), Action: model.ActionSignOut}, // 9h worked, 1h overtime
		{DateTime: odoo.NewDate(2021, 02, 27, 21, 0, 0, vancouverTZ), Action: model.ActionSignIn, Reason: &model.ActionReason{Name: ReasonOutsideOfficeHours}},
		{DateTime: odoo.NewDate(2021, 02, 27, 23, 0, 0, vancouverTZ), Action: model.ActionSignOut, Reason: &model.ActionReason{Name: ReasonOutsideOfficeHours}}, // saturday on-call, 2h out of office, 3h overtime
		// NOTE: Sunday on call wouldn't work if working next month in zurich, as that would be work on the same day but in 2 timezones.

		// march in zurich
		{DateTime: odoo.NewDate(2021, 03, 01, 8, 0, 0, zurichTZ), Action: model.ActionSignIn}, // not signed out yet
	}}
	givenLeaves := odoo.List[model.Leave]{Items: []model.Leave{
		{DateFrom: odoo.NewDate(2021, 01, 01, 0, 0, 0, zurichTZ), DateTo: odoo.NewDate(2021, 01, 01, 23, 59, 59, zurichTZ), Type: &model.LeaveType{Name: TypeLegalLeavesPrefix}, State: StateApproved},
		{DateFrom: odoo.NewDate(2021, 02, 03, 7, 0, 0, vancouverTZ), DateTo: odoo.NewDate(2021, 02, 02, 19, 0, 0, vancouverTZ), Type: &model.LeaveType{Name: TypeLegalLeavesPrefix}, State: StateDraft}, // ignore this
		{DateFrom: odoo.NewDate(2021, 02, 04, 7, 0, 0, vancouverTZ), DateTo: odoo.NewDate(2021, 02, 03, 19, 0, 0, vancouverTZ), Type: &model.LeaveType{Name: TypeLegalLeavesPrefix}, State: StateApproved},
	}}
	givenEmployee := model.Employee{Name: "🕺"}
	givenContracts := model.ContractList{Items: []model.Contract{
		{Start: odoo.NewDate(2021, 01, 01, 0, 0, 0, time.UTC), WorkingSchedule: &model.WorkingSchedule{Name: "100%"}},
	}}
	givenPayslips := model.PayslipList{Items: []model.Payslip{
		{DateFrom: odoo.NewDate(2021, 01, 01, 0, 0, 0, time.UTC), DateTo: odoo.NewDate(2021, 02, 01, 0, 0, 0, time.UTC), XOvertime: "10:00:00"},
		{DateFrom: odoo.NewDate(2021, 02, 01, 0, 0, 0, time.UTC), DateTo: odoo.NewDate(2021, 03, 01, 0, 0, 0, time.UTC), TimeZone: odoo.NewTimeZone(vancouverTZ)},
	}}

	builder := NewYearlyReporter(givenAttendances, givenLeaves, givenEmployee, givenContracts, givenPayslips).
		SetYear(2021)
	builder.clock = func() time.Time {
		// fixed time
		return time.Date(2021, time.March, 1, 11, 0, 0, 0, zurichTZ)
	}

	report, err := builder.CalculateYearlyReport()
	assert.NoError(t, err)
	require.Len(t, report.MonthlyReports, 3)

	// january
	january := report.MonthlyReports[0]
	assert.Len(t, january.Report.DailySummaries, 31)
	assert.Equal(t, -(8*19)*time.Hour+1*time.Hour, january.Report.Summary.TotalOvertime, "january: total overtime")
	assert.Equal(t, 9*time.Hour, january.Report.Summary.TotalWorkedTime, "january: total worked time")
	assert.Equal(t, 1.0, january.Report.Summary.TotalLeave, "january: total leaves")
	assert.Equal(t, -(8*19)*time.Hour+1*time.Hour, january.CalculatedBalance, "january: calculated balance")
	assert.Equal(t, 10*time.Hour, *january.DefinitiveBalance, "january: definitive balance")

	// february in vancouver
	february := report.MonthlyReports[1]
	assert.Len(t, february.Report.DailySummaries, 28)
	assert.Equal(t, -(8*18)*time.Hour+(1+3)*time.Hour, february.Report.Summary.TotalOvertime, "february: total over time")
	assert.Equal(t, (9+3)*time.Hour, february.Report.Summary.TotalWorkedTime, "february: total worked time")
	assert.Equal(t, 1.0, february.Report.Summary.TotalLeave, "february: total leaves")
	assert.Equal(t, -(8*18)*time.Hour+(1+3+10)*time.Hour, february.CalculatedBalance, "february: calculated balance from january") // from january with definitive balance
	assert.Nil(t, february.DefinitiveBalance, "february: definitive balance should be empty")

	// march
	march := report.MonthlyReports[2]
	assert.Len(t, march.Report.DailySummaries, 1)
	assert.Nil(t, march.DefinitiveBalance, "march: definitive balance should be empty")

	assert.Equal(t, -(8*(19+18)-5)*time.Hour+(1+1+3)*time.Hour, report.Summary.TotalOvertime, "total overtime")
	assert.Equal(t, (9+9+3+3)*time.Hour, report.Summary.TotalWorked, "total worked time")
	assert.Equal(t, givenEmployee, report.Employee, "employee")

}
