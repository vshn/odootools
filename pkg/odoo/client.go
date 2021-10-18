package odoo

import (
	"bytes"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

// Client is the base struct that holds information required to talk to Odoo
type Client struct {
	baseURL string
	db      string
	http    *http.Client
}

// NewClient returns a new client with its basic fields set
func NewClient(baseURL, db string) *Client {
	transport := http.DefaultTransport
	if os.Getenv("DEBUG") != "" {
		transport = newDebugTransport()
	}
	return &Client{strings.TrimSuffix(baseURL, "/"), db, &http.Client{
		Timeout:   10 * time.Second,
		Jar:       nil, // don't save any cookies!
		Transport: transport,
	}}
}

type debugTransport struct {
	pwRe *regexp.Regexp
}

func newDebugTransport() *debugTransport {
	return &debugTransport{
		pwRe: regexp.MustCompile(`("password":\s?").+("[,}])`),
	}
}

func (t *debugTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	if r.Body != nil {
		reqBody, _ := r.GetBody()
		defer reqBody.Close()
		buf, _ := ioutil.ReadAll(reqBody)
		buf = t.pwRe.ReplaceAll(buf, []byte(`$1[confidential]$2`))
		log.Printf("%s %s ---> %s", r.Method, r.URL.Path, string(buf))
	}

	res, err := http.DefaultTransport.RoundTrip(r)

	if res.Body != nil {
		defer res.Body.Close()
		buf, _ := ioutil.ReadAll(res.Body)
		log.Print("<--- ", string(buf))
		res.Body = io.NopCloser(bytes.NewReader(buf))
	}

	return res, err
}
