package model

import "encoding/json"

// GroupCategory is the parent group of a Group.
type GroupCategory struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// String implements fmt.Stringer.
func (c *GroupCategory) String() string {
	if c == nil {
		return ""
	}
	return c.Name
}

// MarshalJSON implements json.Marshaler.
func (c *GroupCategory) MarshalJSON() ([]byte, error) {
	arr := []interface{}{c.ID, c.Name}
	return json.Marshal(arr)
}

// UnmarshalJSON implements json.Unmarshaler.
func (c *GroupCategory) UnmarshalJSON(b []byte) error {
	var f bool
	if err := json.Unmarshal(b, &f); err == nil || string(b) == "false" {
		return nil
	}
	var arr []interface{}
	if err := json.Unmarshal(b, &arr); err != nil {
		return err
	}
	if len(arr) >= 2 {
		if v, ok := arr[1].(string); ok {
			*c = GroupCategory{
				ID:   int(arr[0].(float64)),
				Name: v,
			}
		}
	}
	return nil
}
