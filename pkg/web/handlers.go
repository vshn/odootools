package web

import (
	"net/http"
)

// LoginForm GET /login
func (s Server) LoginForm() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.html.Render(w, "login", nil)
	})
}
