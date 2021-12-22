package cmd

import (
	"github.com/urfave/cli/v2"
	"github.com/vshn/odootools/pkg/odoo"
	"github.com/vshn/odootools/pkg/web"
)

func RunWebServer(cli *cli.Context) error {

	server := web.NewServer(
		odoo.NewClient(cli.String("odoo-url"), cli.String("odoo-db")),
		cli.String("secret-key"),
	)

	addr := cli.String("listen-address")

	if certPath := cli.String("tls-cert"); certPath != "" {
		return server.Echo.StartTLS(addr, cli.String("tls-cert"), cli.String("tls-key"))
	}
	return server.Echo.Start(addr)
}

var WebCommand = &cli.Command{
	Name:   "web",
	Usage:  "Starts the web server",
	Action: RunWebServer,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "secret-key",
			Usage:    "Secret Key (e.g. to encrypt cookies). Create a new key with 'openssl rand -base64 32'",
			Required: true,
			EnvVars:  []string{"SECRET_KEY"},
		},
		&cli.StringFlag{
			Name:    "listen-address",
			Usage:   "The interface address where the web server should listen on",
			EnvVars: []string{"LISTEN_ADDRESS"},
			Value:   "odootools.127.0.0.1.nip.io:4200",
		},
		&cli.StringFlag{
			Name:    "tls-cert",
			Usage:   "The path to a certificate file to serve",
			EnvVars: []string{"TLS_CERT"},
		},
		&cli.StringFlag{
			Name:    "tls-key",
			Usage:   "The path to a certificate private key file to serve",
			EnvVars: []string{"TLS_KEY"},
		},
	},
}
