package main

import "github.com/urfave/cli/v2"

func newOdooURLFlag() *cli.StringFlag {
	return &cli.StringFlag{
		Name:     "odoo-url",
		Usage:    "Odoo Base URL",
		Required: true,
		EnvVars:  []string{"ODOO_URL"},
	}
}

func newOdooDBFlag() *cli.StringFlag {
	return &cli.StringFlag{
		Name:     "odoo-db",
		Usage:    "Odoo Database name",
		Required: true,
		EnvVars:  []string{"ODOO_DB"},
	}
}

func newLogLevelFlag() *cli.IntFlag {
	return &cli.IntFlag{
		Name:    "log-level",
		Usage:   "Log verbosity level",
		EnvVars: []string{"LOG_LEVEL"},
	}
}

func newSecretKeyFlag() *cli.StringFlag {
	return &cli.StringFlag{
		Name:     "secret-key",
		Usage:    "Secret Key (e.g. to encrypt cookies). Create a new key with 'openssl rand -base64 32'",
		Required: true,
		EnvVars:  []string{"SECRET_KEY"},
	}
}

func newListenAddress() *cli.StringFlag {
	return &cli.StringFlag{
		Name:    "listen-address",
		Usage:   "The interface address where the web server should listen on",
		EnvVars: []string{"LISTEN_ADDRESS"},
		Value:   ":4200",
	}
}

func newTLSCertFlag() *cli.StringFlag {
	return &cli.StringFlag{
		Name:    "tls-cert",
		Usage:   "The path to a certificate file to serve",
		EnvVars: []string{"TLS_CERT"},
	}
}

func newTLSKeyFlag() *cli.StringFlag {
	return &cli.StringFlag{
		Name:    "tls-key",
		Usage:   "The path to a certificate private key file to serve",
		EnvVars: []string{"TLS_KEY"},
	}
}
