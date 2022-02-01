package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestActionReason_MarshalJSON(t *testing.T) {
	tests := map[string]struct {
		givenInput     ActionReason
		expectedOutput string
	}{
		"GivenInput_WhenEmpty_ThenReturnFalse": {
			givenInput:     ActionReason{},
			expectedOutput: "false",
		},
		"GivenInput_WhenInputArray_ThenReturnString": {
			givenInput: ActionReason{
				ID:   4,
				Name: "Sick / Medical Consultation",
			},
			expectedOutput: "[4,\"Sick / Medical Consultation\"]",
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

func TestActionReason_UnmarshalJSON(t *testing.T) {
	tests := map[string]struct {
		givenInput     []byte
		expectedOutput string
	}{
		"GivenInput_WhenFalse_ThenReturnEmpty": {
			givenInput:     []byte("false"),
			expectedOutput: "",
		},
		"GivenInput_WhenInputArray_ThenReturnString": {
			givenInput:     []byte("[4, \"Sick / Medical Consultation\"]"),
			expectedOutput: "Sick / Medical Consultation",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			actionReason := new(ActionReason)

			err := actionReason.UnmarshalJSON(tt.givenInput)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedOutput, actionReason.String())
		})
	}
}
