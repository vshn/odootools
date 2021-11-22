package odoo

import (
	"fmt"
	"time"
)

type ContractList []Contract

type Contract struct {
	ID              float64          `json:"id"`
	Start           *Date            `json:"date_start"`
	End             *Date            `json:"date_end"`
	WorkingSchedule *WorkingSchedule `json:"working_hours"`
}

func (l ContractList) GetFTERatioForDay(day Date) (float64, error) {
	date := day.ToTime()
	for _, contract := range l {
		if contract.End.IsZero() {
			// current contract
			if contract.Start.ToTime().Add(-1 * time.Second).Before(date) {
				return contract.WorkingSchedule.GetFTERatio()
			}
			continue
		}
		start := contract.Start.ToTime().Add(-1 * time.Second)
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
