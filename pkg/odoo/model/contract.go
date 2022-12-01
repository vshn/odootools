package model

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/vshn/odootools/pkg/odoo"
)

type Contract struct {
	ID float64 `json:"id"`
	// Start is the first day of the contract in UTC.
	Start odoo.Date `json:"date_start"`
	// Start is the last day of the contract in UTC.
	// It is Zero if the contract hasn't ended yet.
	End             odoo.Date        `json:"date_end"`
	WorkingSchedule *WorkingSchedule `json:"working_hours"`
}

// ContractList contains a slice of Contract.
type ContractList odoo.List[Contract]

// NoContractCoversDateErr is an error that indicates a contract doesn't cover a date.
type NoContractCoversDateErr struct {
	Err  error
	Date time.Time
}

// Error implements error.
func (e *NoContractCoversDateErr) Error() string {
	return e.Err.Error()
}

// Unwrap implements Wrapper.
func (e *NoContractCoversDateErr) Unwrap() error {
	return e.Err
}

// GetFTERatioForDay returns the workload ratio that is active for the given day.
// All involved dates are expected to be in UTC.
func (l ContractList) GetFTERatioForDay(date time.Time) (float64, error) {
	for _, contract := range l.Items {
		t := contract.Start

		start := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, date.Location()).Add(-1 * time.Second)
		if contract.End.IsZero() {
			// current contract
			if start.Before(date) {
				return contract.WorkingSchedule.GetFTERatio()
			}
			continue
		}
		end := contract.End.Add(1 * time.Second)
		if start.Before(date) && end.After(date) {
			return contract.WorkingSchedule.GetFTERatio()
		}
	}
	return 0, &NoContractCoversDateErr{
		Err:  fmt.Errorf("no contract found that covers date: %s", date),
		Date: date,
	}
}

// Sort sorts the contracts
func (l ContractList) Sort() {
	sort.Slice(l.Items, func(i, j int) bool {
		a := l.Items[i]
		b := l.Items[j]
		return a.Start.Before(b.Start.Time)
	})
}

func (o Odoo) FetchAllContractsOfEmployee(ctx context.Context, employeeID int) (ContractList, error) {
	return o.readContracts(ctx, []odoo.Filter{
		[]interface{}{"employee_id", "=", employeeID},
	})
}

func (o Odoo) readContracts(ctx context.Context, domainFilters []odoo.Filter) (ContractList, error) {
	result := ContractList{}
	err := o.querier.SearchGenericModel(ctx, odoo.SearchReadModel{
		Model:  "hr.contract",
		Domain: domainFilters,
		Fields: []string{"date_start", "date_end", "working_hours"},
	}, &result)
	return result, err
}
