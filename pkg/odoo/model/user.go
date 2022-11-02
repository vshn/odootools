package model

import (
	"context"

	"github.com/vshn/odootools/pkg/odoo"
)

type User struct {
	ID       int            `json:"id"`
	Name     string         `json:"name"`
	TimeZone *odoo.TimeZone `json:"tz,omitempty"`
	Email    string         `json:"email"`
}

func (o Odoo) FetchUserByID(ctx context.Context, id int) (*User, error) {
	users, err := o.readUser(ctx, []odoo.Filter{
		[]interface{}{"id", "=", id},
	})
	if err != nil {
		return nil, err
	}
	if users.Len() > 0 {
		return &users.Items[0], nil
	}
	return nil, nil
}

func (o Odoo) readUser(ctx context.Context, domainFilters []odoo.Filter) (odoo.List[User], error) {
	result := odoo.List[User]{}
	err := o.querier.SearchGenericModel(ctx, odoo.SearchReadModel{
		Model:  "res.users",
		Domain: domainFilters,
		Fields: []string{"name", "tz", "email"},
		Limit:  0,
		Offset: 0,
	}, &result)
	return result, err
}
