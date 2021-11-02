package web

import (
	"encoding/base64"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/securecookie"
	"github.com/vshn/odootools/pkg/odoo"
	"github.com/vshn/odootools/pkg/web/html"
)

// RootDir optionally describes where the root directory with the "templates"
/// and "public" folders is. This is mainly used in tests where the context dir is different.
var RootDir = ""

type Server struct {
	router       *mux.Router
	html         *html.Renderer
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

	s := Server{
		odoo:         odoo,
		html:         html.NewRenderer(),
		securecookie: securecookie.New(key, key),
	}
	s.routes(middleware...)
	return &s
}

func (s Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}
