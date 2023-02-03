// Package handlers manages the different versions of the API.
package handlers

import (
	"net/http"
	"os"

	"github.com/ardanlabs/service/app/services/sales-api/handlers/v1/testgrp"
	"github.com/ardanlabs/service/app/services/sales-api/handlers/v1/usergrp"
	"github.com/ardanlabs/service/business/core/user"
	"github.com/ardanlabs/service/business/core/user/stores/userdb"
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

	authen := mid.Authenticate(cfg.Auth)
	ruleAdmin := mid.Authorize(cfg.Auth, auth.RuleAdminOnly)
	ruleAny := mid.Authorize(cfg.Auth, auth.RuleAny)

	// =========================================================================

	tg := testgrp.Handlers{
		DB: cfg.DB,
	}

	app.Handle(http.MethodGet, "/status", tg.Status)
	app.Handle(http.MethodGet, "/auth", tg.Status, authen, ruleAdmin)

	// =========================================================================

	ugh := usergrp.Handlers{
		User: user.NewCore(userdb.NewStore(cfg.Log, cfg.DB)),
		Auth: cfg.Auth,
	}
	app.Handle(http.MethodGet, "/users/token/:kid", ugh.Token)
	app.Handle(http.MethodGet, "/users/:page/:rows", ugh.Query, authen, ruleAdmin)
	app.Handle(http.MethodGet, "/users/:id", ugh.QueryByID, ruleAny)
	app.Handle(http.MethodPost, "/users", ugh.Create, authen, ruleAdmin)
	app.Handle(http.MethodPut, "/users/:id", ugh.Update, authen, ruleAny)
	app.Handle(http.MethodDelete, "/users/:id", ugh.Delete, authen, ruleAny)

	return app
}
