package web

import (
	"encoding/base64"
	"fmt"
	"log"
	"net/http"

	"github.com/go-logr/logr"
	"github.com/go-logr/logr/funcr"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/vshn/odootools/pkg/odoo"
	"github.com/vshn/odootools/pkg/odoo/model"
	"github.com/vshn/odootools/pkg/web/controller"
)

type Server struct {
	odooClient  *odoo.Client
	Echo        *echo.Echo
	cookieStore *sessions.CookieStore
	dbName      string
	versionInfo VersionInfo
}

func NewServer(
	odoo *odoo.Client,
	secretKey string,
	dbName string,
	versionInfo VersionInfo,
) *Server {
	key, err := base64.StdEncoding.DecodeString(secretKey)
	if err != nil {
		log.Fatalf("Error decoding secret key. Generate with `openssl rand -base64 32`. %v", err)
	}

	s := Server{
		odooClient:  odoo,
		dbName:      dbName,
		Echo:        echo.New(),
		cookieStore: sessions.NewCookieStore(key, key),
		versionInfo: versionInfo,
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
		KeyLookup: "cookie:" + SessionCookieID,
		Validator: func(s string, context echo.Context) (bool, error) {
			// 's' contains already the encrypted cookie value, which means we likely have a valid odoo session
			return true, nil
		},
		ErrorHandler: func(err error, context echo.Context) error {
			return context.Redirect(http.StatusTemporaryRedirect, "/login")
		},
		Skipper: s.unprotectedRoutes(),
	})
	e.Renderer = controller.NewRenderer()
	s.setupRoutes(authMiddleware)
	return &s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.Echo.ServeHTTP(w, r)
}

func (s *Server) newControllerContext(e echo.Context) *controller.BaseController {
	sess := s.GetOdooSession(e)
	data := s.GetSessionData(e)
	logCtx := logr.NewContext(e.Request().Context(), funcr.NewJSON(func(obj string) {
		// TODO: Integrate with echo logger?
		fmt.Println(obj)
	}, funcr.Options{Verbosity: 2}))
	return &controller.BaseController{Echo: e, OdooClient: model.NewOdoo(sess), OdooSession: sess, SessionData: data, RequestContext: logCtx}
}

func (s *Server) ShowError(e echo.Context, err error) error {
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

func (s *Server) helpPage(e echo.Context) error {
	return e.Render(http.StatusOK, "help",
		controller.Values{
			"Nav": controller.Values{
				"LoggedIn": true,
			},
		})
}
