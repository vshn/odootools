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
	// Prepare request
	body, err := NewJsonRpcRequest(&ReadModelRequest{
		Model:  "hr.employee",
		Domain: []Filter{[]string{"name", "ilike", searchString}},
		Fields: []string{"name"},
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

// FetchEmployee fetches an Employee for the given user ID.
// Returns nil if not found.
func (c *Client) FetchEmployee(sid string, userId int) (*Employee, error) {
	// Prepare request
	body, err := NewJsonRpcRequest(&ReadModelRequest{
		Model:  "hr.employee",
		Domain: []Filter{[]interface{}{"user_id", "=", userId}},
		Fields: []string{"name"},
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
