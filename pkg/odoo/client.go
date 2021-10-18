package odoo

import (
	"net/http"
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
	return &Client{baseURL, db, &http.Client{
		Timeout: 10 * time.Second,
		Jar:     nil, // don't save any cookies!
	}}
}
