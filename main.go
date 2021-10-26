package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/urfave/cli/v2"
	"github.com/vshn/odootools/pkg/odoo"
	"github.com/vshn/odootools/pkg/web"
	"github.com/vshn/odootools/pkg/web/middleware"
)

var (
	version = "unknown"
	commit  = "-dirty-"
	date    = time.Now().Format("2006-01-02")
)

type VersionInfo struct {
	Version string
	Commit  string
	Date    string
}

func (vi VersionInfo) String() string {
	dateLayout := "2006-01-02"
	t, _ := time.Parse(dateLayout, vi.Date)
	return fmt.Sprintf("%s, commit %s, date %s", vi.Version, vi.Commit[0:7], t.Format(dateLayout))
}

func main() {

	app := &cli.App{
		Name:    "odootools",
		Usage:   "Odoo ERP utility tools for everyday things",
		Version: VersionInfo{Version: version, Commit: commit, Date: date}.String(),
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "odoo-url",
				Usage:    "Odoo Base URL",
				Required: true,
				EnvVars:  []string{"ODOO_URL"},
			},
			&cli.StringFlag{
				Name:     "odoo-db",
				Usage:    "Odoo Database name",
				Required: true,
				EnvVars:  []string{"ODOO_DB"},
			},
		},
		Commands: []*cli.Command{
			{
				Name:   "web",
				Usage:  "Starts the web server",
				Action: runServer,
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
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func runServer(context *cli.Context) error {
	server := web.NewServer(
		odoo.NewClient(context.String("odoo-url"), context.String("odoo-db")),
		context.String("secret-key"),
		middleware.AccessLog,
	)

	srv := http.Server{
		Handler:        server,
		Addr:           context.String("listen-address"),
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1MiB
	}

	log.Printf("Starting odoo at %s\n", srv.Addr)
	if certPath := context.String("tls-cert"); certPath != "" {
		return srv.ListenAndServeTLS(
			certPath, context.String("tls-key"),
		)
	}
	return srv.ListenAndServe()
}
