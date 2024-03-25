package mid

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"service/domain/sys/auth"
	"service/domain/sys/validate"
	"service/foundation/web"
	"strings"
)

func Authenticate(a *auth.Auth) web.MiddlewareFunc {

	m := func(handler web.HandlerFunc) web.HandlerFunc {

		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {

			authStr := r.Header.Get("authorization")

			parts := strings.Split(authStr, " ")
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				err := errors.New("expected authorization header format: bearer <token>")
				return validate.NewRequestError(err, http.StatusUnauthorized)
			}

			claims, err := a.ValidateToken(parts[1])
			if err != nil {
				return validate.NewRequestError(err, http.StatusUnauthorized)
			}

			ctx = auth.SetClaims(ctx, claims)

			//Execute the Original One when tmp is called
			return handler(ctx, w, r)
		}
		return h
	}
	return m

}

func Authorize(roles ...string) web.MiddlewareFunc {

	m := func(handler web.HandlerFunc) web.HandlerFunc {

		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {

			claims, err := auth.GetClaims(ctx)

			if err != nil {
				err := fmt.Errorf("you are not authorized for that action")
				return validate.NewRequestError(err, http.StatusForbidden)
			}

			if !claims.Authorized(roles...) {
				err := fmt.Errorf("you are not authorized for that action claims %v roles %v", claims, roles)
				return validate.NewRequestError(err, http.StatusForbidden)
			}

			//Execute the Original One when tmp is called
			return handler(ctx, w, r)
		}
		return h
	}
	return m

}
