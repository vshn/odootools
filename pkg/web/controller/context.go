package controller

import (
	"github.com/labstack/echo/v4"
	"github.com/vshn/odootools/pkg/odoo"
)

type Context struct {
	Echo        echo.Context
	OdooSession *odoo.Session
	OdooClient  *odoo.Client
}
