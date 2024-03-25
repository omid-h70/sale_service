package mid

import (
	"context"
	"go.uber.org/zap"
	"net/http"
	"service/foundation/web"
	"time"
)

func Logger(logger *zap.SugaredLogger) web.MiddlewareFunc {

	m := func(handler web.HandlerFunc) web.HandlerFunc {

		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {

			/*Temp Struct Data */
			//v := struct {
			//	traceID    string
			//	statusCode int
			//	now        time.Time
			//}{
			//	traceID:    "000000000000",
			//	statusCode: http.StatusOK,
			//	now:        time.Now(),
			//}
			/*Temp Struct Data */

			v, err := web.GetValues(ctx)
			if err != nil {
				//if fails shutdown gracefully !
				return err
			}

			logger.Infow("request completed",
				"traceid", v.TraceID,
				"method", r.Method,
				"path", r.URL.Path,
				"remote addr", r.RemoteAddr,
				//"status code", v.statusCode,
				//"since", time.Since(v.Now)
			)

			//Execute the Original One when tmp is called
			err = handler(ctx, w, r)

			logger.Infow("request completed",
				"traceid", v.TraceID,
				"method", r.Method,
				"path", r.URL.Path,
				"remote addr", r.RemoteAddr,
				"status code", v.StatusCode,
				"since", time.Since(v.Now),
			)

			return err
		}
		return h
	}
	return m

}
