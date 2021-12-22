package web

import (
	"encoding/base64"
	"log"
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/vshn/odootools/pkg/odoo"
	"github.com/vshn/odootools/pkg/web/controller"
)

type Server struct {
	odoo        *odoo.Client
	Echo        *echo.Echo
	cookieStore *sessions.CookieStore
}

func NewServer(
	odoo *odoo.Client,
	secretKey string,
) *Server {
	key, err := base64.StdEncoding.DecodeString(secretKey)
	if err != nil {
		log.Fatalf("Error decoding secret key. Generate with `openssl rand -base64 32`. %v", err)
	}

	s := Server{
		odoo:        odoo,
		Echo:        echo.New(),
		cookieStore: sessions.NewCookieStore(key, key),
	}
	e := s.Echo
	e.Pre(middleware.RemoveTrailingSlash())
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Skipper:          s.skipAccessLogs,
		Format:           middleware.DefaultLoggerConfig.Format,
		CustomTimeFormat: middleware.DefaultLoggerConfig.CustomTimeFormat,
	}))
	e.Use(session.MiddlewareWithConfig(session.Config{
		Store: s.cookieStore,
	}))
	authMiddleware := middleware.KeyAuthWithConfig(middleware.KeyAuthConfig{
		KeyLookup: "cookie:" + CookieSID,
		Validator: func(s string, context echo.Context) (bool, error) {
			// 's' contains already the encrypted cookie value, which means we likely have a valid odoo session
			return true, nil
		},
		ErrorHandler: func(err error, context echo.Context) error {
			return context.Redirect(http.StatusTemporaryRedirect, "/login")
		},
	})
	e.Renderer = controller.NewRenderer()
	s.setupRoutes(authMiddleware)
	return &s
}

func (s Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.Echo.ServeHTTP(w, r)
}

func (s Server) newControllerContext(e echo.Context) *controller.Context {
	return &controller.Context{Echo: e, OdooClient: s.odoo, OdooSession: s.GetOdooSession(e)}
}

func (s Server) ShowError(e echo.Context, err error) error {
	return e.Render(http.StatusInternalServerError, "error", controller.AsError(err))
}

var publicRoutes = []string{
	"/favicon.ico",
	"/robots.txt",
	"/static/*",
	"/healthz",
}

func (s *Server) skipAccessLogs(e echo.Context) bool {
	for _, path := range publicRoutes {
		if path == e.Path() {
			return true
		}
	}
	return false
}

func (s *Server) unprotectedRoutes() middleware.Skipper {
	return func(e echo.Context) bool {
		for _, path := range publicRoutes {
			if path == e.Path() {
				return true
			}
		}
		for _, path := range []string{
			"/login",
			"/logout",
			"/",
		} {
			if path == e.Path() {
				return true
			}
		}
		return false
	}
}
