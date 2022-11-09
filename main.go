package main

import (
	"log"
	"os"
	"time"

	"github.com/urfave/cli/v2"
	"github.com/vshn/odootools/pkg/web"
)

var (
	version = "unknown"
	commit  = "-dirty-"
	date    = time.Now().Format("2006-01-02")

	versionInfo web.VersionInfo
)

func main() {
	versionInfo = web.VersionInfo{Version: version, Commit: commit, Date: date}
	app := &cli.App{
		Name:    "odootools",
		Usage:   "Odoo ERP utility tools for everyday things",
		Version: versionInfo.String(),
		Flags: []cli.Flag{
			newOdooURLFlag(),
			newOdooDBFlag(),
			newLogLevelFlag(),
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
