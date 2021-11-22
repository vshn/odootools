package odoo

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
		givenDay      Date
		expectedRatio float64
		expectErr     bool
	}{
		"GivenEmptyList_WhenNil_ThenReturnErr": {
			givenList: nil,
			expectErr: true,
		},
		"GivenEmptyList_WhenNoContracts_ThenReturnErr": {
			givenList: []Contract{},
			expectErr: true,
		},
		"GivenListWith1Contract_WhenOpenEnd_ThenReturnRatio": {
			givenDay: *newDate(t, "2021-12-04"),
			givenList: []Contract{
				{Start: newDate(t, "2021-02-01"), WorkingSchedule: newFTESchedule(100)},
			},
			expectedRatio: 1,
		},
		"GivenListWith1Contract_WhenDayBeforeStart_ThenReturnErr": {
			givenDay: *newDate(t, "2021-02-01"),
			givenList: []Contract{
				{Start: newDate(t, "2021-02-02"), WorkingSchedule: newFTESchedule(100)},
			},
			expectErr: true,
		},
		"GivenListWith2Contract_WhenDayBetweenContract_ThenReturnRatioFromTerminatedContract": {
			givenDay: *newDate(t, "2021-03-31"),
			givenList: []Contract{
				{Start: newDate(t, "2021-02-02"), End: newDate(t, "2021-03-31"), WorkingSchedule: newFTESchedule(90)},
				{Start: newDate(t, "2021-04-01"), WorkingSchedule: newFTESchedule(80)},
			},
			expectedRatio: 0.9,
		},
		"GivenListWith2Contract_WhenDayInOpenContract_ThenReturnRatioFromOpenContract": {
			givenDay: *newDate(t, "2021-04-01"),
			givenList: []Contract{
				{Start: newDate(t, "2021-02-02"), End: newDate(t, "2021-03-31"), WorkingSchedule: newFTESchedule(90)},
				{Start: newDate(t, "2021-04-01"), WorkingSchedule: newFTESchedule(80)},
			},
			expectedRatio: 0.8,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			result, err := tt.givenList.GetFTERatioForDay(tt.givenDay)
			if tt.expectErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.expectedRatio, result)
		})
	}
}
