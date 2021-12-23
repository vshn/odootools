package timesheet

import (
	"time"
)

type DailySummary struct {
	// Date is the localized date of the summary.
	Date     time.Time
	Blocks   []AttendanceBlock
	Absences []AbsenceBlock
	FTERatio float64
}

// NewDailySummary creates a new instance.
// The fteRatio is the percentage (input a value between 0..1) of the employee and is used to calculate the daily maximum hours an employee should work.
// Date is expected to be in a localized timezone.
func NewDailySummary(fteRatio float64, date time.Time) *DailySummary {
	return &DailySummary{
		FTERatio: fteRatio,
		Date:     date,
		Absences: []AbsenceBlock{},
		Blocks:   []AttendanceBlock{},
	}
}

// addAttendanceBlock adds the given block to the existing blocks.
// If the block is not starting in the same day as DailySummary.Date, it will be silently ignored.
func (s *DailySummary) addAttendanceBlock(block AttendanceBlock) {
	if block.Start.Day() != s.Date.Day() {
		// Block is not on the same day
		return
	}
	s.Blocks = append(s.Blocks, block)
}

// addAbsenceBlock adds the given block to the existing absences.
func (s *DailySummary) addAbsenceBlock(block AbsenceBlock) {
	// At VSHN, currently only full-day absences are possible, so no need to check for starting and ending time.
	s.Absences = append(s.Absences, block)
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
func (s *DailySummary) CalculateOvertime() time.Duration {
	workingTime := s.CalculateWorkingTime()
	excusedTime := s.CalculateExcusedTime()

	dailyMax := s.CalculateDailyMax() - s.CalculateAbsenceTime()
	if workingTime >= dailyMax {
		// Can't be on sick leave etc. if working overtime.
		excusedTime = 0
	} else if workingTime+excusedTime > dailyMax {
		// There is overlap: Not enough workHours, but having excused hours = Cap at daily max, no overtime
		excusedTime = dailyMax - workingTime
	}
	overtime := workingTime + excusedTime - dailyMax

	return overtime
}

// CalculateDailyMax returns the theoretical amount of hours that an employee should work on this day.
//  * It returns 0 for weekend days.
//  * It returns 8.5 hours multiplied by FTE ratio for days in 2020 and earlier.
//  * It returns 8.0 hours multiplied by FTE ratio for days in 2021 and later.
func (s *DailySummary) CalculateDailyMax() time.Duration {
	if s.IsWeekend() {
		return 0
	}
	if s.Date.Year() < 2021 {
		// VSHN switched from 42h-a-week to 40h-a-week on 1st of January 2021.
		return time.Duration(8.5 * s.FTERatio * float64(time.Hour))
	}
	return time.Duration(8 * s.FTERatio * float64(time.Hour))
}

// CalculateWorkingTime accumulates all working hours from that day.
// The outside office hours are multiplied with 1.5.
func (s *DailySummary) CalculateWorkingTime() time.Duration {
	workTime := time.Duration(0)
	for _, block := range s.Blocks {
		switch block.Reason {
		case "":
			diff := block.End.Sub(block.Start)
			workTime += diff
		case ReasonOutsideOfficeHours:
			diff := 1.5 * float64(block.End.Sub(block.Start))
			workTime += time.Duration(diff)
		}
	}
	return workTime
}

// CalculateExcusedTime accumulates all hours that are excused in some way (sick leave etc) from that day.
func (s *DailySummary) CalculateExcusedTime() time.Duration {
	total := time.Duration(0)
	for _, block := range s.Blocks {
		switch block.Reason {
		case ReasonSickLeave, ReasonAuthorities, ReasonPublicService:
			diff := block.End.Sub(block.Start)
			total += diff
		}
	}
	return total
}

// CalculateAbsenceTime accumulates all absence hours from that day.
func (s *DailySummary) CalculateAbsenceTime() time.Duration {
	total := time.Duration(0)
	for _, absence := range s.Absences {
		if absence.Reason != TypeUnpaid {
			// VSHN specific: Odoo treats "Unpaid" as normal leave, but for VSHN it's informational-only, meaning one still has to work.
			// For every other type of absence, we add the daily max equivalent.

			total += s.CalculateDailyMax()
		}
	}
	return total
}

// HasAbsences returns true if there are any absences.
func (s *DailySummary) HasAbsences() bool {
	return len(s.Absences) != 0
}

// IsWeekend returns true if the date falls on a Saturday or Sunday.
func (s *DailySummary) IsWeekend() bool {
	return s.Date.Weekday() == time.Saturday || s.Date.Weekday() == time.Sunday
}

func findDailySummaryByDate(dailies []*DailySummary, date time.Time) (*DailySummary, bool) {
	for _, daily := range dailies {
		if daily.Date.Day() == date.Day() {
			return daily, true
		}
	}
	return nil, false
}
