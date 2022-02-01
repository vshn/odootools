package model

import (
	"context"
	"fmt"
	"time"

	"github.com/vshn/odootools/pkg/odoo"
)

type Contract struct {
	ID float64 `json:"id"`
	// Start is the first day of the contract in UTC.
	Start *odoo.Date `json:"date_start"`
	// Start is the last day of the contract in UTC.
	// It is nil or Zero if the contract hasn't ended yet.
	End             *odoo.Date       `json:"date_end"`
	WorkingSchedule *WorkingSchedule `json:"working_hours"`
}

// ContractList contains a slice of Contract.
type ContractList struct {
	Items []Contract `json:"records,omitempty"`
}

// GetFTERatioForDay returns the workload ratio that is active for the given day.
// All involved dates are expected to be in UTC.
func (l ContractList) GetFTERatioForDay(day odoo.Date) (float64, error) {
	date := day.ToTime()
	for _, contract := range l.Items {
		start := contract.Start.ToTime().Add(-1 * time.Second)
		if contract.End.IsZero() {
			// current contract
			if start.Before(date) {
				return contract.WorkingSchedule.GetFTERatio()
			}
			continue
		}
		end := contract.End.ToTime().Add(1 * time.Second)
		if start.Before(date) && end.After(date) {
			return contract.WorkingSchedule.GetFTERatio()
		}
	}
	return 0, fmt.Errorf("no contract found that covers date: %s", day.String())
}

func (o Odoo) FetchAllContracts(employeeID int) (ContractList, error) {
	return o.readContracts([]odoo.Filter{
		[]interface{}{"employee_id", "=", employeeID},
	})
}

func (o Odoo) readContracts(domainFilters []odoo.Filter) (ContractList, error) {
	result := ContractList{}
	err := o.querier.SearchGenericModel(context.Background(), odoo.SearchReadModel{
		Model:  "hr.contract",
		Domain: domainFilters,
		Fields: []string{"date_start", "date_end", "working_hours"},
	}, &result)
	return result, err
}
