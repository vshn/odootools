package timesheet

import (
	"time"
)

type DailySummary struct {
	// Date is the localized date of the summary.
	Date     time.Time
	Shifts   []AttendanceShift
	Absences []AbsenceBlock
	FTERatio float64
}

type OvertimeSummary struct {
	RegularWorkingTime time.Duration
	SickLeaveTime      time.Duration
	AuthoritiesTime    time.Duration
	OutOfOfficeTime    time.Duration
	DailyMax           time.Duration
	PublicServiceTime  time.Duration
}

// NewDailySummary creates a new instance.
// The fteRatio is the percentage (input a value between 0..1) of the employee and is used to calculate the daily maximum hours an employee should work.
// Date is expected to be in a localized timezone.
func NewDailySummary(fteRatio float64, date time.Time) *DailySummary {
	return &DailySummary{
		FTERatio: fteRatio,
		Date:     date,
		Absences: []AbsenceBlock{},
		Shifts:   []AttendanceShift{},
	}
}

// addAttendanceShift adds the given shift to the existing shifts.
// If the shift is not starting in the same day as DailySummary.Date, it will be silently ignored.
func (s *DailySummary) addAttendanceShift(shift AttendanceShift) {
	if shift.Start.Day() != s.Date.Day() {
		// Shift is not on the same day
		return
	}
	s.Shifts = append(s.Shifts, shift)
}

// addAbsenceBlock adds the given block to the existing absences.
func (s *DailySummary) addAbsenceBlock(block AbsenceBlock) {
	// At VSHN, currently only full-day absences are possible, so no need to check for starting and ending time.
	s.Absences = append(s.Absences, block)
}

// CalculateOvertimeSummary returns the duration of overtime.
// If returned duration is positive, then the employee did overtime and undertime if duration is negative.
//
// The overtime is then calculated according to these business rules:
//   - Outside office hours are multiplied by 1.5 (as a compensation)
//   - Excused hours like sick leave, authorities or public service can be used to "fill up" the daily theoretical maximum if the working hours are less than said maximum.
//     However, there's no overtime possible using excused hours
//   - If the working hours exceed the theoretical daily maximum, then the excused hours are basically ignored.
//     Example: it's not possible to work 9 hours, have 1 hour sick leave and expect 2 hours overtime for an 8 hours daily maximum, the overtime here is 1 hour.
func (s *DailySummary) CalculateOvertimeSummary() OvertimeSummary {
	os := OvertimeSummary{}
	dailyMax := s.calculateDailyMax() - s.CalculateAbsenceTime()
	os.DailyMax = dailyMax
	s.calculateWorkingTime(&os)
	s.calculateExcusedTime(&os)
	return os
}

// Overtime returns the total overtime with all business rules respected.
func (s OvertimeSummary) Overtime() time.Duration {
	excusedTime := s.ExcusedTime()
	workingTime := s.WorkingTime()
	if workingTime >= s.DailyMax {
		// Can't be on excused hours. if working overtime.
		excusedTime = 0
	} else if workingTime+excusedTime > s.DailyMax {
		// There is overlap: Not enough workHours, but having excused hours = Cap at daily max, no overtime
		excusedTime = s.DailyMax - workingTime
	}
	return workingTime + excusedTime - s.DailyMax
}

// WorkingTime is the sum of OutOfOfficeTime (multiplied with 1.5) and RegularWorkingTime.
func (s OvertimeSummary) WorkingTime() time.Duration {
	overtime := 1.5 * float64(s.OutOfOfficeTime)
	return s.RegularWorkingTime + time.Duration(overtime)
}

// ExcusedTime returns the sum of AuthoritiesTime, PublicServiceTime and SickLeaveTime, but it can't exceed DailyMax.
func (s OvertimeSummary) ExcusedTime() time.Duration {
	sum := s.AuthoritiesTime + s.PublicServiceTime + s.SickLeaveTime
	if sum >= s.DailyMax {
		return s.DailyMax
	}
	return sum
}

// calculateDailyMax returns the theoretical amount of hours that an employee should work on this day.
//   - It returns 0 for weekend days.
//   - It returns 8.5 hours multiplied by FTE ratio for days in 2020 and earlier.
//   - It returns 8.0 hours multiplied by FTE ratio for days in 2021 and later.
func (s *DailySummary) calculateDailyMax() time.Duration {
	if s.IsWeekend() {
		return 0
	}
	if s.Date.Year() < 2021 {
		// VSHN switched from 42h-a-week to 40h-a-week on 1st of January 2021.
		return time.Duration(8.5 * s.FTERatio * float64(time.Hour))
	}
	return time.Duration(8 * s.FTERatio * float64(time.Hour))
}

// calculateWorkingTime accumulates all working hours from that day.
func (s *DailySummary) calculateWorkingTime(o *OvertimeSummary) {
	for _, shift := range s.Shifts {
		diff := shift.End.Sub(shift.Start)
		switch shift.Reason {
		case "":
			o.RegularWorkingTime += diff
		case ReasonOutsideOfficeHours:
			o.OutOfOfficeTime += diff
		}
	}
}

// calculateExcusedTime accumulates all hours that are excused in some way (sick leave etc) from that day.
func (s *DailySummary) calculateExcusedTime(o *OvertimeSummary) {
	for _, shift := range s.Shifts {
		diff := shift.End.Sub(shift.Start)
		switch shift.Reason {
		case ReasonSickLeave:
			o.SickLeaveTime += diff
		case ReasonAuthorities:
			o.AuthoritiesTime += diff
		case ReasonPublicService:
			o.PublicServiceTime += diff
		}
	}
}

// CalculateAbsenceTime accumulates all absence hours from that day.
func (s *DailySummary) CalculateAbsenceTime() time.Duration {
	total := time.Duration(0)
	for _, absence := range s.Absences {
		if absence.Reason != TypeUnpaid {
			// VSHN specific: Odoo treats "Unpaid" as normal leave, but for VSHN it's informational-only, meaning one still has to work.
			// For every other type of absence, we add the daily max equivalent.

			total += s.calculateDailyMax()
		}
	}
	return total
}

// HasAbsences returns true if there are any absences.
func (s *DailySummary) HasAbsences() bool {
	return len(s.Absences) != 0
}

// IsHoliday returns true if there is a "personalized" leave.
// Public and unpaid holidays return false.
// If the holiday falls on a weekend, the day is not counted.
func (s *DailySummary) IsHoliday() bool {
	for _, absence := range s.Absences {
		if absence.Reason != TypeUnpaid && absence.Reason != TypePublicHoliday {
			return !s.IsWeekend()
		}
	}
	return false
}

// IsWeekend returns true if the date falls on a Saturday or Sunday.
func (s *DailySummary) IsWeekend() bool {
	return s.Date.Weekday() == time.Saturday || s.Date.Weekday() == time.Sunday
}

func findDailySummaryByDate(dailies []*DailySummary, date time.Time) (*DailySummary, bool) {
	for _, daily := range dailies {
		if daily.Date.Day() == date.Day() && daily.Date.Month() == date.Month() && daily.Date.Year() == date.Year() {
			return daily, true
		}
	}
	return nil, false
}
