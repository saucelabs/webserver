// Copyright 2021 The webserver Authors. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package webserver

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/saucelabs/customerror"
	"github.com/saucelabs/sypl"
	"github.com/saucelabs/sypl/level"
	handler "github.com/saucelabs/webserver/handler"
	"github.com/saucelabs/webserver/internal/expvar"
	"github.com/saucelabs/webserver/internal/logger"
	"github.com/saucelabs/webserver/internal/middleware"
	"github.com/saucelabs/webserver/internal/validation"
	"github.com/saucelabs/webserver/telemetry"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux"
)

//////
// Const, and vars.
//////

const (
	defaultTimeout             = 3 * time.Second
	defaultRequestTimeout      = 1 * time.Second
	defaultShutdownTaskTimeout = 10 * time.Second
	frameworkName              = "webserver"
)

// ErrRequesTimeout indicates a request failed to finish, it timed out.
var ErrRequesTimeout = customerror.NewFailedToError(
	"finish request, timed out",
	customerror.WithStatusCode(http.StatusRequestTimeout),
)

//////
// Interfaces.
//////

// IServer defines what a server does.
type IServer interface {
	// GetLogger returns the server logger.
	GetLogger() sypl.ISypl

	// GetRouter returns the server router.
	GetRouter() *mux.Router

	GetTelemetry() telemetry.ITelemetry

	// Start the server.
	Start() error
}

//////
// Definitions.
//////

// Logging definition.
type Logging struct {
	// ConsoleLevel defines the level for the `Console` output.
	ConsoleLevel string `json:"console_level" validate:"required,gte=3,oneof=none fatal error info warn debug trace"`

	// RequestLevel defines the level for logging requests.
	RequestLevel string `json:"request_level" validate:"required,gte=3,oneof=none fatal error info warn debug trace"`

	// Filepath is the file path to optionally write logs.
	Filepath string `json:"filepath" validate:"omitempty,gte=3"`
}

// Timeout definition.
type Timeout struct {
	// ReadTimeout max duration for READING the entire request, including the
	// body, default: 3s.
	ReadTimeout time.Duration `json:"read_timeout"`

	// RequestTimeout max duration to WAIT BEFORE CANCELING A REQUEST,
	// default: 1s.
	//
	// NOTE: It's automatically validated against other timeouts, and needs to
	// be smaller.
	RequestTimeout time.Duration `json:"request_timeout" validate:"ltfield=ReadTimeout"`

	// ShutdownInFlightTimeout max duration to WAIT IN-FLIGHT REQUESTS,
	// default: 3s.
	ShutdownInFlightTimeout time.Duration `json:"shutdown_in_flight_timeout"`

	// ShutdownTaskTimeout max duration TO WAIT for tasks such as flush cache,
	// files, and telemetry, default: 10s.
	ShutdownTaskTimeout time.Duration `json:"shutdown_task_timeout"`

	// ShutdownTimeout max duration for WRITING the response, default: 3s.
	WriteTimeout time.Duration `json:"write_timeout"`
}

// Server definition.
type Server struct {
	// Address is a TCP address to listen on, default: ":4446".
	Address string `json:"address" validate:"tcp_addr"`

	// Name of the server.
	Name string `json:"name" validate:"required,gte=3"`

	// EnableMetrics controls whether metrics are enable, or not, default: true.
	EnableMetrics bool `json:"enable_metrics"`

	// EnableTelemetry controls whether telemetry are enable, or not,
	// default: true.
	EnableTelemetry bool `json:"enable_telemetry"`

	// Logging fine-control.
	*Logging `json:"logging" validate:"required"`

	// Timeouts fine-control.
	*Timeout `json:"timeout" validate:"required"`

	// Logger powered by Sypl.
	logger *sypl.Sypl `json:"-" validate:"required"`

	// Handlers added, and configured before the server starts.
	preLoadedHandlers []handler.Handler `json:"-"`

	// Router powered by Gorilla Mux.
	router *mux.Router `json:"-" validate:"required"`

	// HTTP server powered by Golang's built-in http server.
	server http.Server `json:"-" validate:"required"`

	// Telemetry powered by OpenTelemetry.
	telemetry telemetry.ITelemetry `json:"-"`
}

//////
// IServer implementation.
//////

// GetLogger returns the server logger.
func (s *Server) GetLogger() sypl.ISypl {
	return s.logger
}

// GetRouter returns the server router.
func (s *Server) GetRouter() *mux.Router {
	return s.router
}

// GetTelemetry returns telemetry.
func (s *Server) GetTelemetry() telemetry.ITelemetry {
	return s.telemetry
}

