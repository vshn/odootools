package web

import (
	"errors"
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/vshn/odootools/pkg/odoo"
	"github.com/vshn/odootools/pkg/web/controller"
)

const (
	CookieSID = "odootools"
)

// LoginForm GET /login
func (s Server) LoginForm(e echo.Context) error {
	return e.Render(http.StatusOK, "login", nil)
}

// Login POST /login
func (s Server) Login(e echo.Context) error {

	odooSession, err := s.odoo.Login(e.FormValue("login"), e.FormValue("password"))
	if errors.Is(err, odoo.ErrInvalidCredentials) {
		return e.Render(http.StatusOK, "login", controller.Values{"Error": "Invalid login or password"})
	}
	if err != nil {
		e.Logger().Error(err)
		return e.Render(http.StatusBadGateway, "error", controller.AsError(errors.New("got an error from Odoo, check logs")))
	}
	if err := s.SaveOdooSession(e, odooSession); err != nil {
		return s.ShowError(e, err)
	}
	return e.Redirect(http.StatusFound, "/report")
}

// Logout GET /logout
func (s Server) Logout(e echo.Context) error {
	e.SetCookie(&http.Cookie{Name: CookieSID, MaxAge: -1})
	return e.Redirect(http.StatusTemporaryRedirect, "/login")
}

func (s Server) GetOdooSession(e echo.Context) *odoo.Session {
	sess, _ := session.Get(CookieSID, e)
	odooSess := &odoo.Session{
		SessionID: sess.Values["odoo_id"].(string),
		UID:       sess.Values["odoo_uid"].(int),
	}
	return odooSess
}

func (s Server) SaveOdooSession(e echo.Context, odooSession *odoo.Session) error {
	sess := sessions.NewSession(s.cookieStore, CookieSID)
	sess.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7,
		HttpOnly: true,
		Secure:   true,
	}
	sess.Values["odoo_id"] = odooSession.SessionID
	sess.Values["odoo_uid"] = odooSession.UID
	return sess.Save(e.Request(), e.Response())
}
