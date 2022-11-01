package main

import (
	"github.com/urfave/cli/v2"
	"github.com/vshn/odootools/pkg/odoo"
	"github.com/vshn/odootools/pkg/web"
)

func RunWebServer(cli *cli.Context) error {

	client, err := odoo.NewClient(cli.String(newOdooURLFlag().Name), odoo.ClientOptions{UseDebugLogger: cli.Int(newLogLevelFlag().Name) >= 2})
	if err != nil {
		return err
	}
	server := web.NewServer(
		client,
		cli.String(newSecretKeyFlag().Name),
		cli.String(newOdooDBFlag().Name),
	)

	addr := cli.String(newListenAddress().Name)

	if certPath := cli.String(newTLSCertFlag().Name); certPath != "" {
		return server.Echo.StartTLS(addr, cli.String(newTLSCertFlag().Name), cli.String(newTLSKeyFlag().Name))
	}
	return server.Echo.Start(addr)
}

func newWebCommand() *cli.Command {
	return &cli.Command{
		Name:   "web",
		Usage:  "Starts the web server",
		Action: RunWebServer,
		Flags: []cli.Flag{
			newSecretKeyFlag(),
			newListenAddress(),
			newTLSCertFlag(),
			newTLSKeyFlag(),
		},
	}
}
