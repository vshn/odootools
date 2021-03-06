package controller

import (
	"context"

	"github.com/labstack/echo/v4"
	"github.com/vshn/odootools/pkg/odoo"
	"github.com/vshn/odootools/pkg/odoo/model"
)

type BaseController struct {
	Echo           echo.Context
	OdooClient     *model.Odoo
	OdooSession    *odoo.Session
	SessionData    SessionData
	RequestContext context.Context
}

const HRManagerRoleKey = "HRManager"

// SessionData is an additional data struct.
// Its purpose is to store data in a session cookie in order to avoid repetitive Odoo API calls.
type SessionData struct {
	Employee *model.Employee `json:"employee"`
	Roles    []string        `json:"roles"`
}
