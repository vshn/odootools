package timesheet

import (
	"fmt"
	"strconv"
	"time"
)

type DailySummary struct {
	Date   time.Time
	Blocks []AttendanceBlock
}

func (s *DailySummary) addAttendanceBlock(block AttendanceBlock) {
	if len(s.Blocks) == 0 {
		s.Blocks = []AttendanceBlock{block}
		s.Date = block.Start.Truncate(24 * time.Hour)
		return
	}
	if block.Start.Day() != s.Date.Day() {
		// Block is not on the same day
		return
	}
	s.Blocks = append(s.Blocks, block)
}

func (s DailySummary) CalculateOvertime() time.Duration {
	workHours := float64(0)
	for _, block := range s.Blocks {
		if block.Reason == "" {
			workHours += block.End.Sub(block.Start).Hours()
			continue
		}
		// TODO: Add 1.5 factor in special cases
		// TODO: respect attendance reasons (e.g. no overtime in sick leave)

	}

	// TODO: respect FTE percentage
	overtime := workHours - 8

	duration, err := time.ParseDuration(fmt.Sprintf("%sh", strconv.FormatFloat(overtime, 'f', 2, 64)))
	if err != nil {
		panic(err)
	}
	return duration
}

func findDailySummaryByDate(dailies []*DailySummary, date time.Time) (*DailySummary, bool) {
	for _, daily := range dailies {
		if daily.Date.Day() == date.Day() {
			return daily, true
		}
	}
	return nil, false
}
