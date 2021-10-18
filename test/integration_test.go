package integration_test

import (
	"io/ioutil"
	"net/http/httptest"
	"testing"

	"github.com/matryer/is"
	"github.com/mhutter/vshn-ftb/pkg/web"
)

func newServer() *web.Server {
	return web.NewServer("../templates")
}

func TestHealthz(t *testing.T) {
	is := is.New(t)

	req := httptest.NewRequest("GET", "/healthz", nil)
	res := httptest.NewRecorder()
	newServer().ServeHTTP(res, req)

	is.Equal(200, res.Code)
	body, err := ioutil.ReadAll(res.Body)
	is.NoErr(err)
	is.Equal(body, []byte{}) // response body

}
