package web

import (
	"net/http"
)

// Dashboard GET /
func (s Server) Dashboard() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session := s.sessionFrom(r)
		if session == nil {
			// User is unauthenticated
			http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
		}

		attendances, err := s.odoo.ReadAllAttendances(session.ID, session.UID)
		if err != nil {
			return nil, err
		}

	})
}
