package model

import (
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vshn/odootools/pkg/odoo"
)

func newFTESchedule(ratioPercentage int) *WorkingSchedule {
	return &WorkingSchedule{
		ID:   0,
		Name: strconv.Itoa(ratioPercentage) + "%",
	}
}

func TestContractList_GetFTERatioForDay(t *testing.T) {
	tests := map[string]struct {
		givenList     ContractList
		givenDay      time.Time
		expectedRatio float64
		expectedError string
	}{
		"GivenEmptyList_WhenNil_ThenReturnErr": {
			givenList:     ContractList{},
			expectedError: "no contract found that covers date: 0001-01-01 00:00:00 +0000 UTC",
		},
		"GivenEmptyList_WhenNoContracts_ThenReturnErr": {
			givenList:     ContractList{Items: []Contract{}},
			expectedError: "no contract found that covers date: 0001-01-01 00:00:00 +0000 UTC",
		},
		"GivenListWith1Contract_WhenOpenEnd_ThenReturnRatio": {
			givenDay: time.Date(2021, 12, 04, 0, 0, 0, 0, time.UTC),
			givenList: ContractList{Items: []Contract{
				{Start: odoo.NewDate(2021, 02, 01, 0, 0, 0, time.UTC), WorkingSchedule: newFTESchedule(100)},
			}},
			expectedRatio: 1,
		},
		"GivenListWith1Contract_WhenTimezoneGiven_ThenShouldCoverDate": {
			givenDay: time.Date(2021, 12, 04, 0, 0, 0, 0, vancouverTZ),
			givenList: ContractList{Items: []Contract{
				{Start: odoo.NewDate(2021, 12, 04, 0, 0, 0, time.UTC), WorkingSchedule: newFTESchedule(100)},
			}},
			expectedRatio: 1,
		},
		"GivenListWith1Contract_WhenTimezoneOutsideUTC_ThenReturnError": {
			givenDay: time.Date(2021, 12, 03, 0, 0, 0, 0, vancouverTZ),
			givenList: ContractList{Items: []Contract{
				{Start: odoo.NewDate(2021, 12, 04, 0, 0, 0, time.UTC), WorkingSchedule: newFTESchedule(100)},
			}},
			expectedError: "no contract found that covers date: 2021-12-03 00:00:00 -0800 PST",
		},
		"GivenListWith1Contract_WhenDayBeforeStart_ThenReturnErr": {
			givenDay: time.Date(2021, 02, 01, 0, 0, 0, 0, time.UTC),
			givenList: ContractList{Items: []Contract{
				{Start: odoo.NewDate(2021, 02, 02, 0, 0, 0, time.UTC), WorkingSchedule: newFTESchedule(100)},
			}},
			expectedError: "no contract found that covers date: 2021-02-01 00:00:00 +0000 UTC",
		},
		"GivenListWith2Contract_WhenDayBetweenContract_ThenReturnRatioFromTerminatedContract": {
			givenDay: time.Date(2021, 03, 31, 0, 0, 0, 0, time.UTC),
			givenList: ContractList{Items: []Contract{
				{Start: odoo.NewDate(2021, 02, 02, 0, 0, 0, time.UTC), End: odoo.NewDate(2021, 03, 31, 0, 0, 0, time.UTC), WorkingSchedule: newFTESchedule(90)},
				{Start: odoo.NewDate(2021, 04, 01, 0, 0, 0, time.UTC), WorkingSchedule: newFTESchedule(80)},
			}},
			expectedRatio: 0.9,
		},
		"GivenListWith2Contract_WhenDayInOpenContract_ThenReturnRatioFromOpenContract": {
			givenDay: time.Date(2021, 04, 01, 0, 0, 0, 0, time.UTC),
			givenList: ContractList{Items: []Contract{
				{Start: odoo.NewDate(2021, 02, 02, 0, 0, 0, time.UTC), End: odoo.NewDate(2021, 03, 31, 0, 0, 0, time.UTC), WorkingSchedule: newFTESchedule(90)},
				{Start: odoo.NewDate(2021, 04, 01, 0, 0, 0, time.UTC), WorkingSchedule: newFTESchedule(80)},
			}},
			expectedRatio: 0.8,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			result, err := tt.givenList.GetFTERatioForDay(tt.givenDay)
			if tt.expectedError != "" {
				require.EqualError(t, err, tt.expectedError)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.expectedRatio, result)
		})
	}
}
