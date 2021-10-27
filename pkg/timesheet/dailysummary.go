package timesheet

import (
	"fmt"
	"strconv"
	"time"
)

type DailySummary struct {
	Date     time.Time
	Blocks   []AttendanceBlock
	FTERatio float64
}

// NewDailySummary creates a new instance.
// The fteRatio is the percentage (input a value between 0..1) of the employee and is used to calculate the daily maximum hours an employee should work.
func NewDailySummary(fteRatio float64, date time.Time) *DailySummary {
	return &DailySummary{
		FTERatio: fteRatio,
		Date:     date,
	}
}

// addAttendanceBlock adds the given block to the existing blocks.
// If the block is first in the list, it will set a truncated date in DailySummary.Date.
// If the next block is not starting in the same day, it will be silently ignored.
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

// CalculateOvertime returns the duration of overtime.
// If returned duration is positive, then the employee did overtime and undertime if duration is negative.
//
// The overtime is then calculated according to these business rules:
//  * Outside office hours are multiplied by 1.5 (as a compensation)
//  * Excused hours like sick leave, authorities or public service can be used to "fill up" the daily theoretical maximum if the working hours are less than said maximum.
//    However, there's no overtime possible using excused hours
//  * If the working hours exceed the theoretical daily maximum, then the excused hours are basically ignored.
//    Example: it's not possible to work 9 hours, have 1 hour sick leave and expect 2 hours overtime for an 8 hours daily maximum, the overtime here is 1 hour.
//  * Theoretical daily maximum is 0 (zero) hours on weekends.
func (s *DailySummary) CalculateOvertime() time.Duration {
	workHours := s.CalculateWorkingHours()
	excusedHours := s.CalculateExcusedHours()

	dailyMax := s.CalculateDailyMaxHours()
	if workHours >= dailyMax {
		// Can't be on sick leave etc. if working overtime.
		excusedHours = 0
	} else if workHours+excusedHours > dailyMax {
		// There is overlap: Not enough workHours, but having excused hours = Cap at daily max, no overtime
		excusedHours = dailyMax - workHours
	}
	overtime := workHours + excusedHours - dailyMax

	return toDuration(overtime)
}

// CalculateDailyMaxHours returns the theoretical amount of hours that an employee should work on this day.
//  * It returns 0 for weekend days.
//  * It returns 8.5 hours multiplied by FTE ratio for days in 2020 and earlier.
//  * It returns 8.0 hours multiplied by FTE ratio for days in 2021 and later.
func (s *DailySummary) CalculateDailyMaxHours() float64 {
	if s.Date.Weekday() == time.Saturday || s.Date.Weekday() == time.Sunday {
		return 0
	}
	if s.Date.Year() < 2021 {
		// VSHN switched from 42h-a-week to 40h-a-week on 1st of January 2021.
		return 8.5 * s.FTERatio
	}
	return 8 * s.FTERatio
}

// CalculateWorkingHours accumulates all working hours from that day.
// The outside office hours are multiplied with 1.5.
func (s *DailySummary) CalculateWorkingHours() float64 {
	workHours := float64(0)
	for _, block := range s.Blocks {
		switch block.Reason {
		case "":
			diff := block.End.Sub(block.Start).Hours()
			workHours += diff
		case ReasonOutsideOfficeHours:
			diff := block.End.Sub(block.Start).Hours() * 1.5
			workHours += diff
		}
	}
	return workHours
}

// CalculateExcusedHours accumulates all hours that are excused in some way (sick leave etc) from that day.
func (s *DailySummary) CalculateExcusedHours() float64 {
	excusedHours := float64(0)
	for _, block := range s.Blocks {
		switch block.Reason {
		case ReasonSickLeave, ReasonAuthorities, ReasonPublicService:
			diff := block.End.Sub(block.Start).Hours()
			excusedHours += diff
		}
	}
	return excusedHours
}

func findDailySummaryByDate(dailies []*DailySummary, date time.Time) (*DailySummary, bool) {
	for _, daily := range dailies {
		if daily.Date.Day() == date.Day() {
			return daily, true
		}
	}
	return nil, false
}

func toDuration(hours float64) time.Duration {
	duration, err := time.ParseDuration(fmt.Sprintf("%sh", strconv.FormatFloat(hours, 'f', 2, 64)))
	if err != nil {
		panic(err)
	}
	return duration
}
