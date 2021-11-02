package web

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/vshn/odootools/templates"
)

func (s *Server) routes(middleware ...mux.MiddlewareFunc) {
	router := mux.NewRouter()

	// System routes
	router.HandleFunc("/healthz", Healthz).Methods("GET")

	// Application routes
	r := router.NewRoute().Subrouter()
	r.Use(middleware...)

	r.Handle("/", s.RequestReportForm()).Methods("GET")
	r.Handle("/report", s.RequestReportForm()).Methods("GET")
	r.Handle("/report", s.OvertimeReport()).Methods("POST")

	// Authentication
	r.Handle("/login", s.LoginForm()).Methods("GET")
	r.Handle("/login", s.Login()).Methods("POST")
	r.Handle("/logout", s.Logout()).Methods("GET")

	// static files
	r.PathPrefix("/").Handler(http.FileServer(http.FS(templates.PublicFS)))

	s.router = router
}
