package odoo

import (
	"fmt"
	"time"
)

type ContractList []Contract

type Contract struct {
	ID float64 `json:"id"`
	// Start is the first day of the contract in UTC.
	Start *Date `json:"date_start"`
	// Start is the last day of the contract in UTC.
	// It is nil or Zero if the contract hasn't ended yet.
	End             *Date            `json:"date_end"`
	WorkingSchedule *WorkingSchedule `json:"working_hours"`
}

// GetFTERatioForDay returns the workload ratio that is active for the given day.
// All involved dates are expected to be in UTC.
func (l ContractList) GetFTERatioForDay(day Date) (float64, error) {
	date := day.ToTime()
	for _, contract := range l {
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

func (c Client) FetchAllContracts(sid string, employeeID int) (ContractList, error) {
	return c.readContracts(sid, []Filter{
		[]interface{}{"employee_id", "=", employeeID},
	})
}

func (c Client) readContracts(sid string, domainFilters []Filter) (ContractList, error) {
	// Prepare "search contract" request
	body, err := NewJsonRpcRequest(&ReadModelRequest{
		Model:  "hr.contract",
		Domain: domainFilters,
		Fields: []string{"date_start", "date_end", "working_hours"},
	}).Encode()
	if err != nil {
		return nil, fmt.Errorf("encoding request: %w", err)
	}

	res, err := c.makeRequest(sid, body)
	if err != nil {
		return nil, err
	}

	type readResult struct {
		Length  int        `json:"length,omitempty"`
		Records []Contract `json:"records,omitempty"`
	}
	result := &readResult{}
	if err := c.unmarshalResponse(res.Body, result); err != nil {
		return nil, err
	}
	return result.Records, nil
}
