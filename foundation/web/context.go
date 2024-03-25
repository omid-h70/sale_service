package web

import (
	"context"
	"errors"
	"time"
)

type ctxKey int

const key ctxKey = 1

type Value struct {
	TraceID    string
	Now        time.Time
	StatusCode int
}

func GetValues(ctx context.Context) (*Value, error) {
	v, ok := ctx.Value(key).(*Value)
	if !ok {
		return nil, errors.New("values is missing from context")
	}
	return v, nil
}

func GetTraceID(ctx context.Context) string {
	v, ok := ctx.Value(key).(*Value)
	if !ok {
		return "00000000-0000-0000-0000-000000000000"
	}
	return v.TraceID
}

func SetStatusCode(ctx context.Context, statusCode int) error {
	v, ok := ctx.Value(key).(*Value)
	if !ok {
		return errors.New("values is missing from context")
	}
	v.StatusCode = statusCode
	return nil
}
