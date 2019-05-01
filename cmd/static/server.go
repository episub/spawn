package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/handler"
	"github.com/caarlos0/env"
	api "{{.}}/graph"
	"{{.}}/resolvers"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/opentracing-contrib/go-stdlib/nethttp"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	jaegerConfig "github.com/uber/jaeger-client-go/config"
)

type config struct {
	ExternalPort int    `env:"PORT" envDefault:"8080"`
	InternalPort int    `env:"INTERNAL_PORT" envDefault:"8585"`
	Debug        bool   `env:"DEBUG" envDefault:"false"`
	DBName       string `env:"DB_NAME"`
	DebugSpans   bool   `env:"DEBUG_SPANS" envDefault:"false"`
	DBUser       string `env:"DB_USER"`
	DBPass       string `env:"DB_PASS"`
	DBHost       string `env:"DB_HOST"`
}

var cfg config
var log = logrus.New()

func main() {
	err := env.Parse(&cfg)

	if err != nil {
		log.Fatal(err)
	}

	tracer, closer := initJaeger("graphql")
	if closer != nil {
		defer closer.Close()
	}

	// StartSpanFromContext uses the global tracer, so we need to set it here to
	// be our jaeger tracer
	opentracing.SetGlobalTracer(tracer)

	startRouters(tracer)
}

// jaegerLogger Logger to wrap logrus so we can use with Jaeger
type jaegerLogger struct {
	logger *logrus.Logger
}

func (j *jaegerLogger) Error(msg string) {
	j.logger.Errorf(msg)
}

func (j *jaegerLogger) Infof(msg string, args ...interface{}) {
	j.logger.Infof(msg, args...)
}

func initJaeger(service string) (opentracing.Tracer, io.Closer) {
	// Grab config from environment:
	jcfg, err := jaegerConfig.FromEnv()
	if err != nil {
		log.WithField("error", err).Error("Failed to configure jaeger")
	}
	jcfg.Sampler.Type = "const"
	jcfg.Sampler.Param = 1
	jcfg.Reporter.LogSpans = cfg.DebugSpans

	log.WithField("hostport", jcfg.Reporter.LocalAgentHostPort).Printf("Jaeger connected")

	jl := jaegerLogger{log}

	tracer, closer, err := jcfg.New(service, jaegerConfig.Logger(&jl))
	if err != nil {
		log.WithField("error", err).Error("Cannot init Jaeger, so setting to localhost")
		jcfg.Reporter.LocalAgentHostPort = "127.0.0.1:6831"
		tracer, closer, _ = jcfg.New(service, jaegerConfig.Logger(&jl))
	}

	return tracer, closer
}

func startRouters(tracer opentracing.Tracer) {
	internalRouter := newRouter(tracer)
	internalRouter.Get("/health", healthHandler)
	internalRouter.Get("/live", liveHandler)
	internalRouter.Handle("/metrics", promhttp.Handler())

	externalRouter := newRouter(tracer)
	externalRouter.Handle("/", handler.Playground("GraphQL playground", "/query"))
	externalRouter.Route("/query", func(r chi.Router) {
		r.Use(middleware.Timeout(60 * time.Second))
		r.Handle("/", handler.GraphQL(
			api.NewExecutableSchema(graphqlConfig()),
			handler.RequestMiddleware(requestMiddleware()),
		))
	})

	log.Println(fmt.Sprintf("connect to http://localhost:%d/ for graphql playground", cfg.ExternalPort))
	go func() { log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", cfg.ExternalPort), externalRouter)) }()
	log.Println(fmt.Sprintf("internal endpoints available on port %d", cfg.InternalPort))
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", cfg.InternalPort), internalRouter))
}

// newRouter returns a new router with all default values set
func newRouter(tracer opentracing.Tracer) chi.Router {
	router := chi.NewRouter()
	router.Use(Opentracing(tracer))

	return router
}

// graphqlConfig Returns config for gqlgen graphql handler
func graphqlConfig() api.Config {
	c := api.Config{Resolvers: &resolvers.Resolver{}}
	return c
}

// Opentracing Adds opentracing to context
func Opentracing(tracer opentracing.Tracer) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return nethttp.Middleware(tracer, next)
	}
}

// liveHandler Returns true when the service is live and ready to receive requests
func liveHandler(w http.ResponseWriter, r *http.Request) {
	log.Info("Liveness request received")
	w.WriteHeader(http.StatusOK)
}

// healthHandler Returns true when the service is live and ready to receive requests
func healthHandler(w http.ResponseWriter, r *http.Request) {
	log.Info("Health request received")
	w.WriteHeader(http.StatusOK)
}

func requestMiddleware() graphql.RequestMiddleware {
	return func(ctx context.Context, next func(ctx context.Context) []byte) []byte {
		requestContext := graphql.GetRequestContext(ctx)
		span, ctx := opentracing.StartSpanFromContext(ctx, requestContext.RawQuery)
		defer span.Finish()
		ext.SpanKind.Set(span, "server")
		ext.Component.Set(span, "gqlgen")

		res := next(ctx)

		return res
	}
}
