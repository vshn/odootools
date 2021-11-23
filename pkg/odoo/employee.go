package odoo

import (
	"fmt"
)

type Employee struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// SearchEmployee searches for an Employee with the given searchString in the Employee.Name.
// If multiple employees are found, the first is returned.
// Returns nil if none found.
func (c *Client) SearchEmployee(searchString string, sid string) (*Employee, error) {
	return c.readEmployee(sid, []Filter{[]string{"name", "ilike", searchString}})
}

// FetchEmployeeByID fetches an Employee for the given employee ID.
// Returns nil if not found.
func (c *Client) FetchEmployeeByID(sid string, employeeID int) (*Employee, error) {
	return c.readEmployee(sid, []Filter{[]interface{}{"resource_id", "=", employeeID}})
}

// FetchEmployeeBySession fetches the Employee for the given session.
// Returns nil if not found.
func (c *Client) FetchEmployeeBySession(session *Session) (*Employee, error) {
	return c.readEmployee(session.ID, []Filter{[]interface{}{"user_id", "=", session.UID}})
}

func (c *Client) readEmployee(sid string, filters []Filter) (*Employee, error) {
	// Prepare request
	body, err := NewJsonRpcRequest(&ReadModelRequest{
		Model:  "hr.employee",
		Domain: filters,
		//Fields: []string{"name"},
		Limit:  0,
		Offset: 0,
	}).Encode()
	if err != nil {
		return nil, fmt.Errorf("encoding request: %w", err)
	}

	res, err := c.makeRequest(sid, body)
	if err != nil {
		return nil, err
	}
	type readResults struct {
		Length  int        `json:"length,omitempty"`
		Records []Employee `json:"records,omitempty"`
	}

	result := &readResults{}
	if err := c.unmarshalResponse(res.Body, result); err != nil {
		return nil, err
	}
	if len(result.Records) >= 1 {
		return &result.Records[0], nil
	}
	return nil, nil
}
