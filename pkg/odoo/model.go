package odoo

// ReadModelRequest is used as "params" in requests to "dataset/search_read"
// endpoints.
type ReadModelRequest struct {
	Model  string   `json:"model,omitempty"`
	Domain []Filter `json:"domain,omitempty"`
	Fields []string `json:"fields,omitempty"`
	Limit  int      `json:"limit,omitempty"`
	Offset int      `json:"offset,omitempty"`
}

// Filter to use in queries, usually in the format of
// [predicate, operator, value], eg ["employee_id.user_id.id", "=", 123]
type Filter []interface{}
