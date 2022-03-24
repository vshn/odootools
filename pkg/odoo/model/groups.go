package model

import (
	"context"

	"github.com/vshn/odootools/pkg/odoo"
)

// Group contains a list of users.
type Group struct {
	Name     string        `json:"name"`
	Category GroupCategory `json:"category_id"`
	UserIDs  []int         `json:"users"`
}

func (o Odoo) FetchGroupByName(ctx context.Context, category, name string) (*Group, error) {
	groups, err := o.searchGroups(ctx, []odoo.Filter{
		[]string{"name", "=", name},
		[]interface{}{"category_id.name", "=", category},
	})
	if err != nil {
		return nil, err
	}
	if groups.Len() > 0 {
		return &groups.Items[0], nil
	}
	return nil, nil
}

func (o Odoo) searchGroups(ctx context.Context, domainFilters []odoo.Filter) (odoo.List[Group], error) {
	result := odoo.List[Group]{}
	err := o.querier.SearchGenericModel(ctx, odoo.SearchReadModel{
		Model:  "res.groups",
		Domain: domainFilters,
		Fields: []string{"users", "category_id"},
		Limit:  0,
		Offset: 0,
	}, &result)
	return result, err
}
