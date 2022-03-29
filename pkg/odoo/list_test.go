package odoo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestList_Len(t *testing.T) {
	tests := map[string]struct {
		givenList   *List[interface{}]
		expectedLen int
	}{
		"GivenNilList_ThenExpectZero": {
			givenList:   nil,
			expectedLen: 0,
		},
		"GivenNilItems_ThenExpectZero": {
			givenList:   &List[interface{}]{Items: nil},
			expectedLen: 0,
		},
		"GivenEmptyItems_ThenExpectZero": {
			givenList:   &List[interface{}]{Items: []interface{}{}},
			expectedLen: 0,
		},
		"GivenSingleItem_ThenExpectOne": {
			givenList:   &List[interface{}]{Items: []interface{}{1}},
			expectedLen: 1,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := tc.givenList.Len()
			assert.Equal(t, tc.expectedLen, result)
		})
	}
}
