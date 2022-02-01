package model

import (
	"strconv"
	"testing"

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
		givenDay      odoo.Date
		expectedRatio float64
		expectedError string
	}{
		"GivenEmptyList_WhenNil_ThenReturnErr": {
			givenList:     ContractList{},
			expectedError: "no contract found that covers date: 0001-01-01 00:00:00",
		},
		"GivenEmptyList_WhenNoContracts_ThenReturnErr": {
			givenList:     ContractList{Items: []Contract{}},
			expectedError: "no contract found that covers date: 0001-01-01 00:00:00",
		},
		"GivenListWith1Contract_WhenOpenEnd_ThenReturnRatio": {
			givenDay: *newDate(t, "2021-12-04"),
			givenList: ContractList{Items: []Contract{
				{Start: newDate(t, "2021-02-01"), WorkingSchedule: newFTESchedule(100)},
			}},
			expectedRatio: 1,
		},
		"GivenListWith1Contract_WhenDayBeforeStart_ThenReturnErr": {
			givenDay: *newDate(t, "2021-02-01"),
			givenList: ContractList{Items: []Contract{
				{Start: newDate(t, "2021-02-02"), WorkingSchedule: newFTESchedule(100)},
			}},
			expectedError: "no contract found that covers date: 2021-02-01 00:00:00",
		},
		"GivenListWith2Contract_WhenDayBetweenContract_ThenReturnRatioFromTerminatedContract": {
			givenDay: *newDate(t, "2021-03-31"),
			givenList: ContractList{Items: []Contract{
				{Start: newDate(t, "2021-02-02"), End: newDate(t, "2021-03-31"), WorkingSchedule: newFTESchedule(90)},
				{Start: newDate(t, "2021-04-01"), WorkingSchedule: newFTESchedule(80)},
			}},
			expectedRatio: 0.9,
		},
		"GivenListWith2Contract_WhenDayInOpenContract_ThenReturnRatioFromOpenContract": {
			givenDay: *newDate(t, "2021-04-01"),
			givenList: ContractList{Items: []Contract{
				{Start: newDate(t, "2021-02-02"), End: newDate(t, "2021-03-31"), WorkingSchedule: newFTESchedule(90)},
				{Start: newDate(t, "2021-04-01"), WorkingSchedule: newFTESchedule(80)},
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
