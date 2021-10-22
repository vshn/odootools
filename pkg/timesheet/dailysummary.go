package timesheet

import (
	"fmt"
	"strconv"
	"time"
)

type DailySummary struct {
	Date     time.Time
	Blocks   []AttendanceBlock
	Overtime time.Duration
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

func (s *DailySummary) CalculateOvertime(fteRatio float64) time.Duration {
	workHours := float64(0)
	excusedHours := float64(0)
	for _, block := range s.Blocks {
		switch block.Reason {
		case "":
			diff := block.End.Sub(block.Start).Hours()
			workHours += diff
		case ReasonSickLeave, ReasonAuthorities, ReasonPublicService:
			diff := block.End.Sub(block.Start).Hours()
			excusedHours += diff
		case ReasonOutsideOfficeHours:
			diff := block.End.Sub(block.Start).Hours() * 1.5
			workHours += diff
		}
	}
	dailyMax := 8 * fteRatio
	if workHours >= dailyMax {
		// Can't be on sick leave etc. if working overtime.
		excusedHours = 0
	} else if workHours+excusedHours > dailyMax {
		// There is overlap: Not enough workHours, but having excused hours = Cap at daily max, no overtime
		excusedHours = dailyMax - workHours
	}
	overtime := workHours + excusedHours - dailyMax

	duration, err := time.ParseDuration(fmt.Sprintf("%sh", strconv.FormatFloat(overtime, 'f', 2, 64)))
	if err != nil {
		panic(err)
	}
	s.Overtime = duration
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
