package integration_test

import (
	"io/ioutil"
	"net/http/httptest"
	"testing"

	"github.com/matryer/is"
)

func TestStaticAssets(t *testing.T) {
	is := is.New(t)

	req := httptest.NewRequest("GET", "/robots.txt", nil)
	res := httptest.NewRecorder()
	newServer("").ServeHTTP(res, req)

	is.Equal(200, res.Code)
	is.Equal(res.Header().Get("content-type"), "text/plain; charset=utf-8") // Content-Type
	body, err := ioutil.ReadAll(res.Body)
	is.NoErr(err)
	is.Equal(string(body), "User-agent: *\nDisallow: /\n")
}
