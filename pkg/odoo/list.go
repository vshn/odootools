package odoo

// List contains a slice of T.
type List[T any] struct {
	Items []T `json:"records,omitempty"`
}

// Len returns the length of List.Items.
// This is a pure utility/shortcut function for `len(l.Items)` with nil check.
func (l *List[T]) Len() int {
	if l == nil || l.Items == nil {
		return 0
	}
	return len(l.Items)
}
