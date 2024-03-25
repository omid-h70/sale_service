package mid

import (
	"context"
	"fmt"
	"net/http"
	"runtime/debug"
	"service/domain/sys/metrics"
	"service/foundation/web"
)

func Panics() web.MiddlewareFunc {

	m := func(handler web.HandlerFunc) web.HandlerFunc {

		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) (err error) {

			defer func() {
				if rc := recover(); rc != nil {
					//we get stack trace here
					trace := debug.Stack()
					err = fmt.Errorf("PANIC [%v] TRACE [%s]", rc, string(trace))
					//!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
					//err will be returned as named argument so you can have it

					metrics.AddPanics(ctx)
				}
			}()

			//Execute the Original One when tmp is called
			return handler(ctx, w, r)
		}
		return h
	}
	return m

}
