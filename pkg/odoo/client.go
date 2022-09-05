package odoo

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Client is the base struct that holds information required to talk to Odoo
type Client struct {
	parsedURL *url.URL
	http      *http.Client
}

// ClientOptions configures the Odoo client.
type ClientOptions struct {
	// UseDebugLogger sets the http.Transport field of the internal http client with a transport implementation that logs the raw contents of requests and responses.
	// The logger is retrieved from the request's context via logr.FromContextOrDiscard.
	// The log level used is '2'.
	// Any "password":"..." byte content is replaced with a placeholder to avoid leaking credentials.
	// Still, this should not be called in production as other sensitive information might be leaked.
	// This method is meant to be called before any requests are made (for example after setting up the Client).
	UseDebugLogger bool
}

// Open returns a new Session by trying to log in.
// The URL must be in the format of `https://user:pass@host[:port]/db-name`.
// It returns error if baseURL is not parseable with url.Parse or if the Login failed.
func Open(ctx context.Context, fullURL string, options ClientOptions) (*Session, error) {
	client, err := NewClient(fullURL, options)
	if err != nil {
		return nil, err
	}
	login, err := client.parseOdooURL(fullURL)
	if err != nil {
		return nil, err
	}
	if login.Username == "" || login.Password == "" || login.DatabaseName == "" {
		return nil, fmt.Errorf("missing database name, username or password in URL")
	}
	return client.Login(ctx, login)
}

// NewClient returns a new Client.
func NewClient(baseURL string, options ClientOptions) (*Client, error) {
	client := &Client{}
	parsed, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}
	client.parsedURL = parsed

	client.http = &http.Client{
		Timeout: 10 * time.Second,
		Jar:     nil, // don't save any cookies!
	}

	client.useDebugLogger(options.UseDebugLogger)
	return client, nil
}

// RestoreSession restores a Session based on existing Session.SessionID and Session.UID.
// It's not validated if the session is valid.
func RestoreSession(client *Client, sessionID string, userID int) *Session {
	return &Session{
		SessionID: sessionID,
		UID:       userID,
		client:    client,
	}
}

func (c *Client) parseOdooURL(baseURL string) (LoginOptions, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return LoginOptions{}, fmt.Errorf("proper URL format is required: %w", err)
	}

	options := LoginOptions{}

	if u.User != nil {
		options.Username = u.User.Username()
		pw, _ := u.User.Password()
		options.Password = pw
	}

	// Technical debt: This means an Odoo running under a path like https://odoo/pathprefix/ can't be parsed.
	options.DatabaseName = strings.Trim(u.Path, "/")

	c.parsedURL = &url.URL{Scheme: u.Scheme, Host: u.Host}
	return options, nil
}

// LoginOptions contains all necessary authentication parameters.
type LoginOptions struct {
	DatabaseName string `json:"db,omitempty"`
	Username     string `json:"login,omitempty"`
	Password     string `json:"password,omitempty"`
}

// Login tries to authenticate the user against Odoo.
// It returns a session if authentication was successful. An error is returned if
//   - the credentials were wrong,
//   - encoding or sending the request,
//   - or decoding the request failed.
func (c Client) Login(ctx context.Context, options LoginOptions) (*Session, error) {
	resp, err := c.requestSession(ctx, options)
	if err != nil {
		return nil, err
	}

	return c.decodeSession(resp)
}

func (c Client) requestSession(ctx context.Context, options LoginOptions) (*http.Response, error) {
	// Prepare request
	body, err := NewJSONRPCRequest(options).Encode()
	if err != nil {
		return nil, newEncodingRequestError(err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.parsedURL.String()+"/web/session/authenticate", body)
	if err != nil {
		return nil, newCreatingRequestError(err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("login: sending HTTP request: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("login: expected HTTP status 200 OK, got %s", resp.Status)
	}
	return resp, nil
}

func (c *Client) decodeSession(res *http.Response) (*Session, error) {
	// Decode response
	// We don't use DecodeResult here because we're interested in whether unmarshalling the result failed.
	// If so, this is likely because "uid" is set to `false` which indicates an authentication failure.
	var response JSONRPCResponse
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("login: decode response: %w", err)
	}
	if response.Error != nil {
		return nil, fmt.Errorf("error from Odoo: %v", response.Error)
	}

	// Decode session
	var session Session
	if err := json.Unmarshal(*response.Result, &session); err != nil {
		// UID is not set, authentication failed
		return nil, ErrInvalidCredentials
	}
	session.client = c
	return &session, nil
}
