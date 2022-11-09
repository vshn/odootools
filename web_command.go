package main

import (
	"fmt"
	"time"

	"github.com/urfave/cli/v2"
	"github.com/vshn/odootools/pkg/odoo"
	"github.com/vshn/odootools/pkg/web"
	"github.com/vshn/odootools/pkg/web/controller"
)

func RunWebServer(cli *cli.Context) error {

	loc, err := time.LoadLocation(cli.String(newDefaultTimezoneFlag().Name))
	if err != nil {
		return fmt.Errorf("cannot load timezone: %w", err)
	}
	controller.DefaultTimeZone = loc

	client, err := odoo.NewClient(cli.String(newOdooURLFlag().Name), odoo.ClientOptions{UseDebugLogger: cli.Int(newLogLevelFlag().Name) >= 2})
	if err != nil {
		return err
	}
	server := web.NewServer(
		client,
		cli.String(newSecretKeyFlag().Name),
		cli.String(newOdooDBFlag().Name),
		versionInfo,
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
			newOdooURLFlag(),
			newOdooDBFlag(),
			newSecretKeyFlag(),
			newListenAddress(),
			newDefaultTimezoneFlag(),
			newTLSCertFlag(),
			newTLSKeyFlag(),
		},
	}
}
