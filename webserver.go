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
	"github.com/saucelabs/webserver/internal/logger"
	"github.com/saucelabs/webserver/internal/validation"
	"github.com/saucelabs/webserver/telemetry"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux"

	handler "github.com/saucelabs/webserver/handler"
)

//////
// Const, and vars.
//////

const (
	defaultTimeout             = 3 * time.Second
	defaultRequestTimeout      = 1 * time.Second
	defaultShutdownTaskTimeout = 10 * time.Second
)

//////
// Interfaces.
//////

// IServer defines what a server does.
type IServer interface {
	// GetLogger retuns the server logger.
	GetLogger() sypl.ISypl

	// GetRouter retuns the server router.
	GetRouter() *mux.Router

	// Start the server.
	Start() error
}

//////
// Definitions.
//////

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

	// Controls whether telemetry are enable, or not, default: true.
	EnableTelemetry bool `json:"enable_telemetry"`

	// Timeouts fine-control.
	*Timeout `json:"timeout" validate:"required"`

	// Logger powered by Sypl.
	logger *sypl.Sypl `json:"-" validate:"required"`

	// Handlers added, and configured before the server starts.
	preLoadedHandlers []handler.Handler `json:"-"`

	// Router powered by Gorilla Mux.
	router *mux.Router `json:"-" validate:"required"`

	// Golang's http server.
	server http.Server `json:"-" validate:"required"`

	// Telemetry powered by OpenTelemetry.
	telemetry telemetry.ITelemetry `json:"-"`
}

//////
// IServer implementation.
//////

// GetLogger retuns the server logger.
func (s *Server) GetLogger() sypl.ISypl {
	return s.logger
}

// GetRouter retuns the server router.
func (s *Server) GetRouter() *mux.Router {
	return s.router
}

// GetTelemetry retuns telemetry.
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
			customerror.NewFailedToError(
				"finish request, timed out",
				customerror.WithStatusCode(http.StatusRequestTimeout),
			).Error(),
		),

		// Best practice setting timeouts. It avoid "Slowloris" attacks.
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
		s.GetLogger().Tracelnf("Got %s signal, gracefuly shutting down", sig)
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
		} else {
			// Wait for tasks such as flush cache and files, and telemetry.
			s.GetLogger().Tracelnf("Waiting %s for tasks, %s", s.ShutdownTaskTimeout, crtlCmsg)

			time.Sleep(s.ShutdownTaskTimeout)
		}

		// If reaches here, error can be safely collected.
		return <-serverErr
	}
}

// New is the web server factory.
func New(
	name, address string,
	opts ...Option,
) (IServer, error) {
	s := &Server{
		Address:           address,
		EnableTelemetry:   true,
		logger:            logger.Setup("webserver", "trace", "").New(name),
		Name:              name,
		preLoadedHandlers: []handler.Handler{handler.OK(), handler.Liveness(), handler.ExpVar(), handler.Stop()},
		router:            mux.NewRouter(),
		Timeout: &Timeout{
			ReadTimeout:             defaultTimeout,
			RequestTimeout:          defaultRequestTimeout,
			ShutdownInFlightTimeout: defaultTimeout,
			ShutdownTaskTimeout:     defaultShutdownTaskTimeout,
			WriteTimeout:            defaultTimeout,
		},
	}

	//////
	// Options processing.
	//////

	for _, opt := range opts {
		opt(s)
	}

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

	addHandler(s.GetRouter(), s.preLoadedHandlers)

	//////
	// Server metrics.
	//////

	publishServerMetrics(s)

	return s, nil
}
