package handlers

import (
	"expvar"
	"github.com/dimfeld/httptreemux"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
	"net/http"
	"net/http/pprof"
	"os"
	"service/app/services/sales-api/handlers/debug/checkgrp"
	"service/app/services/sales-api/handlers/v1/testgrp"
	v1UserGrp "service/app/services/sales-api/handlers/v1/usergrp"
	"service/domain/core/user"
	"service/domain/sys/auth"
	"service/domain/web/mid"
	"service/foundation/web"
)

//Keep default mux router clean and use our custom mux

func DebugStandardLibraryMux() *http.ServeMux {

	mux := http.NewServeMux()
	mux.HandleFunc("debug/pprof", pprof.Index)
	mux.HandleFunc("debug/cmdline", pprof.Cmdline)
	mux.HandleFunc("debug/profile", pprof.Profile)
	mux.HandleFunc("debug/symbol", pprof.Symbol)
	mux.HandleFunc("debug/trace", pprof.Trace)
	mux.Handle("debug/vars", expvar.Handler())

	return mux
}

func DebugMux(build string, log *zap.SugaredLogger, db *sqlx.DB) http.Handler {
	cgh := checkgrp.Handlers{
		Build: build,
		Log:   log,
		DB:    db,
	}

	mux := DebugStandardLibraryMux()
	mux.HandleFunc("debug/liveness", cgh.Liveness)
	mux.HandleFunc("debug/readiness", cgh.Readiness)
	return mux
}

type APIMuxConfig struct {
	Build    string
	Shutdown chan os.Signal
	Log      *zap.SugaredLogger
	Auth     *auth.Auth
	DB       *sqlx.DB
}

func APIMux(cfg APIMuxConfig) *httptreemux.ContextMux {
	treeMux := httptreemux.NewContextMux()
	//treeMux.Handle()
	return treeMux
}

func AppAPIMux(cfg APIMuxConfig) *web.App {
	app := web.NewApp(
		cfg.Shutdown,
		mid.Logger(cfg.Log),
		mid.Errors(cfg.Log),
		mid.Panics(),
		mid.Metrics(),
	)
	v1(app, cfg)
	return app
}

func v1(app *web.App, cfg APIMuxConfig) {

	const version = "v1"
	thg := testgrp.Handlers{
		Log: cfg.Log,
	}
	app.Handle(http.MethodGet, version, "/test", thg.Test)
	app.Handle(http.MethodGet, version, "/test/auth", thg.TestAuth, mid.Authenticate(cfg.Auth), mid.Authorize(auth.RoleAdmin))

	ugh := v1UserGrp.Handlers{
		Core: user.NewCore(cfg.Log, cfg.DB),
		Auth: cfg.Auth,
	}

	app.Handle(http.MethodGet, version, "", ugh.Token)
	app.Handle(http.MethodGet, version, "users/:page/:rows", ugh.Query, mid.Authenticate(cfg.Auth), mid.Authorize(auth.RoleAdmin))
	app.Handle(http.MethodGet, version, "users/:id", ugh.QueryByID, mid.Authenticate(cfg.Auth))
	app.Handle(http.MethodPost, version, "users", ugh.Create, mid.Authenticate(cfg.Auth), mid.Authorize(auth.RoleAdmin))
	app.Handle(http.MethodPut, version, "users/:id", ugh.Update, mid.Authenticate(cfg.Auth), mid.Authorize(auth.RoleAdmin))
	app.Handle(http.MethodDelete, version, "users/:id", ugh.Delete, mid.Authenticate(cfg.Auth), mid.Authorize(auth.RoleAdmin))

}
