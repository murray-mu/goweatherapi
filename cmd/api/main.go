package main

import (
	"context"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/goware/cors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"testapi/goweatherapi/city"
	"testapi/goweatherapi/cmd/internal/flag"
	"testapi/goweatherapi/cmd/internal/health"
	"testapi/goweatherapi/handler"
	"testapi/goweatherapi/log"
	"testapi/goweatherapi/weather"
)

type CliFlags struct {
	DocsPath string `long:"docs-path" env:"API_DOCS_PATH" default:"docs" description:"Path to documentation folder."`

	HTTP struct {
		Addr           string   `long:"addr" env:"API_HTTP_ADDR" default:"0.0.0.0:8088" description:"HTTP service address."`
		AllowedOrigins []string `long:"allowed-origins" env:"API_ALLOWED_ORIGINS" description:"The list of origins a cross-domain request can be executed from."`
		AllowedHeaders []string `long:"allowed-headers" env:"API_ALLOWED_HEADERS" description:"The list of non simple headers the client is allowed to use with cross-domain requests."`
		ExposedHeaders []string `long:"exposed-headers" env:"API_EXPOSED_ORIGINS" description:"The list which indicates which headers are safe to expose."`
	}

	Log struct {
		Level  string `long:"log-level" default:"info" choice:"debug" choice:"info" choice:"warn" choice:"error" env:"API_LOG_LEVEL" description:"Log level."`
		Format string `long:"log-format" default:"text" choice:"text" choice:"json" env:"API_LOG_FORMAT" description:"Log format."`
	}

	PrintVersion bool `long:"version" description:"Show application version"`
}

func main() {
	var cfg CliFlags
	flag.ParseFlags(&cfg)

	logger := log.New(cfg.Log.Format, cfg.Log.Level, os.Stdout)
	logger.Infof("started with config: %+v", cfg)

	cm := cors.New(cors.Options{
		AllowedOrigins:   cfg.HTTP.AllowedOrigins,
		AllowedHeaders:   cfg.HTTP.AllowedHeaders,
		ExposedHeaders:   cfg.HTTP.ExposedHeaders,
		AllowedMethods:   []string{http.MethodGet, http.MethodPut, http.MethodDelete},
		AllowCredentials: true,
	})
	// note: order of middlewares is important
	r := chi.NewRouter()
	r.Use(
		middleware.RequestID,
		middleware.RealIP,
		handler.RequestLogger(logger),
		middleware.Recoverer,
		cm.Handler,
	)

	r.Mount("/readiness", health.Routes())
	r.Route("/api", func(r chi.Router) {
		r.Use(handler.ApiVersion("1.0"))
		r.Mount("/city", weather.Routes(""))
		r.Mount("/cities", city.Routes(""))
	})
	handler.FileServer(r, "/docs", http.Dir(cfg.DocsPath))
	srv := &http.Server{Addr: cfg.HTTP.Addr, Handler: r}

	sigquit := make(chan os.Signal, 1)
	signal.Ignore(syscall.SIGHUP, syscall.SIGPIPE)
	signal.Notify(sigquit, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		s := <-sigquit
		logger.Infof("captured %v, exiting...", s)

		health.SetReadinessStatus(http.StatusServiceUnavailable)

		logger.Info("gracefully shutdown server")
		if err := srv.Shutdown(context.Background()); err != nil {
			logger.WithError(err).Error("could not shutdown server")
		}
	}()

	logger.Info("starting http service...")
	logger.Infof("listening on %s", cfg.HTTP.Addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.WithError(err).Error("server error")
	}
}

