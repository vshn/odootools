package web

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/mhutter/vshn-ftb/pkg/web/html"
)

type Server struct {
	router *mux.Router
	html   *html.View
}

func NewServer(templateRoot string, middleware ...mux.MiddlewareFunc) *Server {
	s := Server{html: html.NewView(templateRoot)}
	s.routes(middleware...)
	return &s
}

func (s Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}
