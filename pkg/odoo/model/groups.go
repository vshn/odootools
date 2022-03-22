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

// GroupList contains a slice of Group.
type GroupList struct {
	Items []Group `json:"records,omitempty"`
}

func (o Odoo) FetchGroupByName(ctx context.Context, category, name string) (*Group, error) {
	groups, err := o.searchGroups(ctx, []odoo.Filter{
		[]string{"name", "=", name},
		[]interface{}{"category_id.name", "=", category},
	})
	if err != nil {
		return nil, err
	}
	if len(groups.Items) > 0 {
		return &groups.Items[0], nil
	}
	return nil, nil
}

func (o Odoo) searchGroups(ctx context.Context, domainFilters []odoo.Filter) (GroupList, error) {
	result := GroupList{}
	err := o.querier.SearchGenericModel(ctx, odoo.SearchReadModel{
		Model:  "res.groups",
		Domain: domainFilters,
		Fields: []string{"users", "category_id"},
		Limit:  0,
		Offset: 0,
	}, &result)
	return result, err
}
