package timesheet

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/vshn/odootools/pkg/odoo"
	"github.com/vshn/odootools/pkg/odoo/model"
)

func TestReportBuilder_getEarliestStartContractDate(t *testing.T) {
	tests := map[string]struct {
		givenContracts model.ContractList
		expectedDate   time.Time
		expectedFound  bool
	}{
		"GivenNoContracts_ThenReturnFalse": {
			givenContracts: model.ContractList{},
			expectedFound:  false,
		},
		"GivenContracts_WhenStartDateExists_ThenReturnTrue": {
			givenContracts: model.ContractList{Items: []model.Contract{
				{Start: newDateTime(t, "2021-02-04 08:00")},
			}},
			expectedDate:  newDateTime(t, "2021-02-04 08:00").ToTime(),
			expectedFound: true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			r := NewYearlyReporter(model.AttendanceList{}, odoo.List[model.Leave]{}, nil, tt.givenContracts)
			resultDate, found := r.getEarliestStartContractDate()
			assert.Equal(t, tt.expectedFound, found)
			if tt.expectedFound {
				assert.Equal(t, tt.expectedDate, resultDate)
			}
		})
	}
}
