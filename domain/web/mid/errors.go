package mid

import (
	"context"
	"go.uber.org/zap"
	"net/http"
	"service/domain/sys/validate"
	"service/foundation/web"
)

func Errors(logger *zap.SugaredLogger) web.MiddlewareFunc {

	m := func(handler web.HandlerFunc) web.HandlerFunc {

		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {

			v, err := web.GetValues(ctx)
			if err != nil {
				return web.NewShutdownError("shutdown requested from errors mid")
			}

			//Execute the Original One when tmp is called
			err = handler(ctx, w, r)
			if err != nil {
				logger.Errorw("ERROR", "traceID", v.TraceID, "Error", err)

				var er validate.ErrorResponse
				var status int

				switch act := validate.Cause(err).(type) {
				//its not pointer because, its a slice already
				case validate.FieldErrors:
					er = validate.ErrorResponse{
						Error:  "data validation error",
						Fields: act.Error(),
					}
					status = http.StatusBadRequest
				case *validate.RequestError:
					er = validate.ErrorResponse{
						Fields: act.Error(),
					}
					status = act.Status
				default:
					er = validate.ErrorResponse{
						Fields: http.StatusText(http.StatusInternalServerError),
					}
					status = http.StatusInternalServerError
				}

				if err := web.Respond(ctx, w, status, er); err != nil {
					return err
				}

				if ok := web.IsShutDownError(err); ok {
					return err
				}
			}
			//Error Got Handled, No need to propagate it
			return nil
		}
		return h
	}
	return m

}
