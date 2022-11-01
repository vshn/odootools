package model

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/vshn/odootools/pkg/odoo"
)

func newDate(t *testing.T, value string) *odoo.Date {
	tm, err := time.Parse(odoo.DateFormat, value)
	require.NoError(t, err)
	ptr := odoo.Date(tm)
	return &ptr
}

func mustParseDate(t *testing.T, value string) time.Time {
	tm, err := time.Parse(odoo.DateTimeFormat, value)
	require.NoError(t, err)
	return tm
}
