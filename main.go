package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/ardanlabs/conf"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/zipkin"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.uber.org/automaxprocs/maxprocs"
	"go.uber.org/zap"
	defaultLog "log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"service/app/services/sales-api/handlers"
	"service/domain/sys/auth"
	"service/domain/sys/database"
	"service/foundation/keystore"
	"service/foundation/logger"
	"syscall"
	"time"
)

var build = "develop"
var lesson = "12.2"

func main() {
	fmt.Printf("App is Running On %s Env", build)
	defer fmt.Println("Service Ended")

	zapLogger, err := logger.New("sales-api")
	if err != nil {
		defaultLog.Fatalf("Logger Is Down")
	}
	defer zapLogger.Sync()

	ctx := context.Background()
	run(ctx, zapLogger)

	//Catching Signals from k8s or your deployment Environment
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)
	<-shutdown
}

func run(ctx context.Context, log *zap.SugaredLogger) {

	// =================================== GOMAXPROC
	//Sets the Correct Number For The Service
	//based on what is available either by the machine or quotas

	if _, err := maxprocs.Set(); err != nil {
		defaultLog.Fatalf("%s", err.Error())
	}
	log.Infow("startup", "GOMAXPROCS", runtime.GOMAXPROCS(0))

	cfg := struct {
		conf.Version
		Web struct {
			APIHost         string        `conf:"default:0.0.0.0:8000"`
			DebugHost       string        `conf:"default:0.0.0.0:8001"`
			ReadTimeout     time.Duration `conf:"default:5s"`
			WriteTimeout    time.Duration `conf:"default:10s"`
			IdleTimeout     time.Duration `conf:"default:120s"`
			ShutDownTimeout time.Duration `conf:"default:20s"`
		}
		Auth struct {
			KeysFolder string `conf:"default:/zarf/keys/"`
			ActiveKID  string `conf:"default:private.pem"`
		}
		DB struct {
			User         string `conf:"default:postgres"`
			Password     string `conf:"default:postgres,mask"`
			Host         string `conf:"default:localhost"`
			Name         string `conf:"default:postgres"`
			MaxIdleConns int    `conf:"default:0"`
			MaxOpenConns int    `conf:"default:0"`
			DisableTLS   bool   `conf:"default:true"`
		}
		/*
			Zipkin is a distributed tracing system.
			It helps gather timing data needed to troubleshoot latency problems in service architectures.
			It is a visualization of your request-response and trace it.
		*/
		Zipkin struct {
			ReporterURI string  `conf:"default:http://localhost:9411/api/v2/spans"`
			ServiceName string  `conf:"default:sales-api"`
			Probability float64 `conf:"default:0.05"`
		}
	}{
		Version: conf.Version{
			SVN:  build,
			Desc: "Copy Right Stuff",
		},
	}

	const prefix = "SALES"
	help, err := conf.ParseOSArgs(prefix, &cfg)

	if err != nil {
		if errors.Is(err, conf.ErrHelpWanted) {
			fmt.Println(help)
			return
		}
		fmt.Errorf("parse config error: %w", err)
		return
	}

	// =================================== App Starting

	log.Infow("starting service", "version", build)
	defer log.Infow("shutdown")

	out, err := conf.String(&cfg)
	if err != nil {
		fmt.Errorf("config generation failed: %w", err)
		return
	}
	log.Infow("startup", "config", out)

	// =================================== Initialize Authentication Support
	log.Infow("startup", "status", "initializing authentication support")

	ks, err := keystore.NewFs(os.DirFS(cfg.Auth.KeysFolder))
	if err != nil {
		fmt.Errorf("error while reading keys: %w", err)
		return
	}

	newAuth, err := auth.New(cfg.Auth.ActiveKID, ks)
	if err != nil {
		fmt.Errorf("constructing auth: %w", err)
		return
	}

	// =================================== DB Support

	cfgDB := database.Config{
		User:         cfg.DB.User,
		Password:     cfg.DB.Password,
		Host:         cfg.DB.Host,
		Name:         cfg.DB.Name,
		MaxOpenConns: cfg.DB.MaxOpenConns,
		MaxIdleConns: cfg.DB.MaxIdleConns,
		DisableTLS:   cfg.DB.DisableTLS,
	}
	db, err := database.Open(cfgDB)
	if err != nil {
		fmt.Errorf("connecting to db : %w", err)
		return
	}

	defer func() {
		log.Errorw("shutdown", "status", "stopping database support", "host", cfg.DB.Host)
		db.Close()
	}()
	// =================================== Start Trace Support
	log.Infow("startup", "status", "initializing OT/zipkin tracing support")

	traceProvider, err := startTracing(
		cfg.Zipkin.ServiceName,
		cfg.Zipkin.ReporterURI,
		cfg.Zipkin.Probability,
	)

	if err != nil {
		fmt.Errorf("starting traceing system : %w", err)
		return
	}

	defer func() {
		log.Errorw("shutdown", "status", "stopping zipkin")
		traceProvider.Shutdown(context.Background())
	}()

	// =================================== Start Debug Service
	log.Infow("startup", "status", "debug router started", "host", cfg.Web.DebugHost)

	debugMux := handlers.DebugMux(build, log, db)

	go func() {
		if err := http.ListenAndServe(cfg.Web.DebugHost, debugMux); err != nil {
			log.Errorw("startup", "status", "debug router closed unexpectedly", "host", cfg.Web.DebugHost, "Error", err)
		}
	}()

	// -------------------------------------------------------------------------
	// Start API Service

	log.Info(ctx, "startup", "status", "initializing V1 API support")

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	cfgMux := handlers.APIMuxConfig{
		Build:    build,
		Shutdown: shutdown,
		Log:      log,
		Auth:     newAuth,
		DB:       db,
		//Tracer:   tracer,
	}
	apiMux := handlers.AppAPIMux(cfgMux) //, handlers.WithCORS("*"))

	api := http.Server{
		Addr:         cfg.Web.APIHost,
		Handler:      apiMux,
		ReadTimeout:  cfg.Web.ReadTimeout,
		WriteTimeout: cfg.Web.WriteTimeout,
		IdleTimeout:  cfg.Web.IdleTimeout,
		//	ErrorLog:     defaultLog.NewStdLogger(log, logger.LevelError),
	}

	//serverErrors := make(chan error, 1)

	go func() {
		log.Info(ctx, "startup", "status", "api router started", "host", api.Addr)

		//serverErrors <- api.ListenAndServe()
	}()

	/*
		// -------------------------------------------------------------------------
		// Shutdown

		select {
		case err := <-serverErrors:
			return fmt.Errorf("server error: %w", err)

		case sig := <-shutdown:
			log.Info(ctx, "shutdown", "status", "shutdown started", "signal", sig)
			defer log.Info(ctx, "shutdown", "status", "shutdown complete", "signal", sig)

			ctx, cancel := context.WithTimeout(context.Background(), cfg.Web.ShutdownTimeout)
			defer cancel()

			if err := api.Shutdown(ctx); err != nil {
				api.Close()
				return fmt.Errorf("could not stop server gracefully: %w", err)
			}
		}

		return nil */

	return
}

func startTracing(serviceName string, reporterURI string, probability float64) (*trace.TracerProvider, error) {

	/* Capture a small part of traffic with opentelemetry - capturing all would be too much
	zipkin is the tool for viewing them
	*/

	exporter, err := zipkin.New(
		reporterURI,
		//logger
	)
	if err != nil {
		return nil, fmt.Errorf("creating new exporter %w", err)
	}

	traceProvider := trace.NewTracerProvider(
		trace.WithSampler(trace.TraceIDRatioBased(probability)),
		trace.WithBatcher(exporter,
			trace.WithMaxExportBatchSize(trace.DefaultMaxExportBatchSize),
			trace.WithBatchTimeout(trace.DefaultExportTimeout),
			//trace.WithMaxExportBatchSize(trace.DefaultMaxExportBatchSize),
		),
		trace.WithResource(
			resource.NewWithAttributes(
				semconv.SchemaURL,
				semconv.ServiceNameKey.String(serviceName),
				attribute.String("exporter", "zipkin"),
			),
		),
	)
	otel.SetTracerProvider(traceProvider)
	// it uses singleton
	return traceProvider, nil
}
