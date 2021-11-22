package odoo

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var workingScheduleRegex = regexp.MustCompile("(?P<ratio>[0-9]+\\s*%)")

type WorkingSchedule struct {
	ID   float64
	Name string
}

func (s *WorkingSchedule) String() string {
	if s == nil {
		return ""
	}
	return s.Name
}

func (s WorkingSchedule) MarshalJSON() ([]byte, error) {
	if s.Name == "" {
		return []byte("false"), nil
	}
	arr := []interface{}{s.ID, s.Name}
	return json.Marshal(arr)
}
func (s *WorkingSchedule) UnmarshalJSON(b []byte) error {
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
			*s = WorkingSchedule{
				ID:   arr[0].(float64),
				Name: v,
			}
		}
	}
	return nil
}

// GetFTERatio tries to extract the FTE ratio from the name of the schedule.
// It returns an error if it could not find a match
func (s *WorkingSchedule) GetFTERatio() (float64, error) {
	match := workingScheduleRegex.FindStringSubmatch(s.Name)
	if len(match) > 0 {
		v := match[0]
		v = strings.TrimSpace(v)
		v = strings.ReplaceAll(v, " ", "") // there might be spaces in between
		v = strings.ReplaceAll(v, "%", "")
		ratio, err := strconv.Atoi(v)
		return float64(ratio) / 100, err
	}
	return 0, fmt.Errorf("could not find FTE ratio in name: '%s'", s.Name)
}
