package odoo

import (
	"bytes"
	"fmt"
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

func (c *Client) makeRequest(sid string, body io.Reader) (*http.Response, error) {
	// Create request
	req, err := http.NewRequest("POST", c.baseURL+"/web/dataset/search_read", body)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("content-type", "application/json")
	req.Header.Set("cookie", "session_id="+sid)

	// Send request
	res, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("sending HTTP request: %w", err)
	} else if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("expected HTTP status 200 OK, got %s", res.Status)
	}
	return res, nil
}

func (c *Client) unmarshalResponse(body io.ReadCloser, into interface{}) error {
	b, err := io.ReadAll(body)
	defer body.Close()
	if err != nil {
		return fmt.Errorf("read result: %w", err)
	}

	buf := bytes.NewBuffer(b)
	// decode response
	if err := DecodeResult(buf, &into); err != nil {
		return fmt.Errorf("decoding result: %w", err)
	}
	return nil
}
