package integration_test

import (
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/matryer/is"
	"github.com/mhutter/vshn-ftb/pkg/odoo"
	"github.com/mhutter/vshn-ftb/pkg/web"
)

func TestMain(m *testing.M) {
	rand.Seed(time.Now().UnixNano())
	log.SetOutput(io.Discard)
	web.RootDir = ".."

	os.Exit(m.Run())
}

func TestHealthz(t *testing.T) {
	is := is.New(t)

	req := httptest.NewRequest("GET", "/healthz", nil)
	res := httptest.NewRecorder()
	newServer("").ServeHTTP(res, req)

	is.Equal(200, res.Code)
	body, err := ioutil.ReadAll(res.Body)
	is.NoErr(err)
	is.Equal(body, []byte{}) // response body
}

func newServer(odooURL string) *web.Server {
	var oc *odoo.Client
	if odooURL != "" {
		oc = odoo.NewClient(odooURL, "TestDB")
	}
	return web.NewServer(oc, "0000000000000000000000000000000000000000000=")
}
