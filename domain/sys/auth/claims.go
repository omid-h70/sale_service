package auth

import (
	"context"
	"errors"
	"github.com/golang-jwt/jwt/v4"
)

const (
	RoleAdmin = "ADMIN"
	RoleUser  = "USER"
)

// Claims represents the authorization claims transmitted via a JWT.
type Claims struct {
	jwt.StandardClaims
	Roles []string `json:"roles"`
}

func (c Claims) Authorized(roles ...string) bool {
	for _, role := range c.Roles {
		for _, want := range roles {
			if role == want {
				return true
			}
		}
	}
	return false
}

type ctxKey int

const key ctxKey = 1

// SetClaims - set metrics data in to context
func SetClaims(ctx context.Context, claims Claims) context.Context {
	return context.WithValue(ctx, key, claims)
}

func GetClaims(ctx context.Context) (Claims, error) {
	v, ok := ctx.Value(key).(Claims)
	if !ok {
		return Claims{}, errors.New("claims not found in context")
	}
	return v, nil
}
