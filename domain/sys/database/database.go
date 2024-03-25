package database

//database package is a convenient package, it's ok to use it no more abstraction above it !

import (
	"context"
	_ "database/sql"
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
	"net/url"
	"reflect"
	"service/foundation/web"
	"strings"
	"time"
)

var (
	ErrNotFound              = errors.New("not found")
	ErrInvalidID             = errors.New("invalid id")
	ErrAuthenticationFailure = errors.New("authentication failed")
	ErrForbidden             = errors.New("forbidden")
)

type Config struct {
	User         string
	Password     string
	Host         string
	Name         string
	MaxIdleConns int
	MaxOpenConns int
	DisableTLS   bool
}

func Open(cfg Config) (*sqlx.DB, error) {
	sslMode := "require"
	if cfg.DisableTLS {
		sslMode = "disable"
	}

	q := make(url.Values)
	q.Set("sslmode", sslMode)
	q.Set("timezone", "utc")

	u := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(cfg.User, cfg.Password),
		Host:     cfg.Host,
		Path:     cfg.Name,
		RawQuery: q.Encode(),
	}

	fmt.Println("ul >>>> ", u.String())
	db, err := sqlx.Open("postgres", u.String())
	if err != nil {
		return nil, err
	}
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	return db, nil
}

func StatusCheck(ctx context.Context, db *sqlx.DB) error {

	for attempts := 1; ; attempts++ {
		if db.Ping() == nil {
			break
		}
		time.Sleep(time.Duration(attempts) * 100 * time.Millisecond)

		if ctx.Err() != nil {
			return ctx.Err()
		}
	}

	var result bool
	const q = "select true"
	return db.QueryRowContext(ctx, q).Scan(&result)
}

func NamedExecContext(ctx context.Context, logger *zap.SugaredLogger, db *sqlx.DB, query string, data any) error {

	q := queryString(query, data)
	logger.Infow("database.NamedExecContext", "traceID", web.GetTraceID(ctx), "query", q)

	ctx, span := otel.GetTracerProvider().Tracer("").Start(ctx, "database.query")
	span.SetAttributes(attribute.String("query", q))
	defer span.End()

	if _, err := db.NamedExecContext(ctx, query, data); err != nil {
		return err
	}
	return nil
}

func NamedQuerySlice(ctx context.Context, logger *zap.SugaredLogger, db *sqlx.DB, query string, data any, dest any) error {
	q := queryString(query, data)
	logger.Infow("database.NamedQuerySlice", "traceID", web.GetTraceID(ctx), "query", q)

	ctx, span := otel.GetTracerProvider().Tracer("").Start(ctx, "database.query")
	span.SetAttributes(attribute.String("query", q))
	defer span.End()

	val := reflect.ValueOf(dest)
	if val.Kind() != reflect.Ptr && val.Elem().Kind() != reflect.Slice {
		return errors.New("must provide a pointer to a slice")
	}

	rows, err := db.NamedQueryContext(ctx, query, data)
	if err != nil {
		return err
	}

	slice := val.Elem()
	for rows.Next() {
		v := reflect.New(slice.Type().Elem())
		err := rows.StructScan(v.Interface())
		if err != nil {
			return err
		}
		slice.Set(reflect.Append(slice, v.Elem()))
	}
	return nil
}

func NamedQueryStruct(ctx context.Context, logger *zap.SugaredLogger, db *sqlx.DB, query string, data any, dest any) error {
	q := queryString(query, data)
	logger.Infow("database.NamedQueryStruct", "traceID", web.GetTraceID(ctx), "query", q)

	ctx, span := otel.GetTracerProvider().Tracer("").Start(ctx, "database.query")
	span.SetAttributes(attribute.String("query", q))
	defer span.End()

	rows, err := db.NamedQueryContext(ctx, query, data)
	if err != nil {
		return err
	}

	if !rows.Next() {
		return ErrNotFound
	}

	if err := rows.StructScan(dest); err != nil {
		return err
	}

	return nil
}

func queryString(query string, args ...any) string {
	query, params, err := sqlx.Named(query, args)
	if err != nil {
		return err.Error()
	}

	//loop through and replace question mark with param
	for _, param := range params {

		var value string
		switch v := param.(type) {
		case string:
			value = fmt.Sprintf("%q", v)
		case []byte:
			value = fmt.Sprintf("%q", string(v))
		default:
			value = fmt.Sprintf("%v", v)
		}
		query = strings.Replace(query, "?", value, 1)
	}
	query = strings.ReplaceAll(query, "\t", "")
	query = strings.ReplaceAll(query, "\n", " ")
	return strings.Trim(query, " ")
}
