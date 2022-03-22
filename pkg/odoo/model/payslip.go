package model

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/vshn/odootools/pkg/odoo"
)

type Payslip struct {
	ID       int         `json:"id"`
	Name     string      `json:"name"`
	Overtime interface{} `json:"x_overtime"`
	DateFrom odoo.Date   `json:"date_from"`
	DateTo   odoo.Date   `json:"date_to"`
}

// PayslipList contains a slice of Payslip.
type PayslipList struct {
	Items []Payslip `json:"records,omitempty"`
}

func (o Odoo) FetchPayslipOfLastMonth(ctx context.Context, employeeID int, lastDayOfMonth time.Time) (*Payslip, error) {
	payslips, err := o.readPayslips(ctx, []odoo.Filter{
		[]interface{}{"employee_id", "=", employeeID},
		[]string{"date_to", "<=", lastDayOfMonth.Format(odoo.DateFormat)},
		[]string{"date_from", ">=", lastDayOfMonth.AddDate(0, -1, -1).Format(odoo.DateFormat)},
	})
	for _, payslip := range payslips.Items {
		if strings.Contains(payslip.Name, "Pikett") {
			continue
		}
		if strings.Contains(payslip.Name, "Salary") {
			return &payslip, nil
		}
	}
	return nil, err
}

func (o Odoo) readPayslips(ctx context.Context, domainFilters []odoo.Filter) (PayslipList, error) {
	result := PayslipList{}
	err := o.querier.SearchGenericModel(ctx, odoo.SearchReadModel{
		Model:  "hr.payslip",
		Domain: domainFilters,
		Fields: []string{"date_from", "date_to", "x_overtime", "name"},
	}, &result)
	return result, err
}

// GetOvertime returns the plain field value as string.
func (p Payslip) GetOvertime() string {
	if _, ok := p.Overtime.(bool); ok {
		return ""
	}
	return p.Overtime.(string)
}

// colonFormatRegex searches for string reference that has somewhere a pattern like '123:45' or '123:45:54'
// A match will be divided into subgroups, e.g. '123' for hours, '45' for minutes, '54' for seconds.
// The hours group can have a dash in front of the number to indicate negative hours.
var colonFormatRegex = regexp.MustCompile(".*?((-?\\d+):(\\d{2})(?::?(\\d{2}))?).*")

// ParseOvertime tries to parse the currently inconsistently-formatted custom field to a duration.
// If the field is empty, 0 is returned without error.
// It parses the following formats:
//  * hhh:mm (e.g. '15:54')
//  * hhh:mm:ss (e.g. '153:54:45')
//  * {1,2}d{1,2}h (e.g. '15d54m')
func (p Payslip) ParseOvertime() (time.Duration, error) {
	raw := p.GetOvertime()
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
