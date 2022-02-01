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

	"github.com/stretchr/testify/assert"
	"github.com/vshn/odootools/pkg/odoo"
	"github.com/vshn/odootools/pkg/web"
)

func TestMain(m *testing.M) {
	rand.Seed(time.Now().UnixNano())
	log.SetOutput(io.Discard)

	os.Exit(m.Run())
}

func TestHealthz(t *testing.T) {
	req := httptest.NewRequest("GET", "/healthz", nil)
	res := httptest.NewRecorder()
	newServer("").ServeHTTP(res, req)

	assert.Equal(t, res.Code, 200)
	body, err := ioutil.ReadAll(res.Body)
	assert.NoError(t, err)
	assert.Equal(t, []byte{}, body) // response body
}

func newServer(odooURL string) *web.Server {
	var oc *odoo.Client
	if odooURL != "" {
		c, err := odoo.NewClient(odooURL, odoo.ClientOptions{})
		if err != nil {
			panic(err)
		}
		oc = c
	}
	return web.NewServer(oc, "0000000000000000000000000000000000000000000=", "TestDB")
}
