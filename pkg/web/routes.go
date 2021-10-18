package web

import (
	"github.com/gorilla/mux"
)

func (s *Server) routes(middleware ...mux.MiddlewareFunc) {
	router := mux.NewRouter()

	// System routes
	router.HandleFunc("/healthz", Healthz).Methods("GET")

	// Application routes
	r := router.NewRoute().Subrouter()
	r.Use(middleware...)

	// Authentication
	r.Handle("/login", s.LoginForm())

	s.router = router
}
