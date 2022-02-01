package integration_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDashboardUnauthenticated(t *testing.T) {
	req := httptest.NewRequest("GET", "/report", nil)
	res := httptest.NewRecorder()
	newServer("").ServeHTTP(res, req)

	assert.Equal(t, http.StatusTemporaryRedirect, res.Code)
	assert.Equal(t, "/login", res.Header().Get("Location"))
}
