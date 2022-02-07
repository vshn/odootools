package web

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vshn/odootools/pkg/odoo"
)

func TestRenderLoginForm(t *testing.T) {
	req := httptest.NewRequest("GET", "/login", nil)
	res := httptest.NewRecorder()
	newServer("").ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code, "status code")
	assert.Equal(t, "text/html; charset=UTF-8", res.Header().Get("content-type"), "content-type")
	body, err := ioutil.ReadAll(res.Body)
	assert.NoError(t, err)
	assert.Contains(t, string(body), "<h1>Login</h1>")
	assert.Contains(t, string(body), "<title>Login ")
}

func TestLoginSuccess(t *testing.T) {
	var (
		numRequests  = 0
		testLogin    = "username"
		testPassword = "password"
	)

	odooMock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		numRequests++
		switch numRequests {
		case 1:
			handleLogin(t, w, r)
		case 2:
			handleEmployeeSearch(t, w, r)
		default:
			t.Fail()
		}
	}))

	// Prepare request
	form := url.Values{}
	form.Set("login", testLogin)
	form.Set("password", testPassword)
	req := httptest.NewRequest("POST", "/login", strings.NewReader(form.Encode()))
	req.Header.Set("content-type", "application/x-www-form-urlencoded")

	res := httptest.NewRecorder()
	newServer(odooMock.URL).ServeHTTP(res, req)

	assert.Equal(t, http.StatusFound, res.Code, "http status")
	assert.Equal(t, "/report", res.Header().Get("Location"), "location header")

	require.Len(t, res.Result().Cookies(), 1, "number of cookies")
	c := res.Result().Cookies()[0]
	assert.Equal(t, "odootools", c.Name, "cookie name")
	assert.NotContains(t, c.Value, testLogin, "no password in cookie")
	assert.True(t, c.HttpOnly, "cookie httpOnly flag")
	assert.True(t, c.Secure, "cookie secure flag")
	assert.Equal(t, 3, numRequests, "number of requests")
}

func handleEmployeeSearch(t *testing.T, w http.ResponseWriter, r *http.Request) {
	assert.Equal(t, "/web/dataset/search_read", r.RequestURI)

	b, err := ioutil.ReadAll(r.Body)
	require.NoError(t, err)
	body := string(b)

	t.Log(body)
	assert.Contains(t, body, ``)
}

func handleLogin(t *testing.T, w http.ResponseWriter, r *http.Request) {
	assert.Equal(t, "/web/session/authenticate", r.RequestURI)
	b, err := ioutil.ReadAll(r.Body)
	require.NoError(t, err)
	body := string(b)

	t.Log(body)
	assert.Contains(t, body, `"db":"TestDB"`)
	assert.Contains(t, body, `"login":"username"`)
	assert.Contains(t, body, `"password":"password"`)

	w.Header().Set("content-type", "application/json")
	_, err = w.Write([]byte(`{
			"id": "1337",
			"jsonrpc": "2.0",
			"result": {
				"company_id": 1,
				"db": "TestDB",
				"session_id": "sid",
				"uid": 1,
				"user_context": {
					"lang": "en_US",
					"tz": "Europe/Zurich",
					"uid": 1
				},
				"username": "username"
			}
		}`))
	assert.NoError(t, err)
}

func TestLoginBadCredentials(t *testing.T) {
	var (
		numRequests = 0
	)

	odooMock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		numRequests++
		assert.Equal(t, "/web/session/authenticate", r.RequestURI, "request URI")
		w.Header().Set("content-type", "application/json")
		_, err := w.Write([]byte(`{
			"id": "1337",
			"jsonrpc": "2.0",
			"result": {
				"company_id": null,
				"db": "TestDB",
				"session_id": "sid",
				"uid": false,
				"user_context": {},
				"username": "username"
			}
		}`))
		assert.NoError(t, err)
	}))

	// Prepare request
	form := url.Values{}
	form.Set("login", "some username")
	form.Set("password", "bad password")
	req := httptest.NewRequest("POST", "/login", strings.NewReader(form.Encode()))
	req.Header.Set("content-type", "application/x-www-form-urlencoded")

	// Do request
	res := httptest.NewRecorder()
	newServer(odooMock.URL).ServeHTTP(res, req)

	// Verify that login failed
	assert.Equal(t, http.StatusOK, res.Code, "http status code")
	assert.Equal(t, "", res.Header().Get("Location"), "location header")
	assert.Len(t, res.Result().Cookies(), 0, "number of cookies")
	assert.Equal(t, 1, numRequests, "number of requests")

	// Verify that the login page is rendered
	assert.Equal(t, "text/html; charset=UTF-8", res.Header().Get("content-type"), "content type")
	body, err := ioutil.ReadAll(res.Body)
	assert.NoError(t, err)
	assert.Contains(t, string(body), "<h1>Login</h1>")
	assert.Contains(t, string(body), "Invalid login or password")
}

func TestLoginBadResponse(t *testing.T) {
	numRequests := 0

	odooMock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		numRequests++

		w.Header().Set("content-type", "application/json")
		_, err := w.Write([]byte(`{
			"jsonrpc": "2.0",
			"id": "xxx",
			"error": {
			  "message": "Odoo Server Error",
			  "code": 200,
			  "data": {
				"debug": "Traceback xxx",
				"message": "",
				"name": "werkzeug.exceptions.Foo",
				"arguments": []
			  }
			}
		  }`))
		assert.NoError(t, err)
	}))

	form := url.Values{}
	form.Set("login", "a")
	form.Set("password", "a")
	req := httptest.NewRequest("POST", "/login", strings.NewReader(form.Encode()))
	req.Header.Set("content-type", "application/x-www-form-urlencoded")
	res := httptest.NewRecorder()
	newServer(odooMock.URL).ServeHTTP(res, req)

	assert.Equal(t, 1, numRequests, "number of requests")
	assert.Equal(t, http.StatusBadGateway, res.Code, "http status code")
}

func TestLogout(t *testing.T) {
	req := httptest.NewRequest("GET", "/logout", nil)
	req.AddCookie(&http.Cookie{Name: SessionCookieID, Value: "something"})
	res := httptest.NewRecorder()
	newServer("").ServeHTTP(res, req)

	assert.Equal(t, http.StatusTemporaryRedirect, res.Code, "http status code")
	assert.Equal(t, "/login", res.Header().Get("Location"), "location header")
	require.Len(t, res.Result().Cookies(), 2, "number of cookies")

	c := res.Result().Cookies()[0]
	assert.Equal(t, SessionCookieID, c.Name, "cookie name")
	assert.Equal(t, -1, c.MaxAge, "cookie age reset")
}

func newServer(odooURL string) *Server {
	var oc *odoo.Client
	if odooURL != "" {
		c, err := odoo.NewClient(odooURL, odoo.ClientOptions{})
		if err != nil {
			panic(err)
		}
		oc = c
	}
	return NewServer(oc, "0000000000000000000000000000000000000000000=", "TestDB")
}
