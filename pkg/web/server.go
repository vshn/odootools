package web

import (
	"encoding/base64"
	"log"
	"net/http"
	"path"

	"github.com/gorilla/mux"
	"github.com/gorilla/securecookie"
	"github.com/mhutter/vshn-ftb/pkg/odoo"
	"github.com/mhutter/vshn-ftb/pkg/web/html"
)

// RootDir optionally describes where the root directory with the "templates"
/// and "public" folders is. This is mainly used in tests where the context dir is different.
var RootDir = ""

type Server struct {
	router       *mux.Router
	html         *html.View
	odoo         *odoo.Client
	securecookie *securecookie.SecureCookie
}

func NewServer(
	odoo *odoo.Client,
	secretKey string,
	middleware ...mux.MiddlewareFunc,
) *Server {
	key, err := base64.StdEncoding.DecodeString(secretKey)
	if err != nil {
		log.Fatalf("Error decoding secret key. Generate with `openssl rand -base64 32`. %v", err)
	}

	templateRoot := path.Join(RootDir, "templates")
	s := Server{
		odoo:         odoo,
		html:         html.NewView(templateRoot),
		securecookie: securecookie.New(key, key),
	}
	s.routes(middleware...)
	return &s
}

func (s Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}
