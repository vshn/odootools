package odoo

// List contains a slice of T.
type List[T any] struct {
	Items []T `json:"records,omitempty"`
}

// Len returns the length of List.Items.
func (l *List[T]) Len() int {
	// for some reason len(l.Items) doesn't compile?
	if l == nil || l.Items == nil {
		return 0
	}
	i := 0
	for range l.Items {
		i++
	}
	return i
}
