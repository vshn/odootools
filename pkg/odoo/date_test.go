package odoo

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func newDate(t *testing.T, value string) *Date {
	tm, err := time.Parse(DateTimeFormat, fmt.Sprintf("%s:00", value))
	require.NoError(t, err)
	ptr := Date(tm)
	return &ptr
}
