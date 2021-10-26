package integration_test

import (
	"bytes"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/matryer/is"
	"github.com/vshn/odootools/pkg/web"
)

func TestRenderLoginForm(t *testing.T) {
	is := is.New(t)

	req := httptest.NewRequest("GET", "/login", nil)
	res := httptest.NewRecorder()
	newServer("").ServeHTTP(res, req)

	is.Equal(200, res.Code)
	is.Equal("text/html", res.Header().Get("content-type")) // Content-Type
	body, err := ioutil.ReadAll(res.Body)
	is.NoErr(err)
	is.True(bytes.Contains(body, []byte("<h1>Login</h1>")))
	is.True(bytes.Contains(body, []byte("<title>Login ")))
}

func TestLoginSuccess(t *testing.T) {
	var (
		is           = is.New(t)
		numRequests  = 0
		testLogin    = uuid.NewString()
		testPassword = uuid.NewString()
		testUID      = rand.Int()
		testSID      = uuid.NewString()
	)

	odooMock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		numRequests++

		is.Equal("/web/session/authenticate", r.RequestURI) // request URI
		b, err := ioutil.ReadAll(r.Body)
		is.NoErr(err)
		body := string(b)

		log.Println(body)
		is.True(strings.Contains(body, `"db":"TestDB"`))
		is.True(strings.Contains(body, `"login":"`+testLogin+`"`))
		is.True(strings.Contains(body, `"password":"`+testPassword+`"`))

		w.Header().Set("content-type", "application/json")
		_, err = w.Write([]byte(`{
			"id": "1337",
			"jsonrpc": "2.0",
			"result": {
				"company_id": 1,
				"db": "TestDB",
				"session_id": "` + testSID + `",
				"uid": ` + strconv.Itoa(testUID) + `,
				"user_context": {
					"lang": "en_US",
					"tz": "Europe/Zurich",
					"uid": ` + strconv.Itoa(testUID) + `
				},
				"username": "` + testLogin + `"
			}
		}`))
		is.NoErr(err)
	}))

	// Prepare request
	form := url.Values{}
	form.Set("login", testLogin)
	form.Set("password", testPassword)
	req := httptest.NewRequest("POST", "/login", strings.NewReader(form.Encode()))
	req.Header.Set("content-type", "application/x-www-form-urlencoded")

	res := httptest.NewRecorder()
	newServer(odooMock.URL).ServeHTTP(res, req)

	is.Equal(res.Code, http.StatusTemporaryRedirect)  // HTTP status
	is.Equal(res.Header().Get("Location"), "/report") // Location header

	is.Equal(1, len(res.Result().Cookies())) // number of cookies
	c := res.Result().Cookies()[0]
	is.Equal("ftb_sid", c.Name) // cookie name
	is.True(!strings.Contains(c.Value, testLogin))
	is.Equal(true, c.HttpOnly) // cookie HttpOnly
	is.Equal(true, c.Secure)   // cookie Secure

	is.Equal(numRequests, 1) // total number of requests
}

func TestLoginBadCredentials(t *testing.T) {
	var (
		is           = is.New(t)
		numRequests  = 0
		testLogin    = uuid.NewString()
		testPassword = uuid.NewString()
		testSID      = uuid.NewString()
	)

	odooMock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		numRequests++

		is.Equal("/web/session/authenticate", r.RequestURI) // request URI

		w.Header().Set("content-type", "application/json")
		_, err := w.Write([]byte(`{
			"id": "1337",
			"jsonrpc": "2.0",
			"result": {
				"company_id": null,
				"db": "TestDB",
				"session_id": "` + testSID + `",
				"uid": false,
				"user_context": {},
				"username": "` + testLogin + `"
			}
		}`))
		is.NoErr(err)
	}))

	// Prepare request
	form := url.Values{}
	form.Set("login", testLogin)
	form.Set("password", testPassword)
	req := httptest.NewRequest("POST", "/login", strings.NewReader(form.Encode()))
	req.Header.Set("content-type", "application/x-www-form-urlencoded")

	// Do request
	res := httptest.NewRecorder()
	newServer(odooMock.URL).ServeHTTP(res, req)

	// Verify that login failed
	is.Equal(res.Code, 200)                    // HTTP status
	is.Equal(res.Header().Get("Location"), "") // Location header
	is.Equal(0, len(res.Result().Cookies()))   // number of cookies
	is.Equal(numRequests, 1)                   // total number of requests

	// Verify that the login page is rendered
	is.Equal("text/html", res.Header().Get("content-type")) // Content-Type
	body, err := ioutil.ReadAll(res.Body)
	is.NoErr(err)
	is.True(bytes.Contains(body, []byte("<h1>Login</h1>")))
	is.True(bytes.Contains(body, []byte("Invalid login or password")))
}

func TestLoginBadResponse(t *testing.T) {
	is := is.New(t)
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
		is.NoErr(err)
	}))

	form := url.Values{}
	form.Set("login", "a")
	form.Set("password", "a")
	req := httptest.NewRequest("POST", "/login", strings.NewReader(form.Encode()))
	req.Header.Set("content-type", "application/x-www-form-urlencoded")
	res := httptest.NewRecorder()
	newServer(odooMock.URL).ServeHTTP(res, req)

	is.Equal(numRequests, 1) // total number of requests
	is.Equal(res.Code, http.StatusBadGateway)
}

func TestLogout(t *testing.T) {
	is := is.New(t)

	req := httptest.NewRequest("GET", "/logout", nil)
	req.AddCookie(&http.Cookie{Name: web.CookieSID, Value: "something"})
	res := httptest.NewRecorder()
	newServer("").ServeHTTP(res, req)

	is.Equal(307, res.Code)                          // http status code
	is.Equal("/login", res.Header().Get("Location")) // location header
	is.Equal(1, len(res.Result().Cookies()))         // cookies
	c := res.Result().Cookies()[0]
	is.Equal(web.CookieSID, c.Name)
	is.Equal(-1, c.MaxAge)
}
