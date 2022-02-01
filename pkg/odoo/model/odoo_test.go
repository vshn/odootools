package model

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/vshn/odootools/pkg/odoo"
)

func newDateTime(t *testing.T, value string) *odoo.Date {
	tm, err := time.Parse(odoo.DateTimeFormat, fmt.Sprintf("%s:00", value))
	require.NoError(t, err)
	ptr := odoo.Date(tm)
	return &ptr
}

func newDate(t *testing.T, value string) *odoo.Date {
	tm, err := time.Parse(odoo.DateFormat, value)
	require.NoError(t, err)
	ptr := odoo.Date(tm)
	return &ptr
}
