package model

import (
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/vshn/odootools/pkg/odoo"
)

func TestAttendanceList_Sort(t *testing.T) {
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
					{DateTime: odoo.MustParseDateTime("2022-03-29 15:37:46")},
					{DateTime: odoo.MustParseDateTime("2022-03-28 15:37:46")},
					{DateTime: odoo.MustParseDateTime("2022-03-27 15:37:46")},
					{DateTime: odoo.MustParseDateTime("2022-03-26 15:37:46")},
				},
			},
			expectedOrder: []Attendance{
				{DateTime: odoo.MustParseDateTime("2022-03-26 15:37:46")},
				{DateTime: odoo.MustParseDateTime("2022-03-27 15:37:46")},
				{DateTime: odoo.MustParseDateTime("2022-03-28 15:37:46")},
				{DateTime: odoo.MustParseDateTime("2022-03-29 15:37:46")},
				{DateTime: odoo.MustParseDateTime("2022-03-30 15:37:46")},
				{DateTime: odoo.MustParseDateTime("2022-03-31 15:37:46")},
			},
		},
		"GivenListWithElements_WhenSameDate_ThenSortByReason": {
			givenList: AttendanceList{
				Items: []Attendance{
					{DateTime: odoo.MustParseDateTime("2022-03-31 08:00:00"), Reason: &ActionReason{Name: ""}, Action: ActionSignIn},
					{DateTime: odoo.MustParseDateTime("2022-03-31 10:00:00"), Reason: &ActionReason{Name: "AReason"}, Action: ActionSignIn},
					{DateTime: odoo.MustParseDateTime("2022-03-31 16:00:00"), Reason: &ActionReason{Name: ""}, Action: ActionSignIn},
					{DateTime: odoo.MustParseDateTime("2022-03-31 11:00:00"), Reason: &ActionReason{Name: "BReason"}, Action: ActionSignIn},
					{DateTime: odoo.MustParseDateTime("2022-03-31 11:00:00"), Reason: &ActionReason{Name: "AReason"}, Action: ActionSignOut},
					{DateTime: odoo.MustParseDateTime("2022-03-31 10:00:00"), Reason: &ActionReason{Name: ""}, Action: ActionSignOut},
					{DateTime: odoo.MustParseDateTime("2022-03-31 12:00:00"), Reason: &ActionReason{Name: ""}, Action: ActionSignIn},
					{DateTime: odoo.MustParseDateTime("2022-03-31 14:00:00"), Reason: &ActionReason{Name: ""}, Action: ActionSignIn},
					{DateTime: odoo.MustParseDateTime("2022-03-31 12:00:00"), Reason: &ActionReason{Name: "BReason"}, Action: ActionSignOut},
					{DateTime: odoo.MustParseDateTime("2022-04-01 00:00:00"), Reason: &ActionReason{Name: ""}, Action: ActionSignIn},
					{DateTime: odoo.MustParseDateTime("2022-04-01 00:00:00"), Reason: &ActionReason{Name: ""}, Action: ActionSignOut},
				},
			},
			expectedOrder: []Attendance{
				{DateTime: odoo.MustParseDateTime("2022-03-31 08:00:00"), Reason: &ActionReason{Name: ""}, Action: ActionSignIn},
				{DateTime: odoo.MustParseDateTime("2022-03-31 10:00:00"), Reason: &ActionReason{Name: ""}, Action: ActionSignOut},
				{DateTime: odoo.MustParseDateTime("2022-03-31 10:00:00"), Reason: &ActionReason{Name: "AReason"}, Action: ActionSignIn},
				{DateTime: odoo.MustParseDateTime("2022-03-31 11:00:00"), Reason: &ActionReason{Name: "AReason"}, Action: ActionSignOut},
				{DateTime: odoo.MustParseDateTime("2022-03-31 11:00:00"), Reason: &ActionReason{Name: "BReason"}, Action: ActionSignIn},
				{DateTime: odoo.MustParseDateTime("2022-03-31 12:00:00"), Reason: &ActionReason{Name: "BReason"}, Action: ActionSignOut},
				{DateTime: odoo.MustParseDateTime("2022-03-31 12:00:00"), Reason: &ActionReason{Name: ""}, Action: ActionSignIn},
				{DateTime: odoo.MustParseDateTime("2022-03-31 14:00:00"), Reason: &ActionReason{Name: ""}, Action: ActionSignIn},
				{DateTime: odoo.MustParseDateTime("2022-03-31 16:00:00"), Reason: &ActionReason{Name: ""}, Action: ActionSignIn},
				{DateTime: odoo.MustParseDateTime("2022-04-01 00:00:00"), Reason: &ActionReason{Name: ""}, Action: ActionSignOut},
				{DateTime: odoo.MustParseDateTime("2022-04-01 00:00:00"), Reason: &ActionReason{Name: ""}, Action: ActionSignIn},
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
			// randomize each list a few times, should always return same stable order.
			for i := 0; i < 5; i++ {
				rand.Seed(time.Now().UnixNano())
				rand.Shuffle(len(tc.givenList.Items), func(i, j int) {
					tc.givenList.Items[i], tc.givenList.Items[j] = tc.givenList.Items[j], tc.givenList.Items[i]
				})

				tc.givenList.Sort()
				assert.Equalf(t, tc.expectedOrder, tc.givenList.Items, "attempt %d", i)
			}
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
