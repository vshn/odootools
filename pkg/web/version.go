package web

import (
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/vshn/odootools/pkg/web/controller"
)

type VersionInfo struct {
	Version string
	Commit  string
	Date    string
}

func (vi VersionInfo) String() string {
	dateLayout := "2006-01-02"
	t, _ := time.Parse(dateLayout, vi.Date)
	return fmt.Sprintf("%s, commit %s, date %s", vi.Version, vi.Commit[0:7], t.Format(dateLayout))
}

func (s *Server) aboutPage(e echo.Context) error {
	return e.Render(http.StatusOK, "about",
		controller.Values{
			"Nav": controller.Values{
				"LoggedIn": s.GetOdooSession(e) != nil,
			},
			"ChangelogLink": "https://github.com/vshn/odootools/releases",
			"Version":       s.versionInfo.Version,
			"Date":          s.versionInfo.Date,
		})
}
