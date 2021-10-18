package integration_test

import (
	"bytes"
	"io/ioutil"
	"net/http/httptest"
	"testing"

	"github.com/matryer/is"
)

func TestRenderLoginForm(t *testing.T) {
	is := is.New(t)

	req := httptest.NewRequest("GET", "/login", nil)
	res := httptest.NewRecorder()
	newServer().ServeHTTP(res, req)

	is.Equal(200, res.Code)
	is.Equal("text/html", res.Header().Get("content-type")) // Content-Type
	body, err := ioutil.ReadAll(res.Body)
	is.NoErr(err)
	bytes.Contains(body, []byte("<h1>Login</h1>"))
}
