package web

import (
	"embed"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/vshn/odootools/templates"
)

func (s *Server) setupRoutes(middleware ...echo.MiddlewareFunc) {
	e := s.Echo
	// System setupRoutes
	e.GET("/healthz", Healthz)

	// Application routes
	e.GET("/", s.RedirectTo("/report"))

	report := e.Group("/report", middleware...)
	report.GET("", s.RequestReportForm)
	report.POST("", s.ProcessReportInput)
	report.GET("/employees/:year/:month", s.EmployeeReport)
	report.GET("/:employee/:year", s.OvertimeReport)
	report.GET("/:employee/:year/:month", s.OvertimeReport)

	e.GET("/help", s.helpPage, middleware...)

	// Authentication
	e.GET("/login", s.LoginForm)
	e.POST("/login", s.Login)
	e.GET("/logout", s.Logout)

	// static files
	e.GET("/robots.txt", EmbeddedFile(templates.PublicFS, "robots.txt", "text/plain; charset=UTF-8"))
	e.GET("/favicon.ico", EmbeddedFile(templates.PublicFS, "favicon.png", "image/png"))
	e.GET("/static/*", echo.WrapHandler(http.StripPrefix("/static/", http.FileServer(http.FS(templates.PublicFS)))))
}

func Healthz(e echo.Context) error {
	return e.String(http.StatusOK, "")
}

func EmbeddedFile(fs embed.FS, fileName string, contentType string) echo.HandlerFunc {
	return func(e echo.Context) error {
		file, err := fs.Open(fileName)
		if err != nil {
			return err
		}
		return e.Stream(http.StatusOK, contentType, file)
	}
}

func (s Server) RedirectTo(url string) echo.HandlerFunc {
	return func(context echo.Context) error {
		return context.Redirect(http.StatusTemporaryRedirect, url)
	}
}
