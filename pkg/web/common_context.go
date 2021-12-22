package web

import (
	"net/http"
	"strconv"

	pipeline "github.com/ccremer/go-command-pipeline"
	"github.com/vshn/odootools/pkg/odoo"
	"github.com/vshn/odootools/pkg/web/views"
)

type CommonContext struct {
	Request    *http.Request
	Response   http.ResponseWriter
	Session    *odoo.Session
	OdooClient *odoo.Client
	ErrorView  *views.ErrorView
}

type CommonContextAccessor interface {
	GetCommonContext() *CommonContext
}

func (CommonContext) checkLogin(ctx pipeline.Context) error {
	c := ctx.(CommonContextAccessor).GetCommonContext()
	if c.Session == nil {
		// User is unauthenticated
		http.Redirect(c.Response, c.Request, "/login", http.StatusTemporaryRedirect)
		return pipeline.ErrAbort
	}
	return nil
}

func (CommonContext) showError(ctx pipeline.Context, err error) error {
	c := ctx.(CommonContextAccessor).GetCommonContext()
	c.ErrorView.ShowError(c.Response, err)
	return err
}

func parseIntOrDefault(toParse string, def int) int {
	if toParse == "" {
		return def
	}
	if v, err := strconv.Atoi(toParse); err == nil {
		return v
	}
	return def
}
