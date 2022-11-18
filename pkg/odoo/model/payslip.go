package model

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"time"

	"github.com/vshn/odootools/pkg/odoo"
)

type Payslip struct {
	ID        int            `json:"id"`
	Name      string         `json:"name"`
	DateFrom  odoo.Date      `json:"date_from"`
	DateTo    odoo.Date      `json:"date_to"`
	XOvertime interface{}    `json:"x_overtime"`
	TimeZone  *odoo.TimeZone `json:"x_timezone"`
}

type PayslipList odoo.List[Payslip]

// FilterInMonth returns the payslip that covers the given date.
// Returns nil if no matching payslip found.
// Respects time.Location if the given date has also a time.Location set.
func (l PayslipList) FilterInMonth(dayInMonth time.Time) *Payslip {
	for _, payslip := range l.Items {
		from := payslip.DateFrom.Time
		to := payslip.DateTo.Time
		from = time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, dayInMonth.Location())
		to = time.Date(to.Year(), to.Month(), to.Day(), 0, 0, 0, 0, dayInMonth.Location())
		if odoo.IsWithinTimeRange(dayInMonth, from, to) {
			return &payslip
		}
	}
	return nil
}

func (l PayslipList) Len() int {
	return len(l.Items)
}

func (o Odoo) FetchPayslipInMonth(ctx context.Context, employeeID int, firstDayOfMonth time.Time) (*Payslip, error) {
	payslips, err := o.readPayslips(ctx, []odoo.Filter{
		[]interface{}{"employee_id", "=", employeeID},
		[]string{"date_from", ">=", firstDayOfMonth.AddDate(0, 0, -1).Format(odoo.DateFormat)},
		[]string{"date_to", "<=", firstDayOfMonth.AddDate(0, 1, -1).Format(odoo.DateFormat)},
		[]string{"name", "ilike", "Salary Slip"},
	})
	if payslips.Len() > 0 {
		return &payslips.Items[0], err
	}
	return nil, err
}

// FetchPayslipBetween returns all payslips that are between the given dates.
// It may return empty list if none found.
// If multiple found, they are sorted ascending by to their Payslip.DateFrom (earliest first).
func (o Odoo) FetchPayslipBetween(ctx context.Context, employeeID int, firstDay, lastDay time.Time) (PayslipList, error) {
	payslips, err := o.readPayslips(ctx, []odoo.Filter{
		[]interface{}{"employee_id", "=", employeeID},
		[]string{"date_from", ">=", firstDay.AddDate(0, 0, -1).Format(odoo.DateFormat)},
		[]string{"date_to", "<=", lastDay.Format(odoo.DateFormat)},
		[]string{"name", "ilike", "Salary Slip"},
	})
	// sort by start date ascending
	sort.SliceStable(payslips.Items, func(i, j int) bool {
		fromFirst := payslips.Items[i].DateFrom.Time
		fromSecond := payslips.Items[j].DateFrom.Time
		return fromFirst.Before(fromSecond)
	})
	return payslips, err
}

func (o Odoo) UpdatePayslip(ctx context.Context, payslip *Payslip) error {
	err := o.querier.UpdateGenericModel(ctx, "hr.payslip", payslip.ID, payslip)
	return err
}

func (o Odoo) readPayslips(ctx context.Context, domainFilters []odoo.Filter) (PayslipList, error) {
	result := PayslipList{}
	err := o.querier.SearchGenericModel(ctx, odoo.SearchReadModel{
		Model:  "hr.payslip",
		Domain: domainFilters,
		Fields: []string{"date_from", "date_to", "x_overtime", "name", "x_timezone"},
	}, &result)
	return result, err
}

// Overtime returns the plain field value as string.
func (p Payslip) Overtime() string {
	if p.XOvertime == nil {
		return ""
	}
	if _, ok := p.XOvertime.(bool); ok {
		return ""
	}
	return p.XOvertime.(string)
}

// colonFormatRegex searches for string reference that has somewhere a pattern like '123:45' or '123:45:54'
// A match will be divided into subgroups, e.g. '123' for hours, '45' for minutes, '54' for seconds.
// The hours group can have a dash in front of the number to indicate negative hours.
var colonFormatRegex = regexp.MustCompile(`.*?((-?\d+):(\d{2})(?::?(\d{2}))?).*`)

// ParseOvertime tries to parse the currently inconsistently-formatted custom field to a duration.
// If the field is empty, 0 is returned without error.
// It parses the following formats:
//   - hhh:mm (e.g. '15:54')
//   - hhh:mm:ss (e.g. '153:54:45')
//   - {1,2}d{1,2}h (e.g. '15d54m')
func (p Payslip) ParseOvertime() (time.Duration, error) {
	raw := p.Overtime()
	if raw == "" {
		return 0, nil
	}
	if matches := colonFormatRegex.FindStringSubmatch(raw); matches != nil {
		rawHours := matches[2]
		rawMinutes := matches[3]
		rawSeconds := matches[4]
		if rawSeconds == "" {
			rawSeconds = "0"
		}
		t, err := time.ParseDuration(fmt.Sprintf("%sh%sm%ss", rawHours, rawMinutes, rawSeconds))
		return t, err
	}
	return 0, fmt.Errorf("format not parseable: %s", raw)
}
