package web

import (
	"encoding/json"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/vshn/odootools/pkg/odoo"
	"github.com/vshn/odootools/pkg/web/controller"
)

const (
	// DataCookieID is the cookie identifier for storing additional data.
	DataCookieID = "odootools-data"
)

// GetOdooSession returns the Odoo session from the session cookie.
// Returns nil if there is no active session.
func (s *Server) GetOdooSession(e echo.Context) *odoo.Session {
	sess, _ := session.Get(SessionCookieID, e)
	odooId := sess.Values["odoo_id"]
	odooUid := sess.Values["odoo_uid"]
	if odooId == nil && odooUid == nil {
		return nil
	}
	odooSess := odoo.RestoreSession(s.odooClient, odooId.(string), odooUid.(int))
	return odooSess
}

func (s *Server) GetSessionData(e echo.Context) controller.SessionData {
	sess, _ := session.Get(DataCookieID, e)
	data := controller.SessionData{}
	if raw, found := sess.Values["data"]; found {
		err := json.Unmarshal([]byte(raw.(string)), &data)
		if err != nil {
			e.Logger().Errorf("no session data found: %v", err)
		}
	}
	return data
}

func (s *Server) SaveOdooSession(e echo.Context, odooSession *odoo.Session) error {
	sess := sessions.NewSession(s.cookieStore, SessionCookieID)
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

func (s *Server) SaveSessionData(e echo.Context, data controller.SessionData) error {
	sess := sessions.NewSession(s.cookieStore, DataCookieID)
	sess.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7,
		HttpOnly: true,
		Secure:   true,
	}
	b, err := json.Marshal(data)
	if err != nil {
		return err
	}
	sess.Values["data"] = string(b)
	return sess.Save(e.Request(), e.Response())
}
