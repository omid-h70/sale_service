package mid

import (
	"context"
	"net/http"
	"service/domain/sys/metrics"
	"service/foundation/web"
)

func Metrics() web.MiddlewareFunc {

	m := func(handler web.HandlerFunc) web.HandlerFunc {

		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {

			metrics.Set(ctx)

			//Execute the Original One when tmp is called
			err := handler(ctx, w, r)

			metrics.AddRequests(ctx)
			metrics.AddGoroutines(ctx)

			if err != nil {
				metrics.AddErrors(ctx)
			}

			return err
		}
		return h
	}
	return m

}
