package web

import "github.com/vshn/odootools/pkg/odoo"

func newTestServer(odooURL string) *Server {
	var oc *odoo.Client
	if odooURL != "" {
		c, err := odoo.NewClient(odooURL, odoo.ClientOptions{})
		if err != nil {
			panic(err)
		}
		oc = c
	}
	return NewServer(oc, "0000000000000000000000000000000000000000000=", "TestDB", VersionInfo{})
}
