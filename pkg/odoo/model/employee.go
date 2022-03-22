package model

import (
	"context"

	"github.com/vshn/odootools/pkg/odoo"
)

type Employee struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// EmployeeList contains a slice of Employee.
type EmployeeList struct {
	Items []Employee `json:"records,omitempty"`
}

// SearchEmployee searches for an Employee with the given searchString in the Employee.Name.
// If multiple employees are found, the first is returned.
// Returns nil if none found.
func (o Odoo) SearchEmployee(ctx context.Context, searchString string) (*Employee, error) {
	return o.readEmployee(ctx, []odoo.Filter{[]string{"name", "ilike", searchString}})
}

// FetchEmployeeByID fetches an Employee for the given employee ID.
// Returns nil if not found.
func (o Odoo) FetchEmployeeByID(ctx context.Context, employeeID int) (*Employee, error) {
	return o.readEmployee(ctx, []odoo.Filter{[]interface{}{"resource_id", "=", employeeID}})
}

// FetchEmployeeByUserID fetches the Employee for the given user ID (which might not be the same as Employee.ID.
// Returns nil if not found.
func (o Odoo) FetchEmployeeByUserID(ctx context.Context, userID int) (*Employee, error) {
	return o.readEmployee(ctx, []odoo.Filter{[]interface{}{"user_id", "=", userID}})
}

func (o Odoo) readEmployee(ctx context.Context, filters []odoo.Filter) (*Employee, error) {
	result := EmployeeList{}
	err := o.querier.SearchGenericModel(ctx, odoo.SearchReadModel{
		Model:  "hr.employee",
		Domain: filters,
		Fields: []string{"name"},
	}, &result)
	if err != nil {
		return nil, err
	}
	if len(result.Items) > 0 {
		return &result.Items[0], nil
	}
	// not found
	return nil, nil
}
