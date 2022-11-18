package model

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/vshn/odootools/pkg/odoo"
)

func TestAttendanceList_SortByDate(t *testing.T) {
	tests := map[string]struct {
		givenList     AttendanceList
		expectedOrder []Attendance
	}{
		"GivenListWithElements_WhenAlreadyOrdered_ThenIgnoreSort": {
			givenList: AttendanceList{
				Items: []Attendance{
					{DateTime: odoo.MustParseDateTime("2022-03-30 15:37:46")},
					{DateTime: odoo.MustParseDateTime("2022-03-31 15:37:46")},
				},
			},
			expectedOrder: []Attendance{
				{DateTime: odoo.MustParseDateTime("2022-03-30 15:37:46")},
				{DateTime: odoo.MustParseDateTime("2022-03-31 15:37:46")},
			},
		},
		"GivenListWithElements_WhenUnordered_ThenSortByDateAscending": {
			givenList: AttendanceList{
				Items: []Attendance{
					{DateTime: odoo.MustParseDateTime("2022-03-31 15:37:46")},
					{DateTime: odoo.MustParseDateTime("2022-03-30 15:37:46")},
				},
			},
			expectedOrder: []Attendance{
				{DateTime: odoo.MustParseDateTime("2022-03-30 15:37:46")},
				{DateTime: odoo.MustParseDateTime("2022-03-31 15:37:46")},
			},
		},
		"GivenNilList_ThenDoNothing": {
			givenList: AttendanceList{
				Items: nil,
			},
			expectedOrder: nil,
		},
		"GivenEmptyList_ThenDoNothing": {
			givenList: AttendanceList{
				Items: []Attendance{},
			},
			expectedOrder: []Attendance{},
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			tc.givenList.SortByDate()
			assert.Equal(t, tc.expectedOrder, tc.givenList.Items)
		})
	}
}

func TestAttendanceList_FilterAttendanceBetweenDates(t *testing.T) {
	tests := map[string]struct {
		givenList    AttendanceList
		givenFrom    time.Time
		givenTo      time.Time
		expectedList AttendanceList
	}{
		"GivenNilList_ThenExpectNilList": {},
		"GivenEmptyList_ThenExpectEmptyList": {
			givenList:    AttendanceList{Items: []Attendance{}},
			expectedList: AttendanceList{Items: []Attendance{}},
		},
		"GivenListWithOneEntry_WhenDateBetween_ThenExpectEntry": {
			givenList: AttendanceList{Items: []Attendance{
				{DateTime: odoo.MustParseDateTime("2022-03-30 15:00:00")},
			}},
			givenFrom: odoo.MustParseDateTime("2022-03-30 12:00:00").Time,
			givenTo:   odoo.MustParseDateTime("2022-03-30 17:00:00").Time,
			expectedList: AttendanceList{Items: []Attendance{
				{DateTime: odoo.MustParseDateTime("2022-03-30 15:00:00")},
			}},
		},
		"GivenListWithOneEntry_WhenDateEqualWithFrom_ThenExpectEntry": {
			givenList: AttendanceList{Items: []Attendance{
				{DateTime: odoo.MustParseDateTime("2022-03-30 15:00:00")},
			}},
			givenFrom: odoo.MustParseDateTime("2022-03-30 15:00:00").Time,
			givenTo:   odoo.MustParseDateTime("2022-03-30 17:00:00").Time,
			expectedList: AttendanceList{Items: []Attendance{
				{DateTime: odoo.MustParseDateTime("2022-03-30 15:00:00")},
			}},
		},
		"GivenListWithOneEntry_WhenDateEqualWithTo_ThenExpectEntry": {
			givenList: AttendanceList{Items: []Attendance{
				{DateTime: odoo.MustParseDateTime("2022-03-30 15:00:00")},
			}},
			givenFrom: odoo.MustParseDateTime("2022-03-30 12:00:00").Time,
			givenTo:   odoo.MustParseDateTime("2022-03-30 15:00:00").Time,
			expectedList: AttendanceList{Items: []Attendance{
				{DateTime: odoo.MustParseDateTime("2022-03-30 15:00:00")},
			}},
		},
		"GivenListWithOneEntry_WhenDateOutsideRange_ThenExpectEmptyList": {
			givenList: AttendanceList{Items: []Attendance{
				{DateTime: odoo.MustParseDateTime("2022-03-31 15:00:00")},
			}},
			givenFrom:    odoo.MustParseDateTime("2022-03-30 12:00:00").Time,
			givenTo:      odoo.MustParseDateTime("2022-03-30 15:00:00").Time,
			expectedList: AttendanceList{Items: []Attendance{}},
		},
		"GivenListWithMultipleEntries_WhenOneOutsideRange_ThenExpectSingleEntry": {
			givenList: AttendanceList{Items: []Attendance{
				{DateTime: odoo.MustParseDateTime("2022-03-31 15:00:00")},
				{DateTime: odoo.MustParseDateTime("2022-03-30 14:00:00")},
			}},
			givenFrom: odoo.MustParseDateTime("2022-03-30 12:00:00").Time,
			givenTo:   odoo.MustParseDateTime("2022-03-30 15:00:00").Time,
			expectedList: AttendanceList{Items: []Attendance{
				{DateTime: odoo.MustParseDateTime("2022-03-30 14:00:00")},
			}},
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := tc.givenList.FilterAttendanceBetweenDates(tc.givenFrom, tc.givenTo)
			assert.Equal(t, tc.expectedList, result)
		})
	}
}
