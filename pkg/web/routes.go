package web

import (
	"net/http"
	"path"

	"github.com/gorilla/mux"
)

func (s *Server) routes(middleware ...mux.MiddlewareFunc) {
	router := mux.NewRouter()

	// System routes
	router.HandleFunc("/healthz", Healthz).Methods("GET")

	// Application routes
	r := router.NewRoute().Subrouter()
	r.Use(middleware...)

	r.Handle("/", s.Dashboard()).Methods("GET")

	// Authentication
	r.Handle("/login", s.LoginForm()).Methods("GET")
	r.Handle("/login", s.Login()).Methods("POST")
	r.Handle("/logout", s.Logout()).Methods("GET")

	// static files
	publicRoot := path.Join(RootDir, "public")
	r.PathPrefix("/").Handler(http.FileServer(http.Dir(publicRoot)))

	s.router = router
}
