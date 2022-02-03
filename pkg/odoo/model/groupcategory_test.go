package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGroupCategory_MarshalJSON(t *testing.T) {
	tests := map[string]struct {
		givenInput     GroupCategory
		expectedOutput string
	}{
		"GivenInput_WhenEmpty_ThenReturnFalse": {
			givenInput:     GroupCategory{},
			expectedOutput: `[0,""]`,
		},
		"GivenInput_WhenInputArray_ThenReturnString": {
			givenInput: GroupCategory{
				ID:   19,
				Name: "Human Resources",
			},
			expectedOutput: `[19,"Human Resources"]`,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {

			result, err := tt.givenInput.MarshalJSON()
			require.NoError(t, err)
			assert.Equal(t, tt.expectedOutput, string(result))
		})
	}
}

func TestGroupCategory_UnmarshalJSON(t *testing.T) {
	tests := map[string]struct {
		givenInput     []byte
		expectedResult GroupCategory
		expectedError  string
	}{
		"GivenFalse_ThenExpectNil": {
			givenInput: []byte("false"),
		},
		"GivenCategory_ThenExpectCorrectValues": {
			givenInput: []byte("[ 19, \"Human Resources\" ]"),
			expectedResult: GroupCategory{
				ID:   19,
				Name: "Human Resources",
			},
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := new(GroupCategory)
			err := result.UnmarshalJSON(tc.givenInput)
			if tc.expectedError != "" {
				assert.EqualError(t, err, tc.expectedError)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.expectedResult, *result)
		})
	}
}
