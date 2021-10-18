package integration_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/matryer/is"
)

func TestDashboardUnauthenticated(t *testing.T) {
	is := is.New(t)

	req := httptest.NewRequest("GET", "/", nil)
	res := httptest.NewRecorder()
	newServer("").ServeHTTP(res, req)

	is.Equal(res.Code, http.StatusTemporaryRedirect) // HTTP status code
	is.Equal(res.Header().Get("Location"), "/login") // Location header
}
