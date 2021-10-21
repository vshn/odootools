package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/mhutter/vshn-ftb/pkg/odoo"
	"github.com/mhutter/vshn-ftb/pkg/web"
	"github.com/mhutter/vshn-ftb/pkg/web/middleware"
)

func main() {
	app := web.NewServer(
		odoo.NewClient(mustGetEnv("ODOO_URL"), mustGetEnv("ODOO_DB")),
		mustGetEnv("SECRET_KEY"),
		middleware.AccessLog,
	)

	srv := http.Server{
		Handler:        app,
		Addr:           getEnvOr("LISTEN", "vshn-ftb.127.0.0.1.nip.io") + ":" + getEnvOr("PORT", "4200"),
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1MiB
	}

	log.Printf("Starting odoo at https://%s\n", srv.Addr)
	//log.Println(srv.ListenAndServeTLS(
	//	getEnvOr("TLS_CERT", "tls/cert.pem"), getEnvOr("TLS_KEY", "tls/key.pem"),
	//))
	log.Println(srv.ListenAndServe())
}

func getEnvOr(name, fallback string) string {
	if v := os.Getenv(name); v != "" {
		return v
	}
	return fallback
}

func mustGetEnv(name string) string {
	if v := os.Getenv(name); v != "" {
		return v
	}

	log.Fatalf("Mandatory $%s is not set", name)
	return ""
}
