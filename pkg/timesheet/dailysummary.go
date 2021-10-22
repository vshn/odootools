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
		switch block.Reason {
		case "":
			workHours += block.End.Sub(block.Start).Hours()
		case ReasonOutsideOfficeHours:
			workHours += block.End.Sub(block.Start).Hours() * 1.5
		}
		// TODO: respect sick leave:
		/* Sick leaves don't get counted if logged hours exceed daily FTE
		Examples with 8 hrs:
		- 7 hours logged time + 1h sick leave = 8hrs, 0 overtime
		- 8 hours logged time + 1h sick leave = 8hrs, 0 overtime
		- 9 hours logged time + 1h sick leave = 9 hrs, 1h overtime (sick leave doesn't count, it's not 10-8=2h)
		- 6 hours logged time + 1h sick leave = 7hrs, -1h overtime (1h undertime)
		*/
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