// Start the server.
func (s *Server) Start() error {
	// Instantiate the underlying HTTP server.
	s.server = http.Server{
		Addr: s.Address,
		Handler: http.TimeoutHandler(
			s.GetRouter(),
			s.Timeout.RequestTimeout,
			ErrRequesTimeout.Error(),
		),

		// Best practice setting timeouts. It avoid "slowloris" attacks.
		ReadTimeout:  s.Timeout.ReadTimeout,
		WriteTimeout: s.Timeout.WriteTimeout,
	}

	serverErr := make(chan error, 1)

	// Non-blocking server start up.
	go func() {
		s.GetLogger().Debuglnf("server is about to start @ %s", s.Address)
		serverErr <- s.server.ListenAndServe()
	}()

	// Listen for "catchable" OS signals, forget SIGKILL...
	osSignals := make(chan os.Signal, 1)
	signal.Notify(osSignals, os.Interrupt, syscall.SIGTERM)

	// Block execution, and listen for any server errors (e.g.: "port in use"),
	// or OS signals.
	select {
	// These errors don't require graceful shutdown.
	case err := <-serverErr:
		return err
	case sig := <-osSignals:
		const crtlCmsg = "press ctrl+c to stop anyway"

		s.logger.PrintNewLine()
		s.GetLogger().Tracelnf("Got %s signal, gracefully shutting down", sig)
		s.GetLogger().Tracelnf("Waiting %s for inflight requests to finish, %s", s.ShutdownInFlightTimeout, crtlCmsg)

		// Let Go terminate the program if we get that signal again.
		signal.Reset(sig)

		ctx, cancel := context.WithTimeout(context.Background(), s.ShutdownInFlightTimeout)
		defer cancel()

		var shutdownErr error

		s.server.SetKeepAlivesEnabled(false)

		// Attempt to gracefully shutdown by closing the listener, waiting the
		// completion of all inflight requests.
		if err := s.server.Shutdown(ctx); err != nil {
			if isTimeoutError(shutdownErr) {
				shutdownErr = customerror.NewFailedToError(
					"gracefully shutdown, timeout reached. Stopping hard...",
					customerror.WithError(err),
				)
			} else {
				shutdownErr = err
			}

			// Well.. KIH: Kill It Hard.
			if err := s.server.Close(); err != nil {
				shutdownErr = customerror.NewFailedToError(
					"hardly shutdown the server",
					customerror.WithError(err),
				)
			}
		}

		if shutdownErr != nil {
			return shutdownErr
		}

		// Wait for tasks such as flush cache and files, and telemetry.
		s.GetLogger().Tracelnf("Waiting %s for tasks, %s", s.ShutdownTaskTimeout, crtlCmsg)

		time.Sleep(s.ShutdownTaskTimeout)

		// If reaches here, error can be safely collected.
		return <-serverErr
	}
}

//////
// Factory.
//////

// New is the web server factory. It returns a web server with observability:
// - Metrics: `cmdline`, `memstats`, and `server`.
// - Telemetry: `stdout` exporter.
// - Logging: `error`, no file.
// - Pre-loaded handlers (Liveness, OK, and Stop).
func New(
	name, address string,
	opts ...Option,
) (IServer, error) {
	s := &Server{
		Address:         address,
		EnableMetrics:   true,
		EnableTelemetry: true,
		Logging: &Logging{
			ConsoleLevel: level.Error.String(),
			RequestLevel: level.Error.String(),
			Filepath:     "",
		},
		Name: name,
		Timeout: &Timeout{
			ReadTimeout:             defaultTimeout,
			RequestTimeout:          defaultRequestTimeout,
			ShutdownInFlightTimeout: defaultTimeout,
			ShutdownTaskTimeout:     defaultShutdownTaskTimeout,
			WriteTimeout:            defaultTimeout,
		},

		preLoadedHandlers: []handler.Handler{handler.OK(), handler.Liveness(), handler.Stop()},
		router:            mux.NewRouter(),
	}

	//////
	// Options processing.
	//////

	for _, opt := range opts {
		opt(s)
	}

	//////
	// Logging.
	//////

	s.logger = logger.Setup(
		frameworkName,
		s.Logging.ConsoleLevel,
		s.Logging.RequestLevel,
		s.Logging.Filepath,
	).New(name)

	s.router.Use(middleware.Logger(s.logger))

	//////
	// Telemetry.
	//////

	if s.EnableTelemetry {
		if s.GetTelemetry() == nil {
			defaultTelemetry, err := telemetry.NewDefault(name)
			if err != nil {
				return nil, err
			}

			s.telemetry = defaultTelemetry
		}

		s.GetRouter().Use(otelmux.Middleware(name))
	}

	//////
	// Validation.
	//////

	if err := validation.ValidateStruct(s); err != nil {
		return nil, err
	}

	//////
	// Load handlers.
	//////

	addHandler(s.GetRouter(), s.preLoadedHandlers...)

	//////
	// Server metrics.
	//////

	if s.EnableMetrics {
		// Publish Golang's metrics: cmdline, and memstats.
		expvar.PublishCmdLine()
		expvar.PublishMemStats()

		// Publish specific server's information.
		publishServerMetrics(s)

		// Gorilla Mux exp var route registration.
		addHandler(s.GetRouter(), handler.ExpVar())
	}

	return s, nil
}

// NewBasic returns a basic web server without observability:
// - Metrics
// - Telemetry
// - Logging
// - Pre-loaded handlers (Liveness, Readiness, OK, and Stop).
func NewBasic(name, address string, opts ...Option) (IServer, error) {
	// Merge default options with new ones (`opts`).
	finalOpts := append([]Option{
		WithoutMetrics(),
		WithoutTelemetry(),
		WithoutLogging(),
		WithoutPreLoadedHandlers(),
	}, opts...)

	return New(name, address, finalOpts...)
}
