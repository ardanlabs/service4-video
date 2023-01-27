// Package handlers manages the different versions of the API.
package handlers

import (
	"net/http"
	"os"

	"github.com/ardanlabs/service/app/services/sales-api/handlers/v1/testgrp"
	"github.com/ardanlabs/service/business/web/auth"
	"github.com/ardanlabs/service/business/web/v1/mid"
	"github.com/ardanlabs/service/foundation/web"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

// APIMuxConfig contains all the mandatory systems required by handlers.
type APIMuxConfig struct {
	Shutdown chan os.Signal
	Log      *zap.SugaredLogger
	Auth     *auth.Auth
	DB       *sqlx.DB
}

// APIMux constructs a http.Handler with all application routes defined.
func APIMux(cfg APIMuxConfig) *web.App {
	app := web.NewApp(cfg.Shutdown, mid.Logger(cfg.Log), mid.Errors(cfg.Log), mid.Metrics(), mid.Panics())

	tg := testgrp.Handlers{
		DB: cfg.DB,
	}

	app.Handle(http.MethodGet, "/status", tg.Status)

	authen := mid.Authenticate(cfg.Auth)
	admin := mid.Authorize(cfg.Auth, auth.RuleAdminOnly)

	app.Handle(http.MethodGet, "/auth", tg.Status, authen, admin)

	return app
}
