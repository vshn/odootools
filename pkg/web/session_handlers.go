package web

import (
	"context"
	"errors"
	"net/http"

	pipeline "github.com/ccremer/go-command-pipeline"
	"github.com/labstack/echo/v4"
	"github.com/vshn/odootools/pkg/odoo"
	"github.com/vshn/odootools/pkg/odoo/model"
	"github.com/vshn/odootools/pkg/web/controller"
)

const (
	// SessionCookieID is the session cookie identifier.
	SessionCookieID = "odootools"
)

// LoginForm GET /login
func (s Server) LoginForm(e echo.Context) error {
	return e.Render(http.StatusOK, "login", nil)
}

// Login POST /login
func (s Server) Login(e echo.Context) error {
	odooSession, err := s.odooClient.Login(context.Background(), odoo.LoginOptions{
		DatabaseName: s.dbName,
		Username:     e.FormValue("login"),
		Password:     e.FormValue("password"),
	})
	if errors.Is(err, odoo.ErrInvalidCredentials) {
		return e.Render(http.StatusOK, "login", controller.Values{"Error": "Invalid login or password"})
	}
	if err != nil {
		e.Logger().Error(err)
		return e.Render(http.StatusBadGateway, "error", controller.AsError(errors.New("got an error from Odoo, check logs")))
	}

	return s.runPostLogin(e, odooSession)
}

func (s Server) runPostLogin(e echo.Context, odooSession *odoo.Session) error {
	o := model.NewOdoo(odooSession)
	sessionData := controller.SessionData{}
	p := pipeline.NewPipeline().WithSteps(
		pipeline.NewStepFromFunc("fetch employee", func(ctx context.Context) error {
			e, err := o.FetchEmployeeByUserID(ctx, odooSession.UID)
			sessionData.Employee = e
			return err
		}),
		pipeline.NewStepFromFunc("fetch manager group", func(ctx context.Context) error {
			group, err := o.FetchGroupByName(ctx, "Human Resources", "Manager")
			if group != nil {
				for _, userID := range group.UserIDs {
					if odooSession.UID == userID {
						sessionData.Roles = []string{controller.HRManagerRoleKey}
					}
				}
			}
			return err
		}),
		pipeline.NewStepFromFunc("save session", func(ctx context.Context) error {
			if err := s.SaveOdooSession(e, odooSession); err != nil {
				return err
			}
			if err := s.SaveSessionData(e, sessionData); err != nil {
				return err
			}
			return e.Redirect(http.StatusFound, "/report")
		}).WithErrorHandler(func(ctx context.Context, err error) error {
			return s.ShowError(e, err)
		}),
	)
	return p.RunWithContext(e.Request().Context()).Err()
}

// Logout GET /logout
func (s Server) Logout(e echo.Context) error {
	e.SetCookie(&http.Cookie{Name: SessionCookieID, MaxAge: -1})
	e.SetCookie(&http.Cookie{Name: DataCookieID, MaxAge: -1})
	return e.Redirect(http.StatusTemporaryRedirect, "/login")
}
