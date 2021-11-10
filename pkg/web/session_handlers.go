package web

import (
	"errors"
	"log"
	"net/http"

	"github.com/vshn/odootools/pkg/odoo"
	"github.com/vshn/odootools/pkg/web/html"
)

const (
	CookieSID = "ftb_sid"
)

func (s Server) sessionFrom(r *http.Request) *odoo.Session {
	if c, err := r.Cookie(CookieSID); err == nil {
		var sess odoo.Session
		if err = s.securecookie.Decode(CookieSID, c.Value, &sess); err == nil {
			return &sess
		}
	}

	return nil
}

// LoginForm GET /login
func (s Server) LoginForm() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.html.Render(w, "login", nil)
	})
}

// Login POST /login
func (s Server) Login() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sess, err := s.odoo.Login(r.FormValue("login"), r.FormValue("password"))
		if errors.Is(err, odoo.ErrInvalidCredentials) {
			s.html.Render(w, "login", html.Values{"Error": "Invalid login or password"})
			return
		}
		if err != nil {
			log.Println(err)
			http.Error(w, "Got an error from Odoo, check logs", http.StatusBadGateway)
			return
		}

		// Set session cookie
		val, err := s.securecookie.Encode(CookieSID, sess)
		if err != nil {
			log.Printf("Login: error encoding session: %v\n", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.SetCookie(w, &http.Cookie{
			Name:     CookieSID,
			Value:    val,
			HttpOnly: true,
			Secure:   true,
		})
		http.Redirect(w, r, "/", http.StatusFound)
	})
}

// Logout GET /logout
func (s Server) Logout() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{Name: CookieSID, MaxAge: -1})
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
	})
}
