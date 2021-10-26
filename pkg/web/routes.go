package web

import (
	"net/http"
	"path/filepath"

	"github.com/gorilla/mux"
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
	publicRoot := filepath.Join(RootDir, "public")
	r.PathPrefix("/").Handler(http.FileServer(http.Dir(publicRoot)))

	s.router = router
}
