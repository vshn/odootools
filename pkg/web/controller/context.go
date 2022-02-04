package controller

import (
	"github.com/labstack/echo/v4"
	"github.com/vshn/odootools/pkg/odoo/model"
)

type Context struct {
	Echo       echo.Context
	OdooClient *model.Odoo
	OwnUserID  int
}
