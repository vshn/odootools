package employeereport

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vshn/odootools/pkg/odoo/model"
)

func Test_EmployeeReport_ignoreNoContractFound(t *testing.T) {
	tests := map[string]struct {
		givenError    error
		expectedError string
	}{
		"GivenNilError_ThenExpectNil": {
			givenError: nil, expectedError: "",
		},
		"GivenAnyError_ThenExpectSameError": {
			givenError:    errors.New("test"),
			expectedError: "test",
		},
		"GivenNoContractCoversDateError_ThenExpectNil": {
			givenError:    &model.NoContractCoversDateErr{Err: errors.New("no contract")},
			expectedError: "",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			r := &EmployeeReport{}
			err := r.ignoreNoContractFound(nil, tc.givenError)
			if tc.expectedError != "" {
				assert.EqualError(t, err, tc.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
