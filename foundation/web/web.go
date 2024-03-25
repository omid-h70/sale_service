package web

import (
	"context"
	"github.com/dimfeld/httptreemux"
	"github.com/google/uuid"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/trace"
	"net/http"
	"os"
	"syscall"
	"time"
)

type HandlerFunc func(context.Context, http.ResponseWriter, *http.Request) error

/* in v1 we embedded our httptreemux inside but in v2 we need
otel to be executed before treemux
type App struct {
	*httptreemux.ContextMux
	Shutdown    chan os.Signal
	middlewares []MiddlewareFunc
}
*/

type App struct {
	mux         *httptreemux.ContextMux
	otmux       http.Handler
	Shutdown    chan os.Signal
	middlewares []MiddlewareFunc
}

func NewApp(shutdown chan os.Signal, mw ...MiddlewareFunc) *App {

	mux := httptreemux.NewContextMux()
	return &App{
		//pre mux is like middleware
		mux: mux,
		//otmux is the outer layer
		otmux:       otelhttp.NewHandler(mux, "request"),
		Shutdown:    shutdown,
		middlewares: mw,
	}
}

func (a *App) shutdownSignal() {
	a.Shutdown <- syscall.SIGTERM
}

/* ServeHTTP
we implemented the method instead of type embedding treemux
force otmux to execute first
*/

func (a *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.otmux.ServeHTTP(w, r)
}

func (a *App) Handle(method string, prefix string, path string, hFunc HandlerFunc, mw ...MiddlewareFunc) {

	//Pre Code processing
	hFunc = wrapMiddlewares(mw, hFunc)
	hFunc = wrapMiddlewares(a.middlewares, hFunc)

	treeMuxFunc := func(w http.ResponseWriter, r *http.Request) {

		ctx := r.Context()

		span := trace.SpanFromContext(ctx)
		_ = span.SpanContext().TraceID().String()
		//TODO: replace it with uuid TraceID

		v := Value{
			TraceID: uuid.New().String(),
			Now:     time.Now(),
		}
		ctx = context.WithValue(ctx, key, &v)

		if err := hFunc(ctx, w, r); err != nil {
			a.shutdownSignal()
			return
		}

		//Post Code processing
	}
	routePath := path
	if len(prefix) > 0 {
		routePath = "/" + prefix + path
	}
	a.mux.Handle(method, routePath, treeMuxFunc)
}
