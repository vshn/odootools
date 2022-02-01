package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/urfave/cli/v2"
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
			&cli.IntFlag{
				Name:    "log-level",
				Usage:   "Log verbosity level",
				EnvVars: []string{"LOG_LEVEL"},
			},
		},
		Commands: []*cli.Command{
			newWebCommand(),
		},
	}

	log.Printf("odootools %s", app.Version)
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
