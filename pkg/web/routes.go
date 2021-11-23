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
	r.Handle("/report", s.RedirectReport()).Methods("POST")
	r.Handle("/report/{employee:[0-9]+}/{year:[0-9]+}/{month:[0-9]+}", s.OvertimeReport()).Methods("GET")

	// Authentication
	r.Handle("/login", s.LoginForm()).Methods("GET")
	r.Handle("/login", s.Login()).Methods("POST")
	r.Handle("/logout", s.Logout()).Methods("GET")

	// static files
	r.PathPrefix("/").Handler(http.FileServer(http.FS(templates.PublicFS)))

	s.router = router
}
