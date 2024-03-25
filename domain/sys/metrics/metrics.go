package metrics

import (
	"context"
	"expvar"
)

var m *metric

// Metric - these values are thread safe - thanks to expvar package
type metric struct {
	Goroutines *expvar.Int
	Requests   *expvar.Int
	Errors     *expvar.Int
	Panics     *expvar.Int
}

// to make sure, all of these called once
func init() {

	m = &metric{
		Goroutines: expvar.NewInt("goroutines"),
		Requests:   expvar.NewInt("requests"),
		Errors:     expvar.NewInt("errors"),
		Panics:     expvar.NewInt("panics"),
	}
}

type ctxKey int

const key ctxKey = 1

// Set - set metrics data in to context
func Set(ctx context.Context) context.Context {
	return context.WithValue(ctx, key, m)
}

func AddGoroutines(ctx context.Context) error {
	if v, ok := ctx.Value(key).(*metric); ok {
		if v.Goroutines.Value()%100 == 0 {
			v.Goroutines.Add(1)
		}
	}
	return nil
}

func AddRequests(ctx context.Context) error {
	if v, ok := ctx.Value(key).(*metric); ok {
		v.Requests.Add(1)
	}
	return nil
}

func AddPanics(ctx context.Context) error {
	if v, ok := ctx.Value(key).(*metric); ok {
		v.Panics.Add(1)
	}
	return nil
}

func AddErrors(ctx context.Context) error {
	if v, ok := ctx.Value(key).(*metric); ok {
		v.Errors.Add(1)
	}
	return nil
}
