// Copyright 2021 The webserver Authors. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.
//
// It follows Rob Spike, and Dave Cheney design pattern for options.
//
// - Sensible defaults.
// - Highly configurable.
// - Allows anyone to easily implement their own options.
// - Can grow over time.
// - Self-documenting.
// - Safe for newcomers.
// - Never requires `nil` or an `empty` value to keep the compiler happy.
//
// SEE: https://commandcenter.blogspot.com/2014/01/self-referential-functions-and-design.html
// SEE: https://dave.cheney.net/2014/10/17/functional-options-for-friendly-apis

package webserver

import (
	"time"

	"github.com/gorilla/mux"
	"github.com/saucelabs/sypl/level"
	handler "github.com/saucelabs/webserver/handler"
	"github.com/saucelabs/webserver/metric"
	"github.com/saucelabs/webserver/telemetry"
)

//////
// Consts, vars, and types.
//////

// Option allows to define options for the Server.
type Option func(s *Server)

//////
// Server.
//////

// WithRouter sets the base router.
func WithRouter(router *mux.Router) Option {
	return func(s *Server) {
		s.router = router
	}
}

// WithTimeout sets the maximum duration for each individual timeouts.
func WithTimeout(read, request, inflight, tasks, write time.Duration) Option {
	return func(s *Server) {
		s.Timeout.ReadTimeout = read
		s.Timeout.RequestTimeout = request
		s.Timeout.ShutdownInFlightTimeout = inflight
		s.Timeout.ShutdownTaskTimeout = tasks
		s.Timeout.WriteTimeout = write
	}
}

//////
// Telemetry.
//////

// WithTelemetry sets telemetry.
//
// NOTE: Use `telemetry.New` to bring your own telemetry.
//
// SEE: https://opentelemetry.io/vendors
func WithTelemetry(t *telemetry.Telemetry) Option {
	return func(s *Server) {
		s.EnableTelemetry = true

		s.telemetry = t
	}
}

// WithoutTelemetry disables telemetry.
func WithoutTelemetry() Option {
	return func(s *Server) {
		s.EnableTelemetry = false
	}
}

//////
// Metrics.
//////

// WithMetrics adds metrics to the list of pre-loaded metrics.
//
// NOTE: Use `metric.New` to bring your own metric.
func WithMetrics(metrics ...metric.Metric) Option {
	return func(s *Server) {
		s.EnableMetrics = true

		for _, m := range metrics {
			metric.Publish(m.Name, m.Value)
		}
	}
}

// WithMetricsFunc provides a quick way to add metrics.
func WithMetricsFunc(name string, v interface{}) Option {
	return func(s *Server) {
		s.EnableMetrics = true

		metric.Publish(name, metric.Func(func() interface{} {
			return v
		}))
	}
}

// WithoutMetrics disables metrics.
func WithoutMetrics() Option {
	return func(s *Server) {
		s.EnableMetrics = false
	}
}

//////
// Logging.
//////

// WithLogging sets logging configuration.
//
// NOTE: Set filepath to "" to disabled that.
func WithLogging(console, request, filepath string) Option {
	return func(s *Server) {
		s.Logging.ConsoleLevel = console
		s.Logging.RequestLevel = request
		s.Logging.Filepath = filepath
	}
}

// WithoutLogging disables logging.
func WithoutLogging() Option {
	return func(s *Server) {
		s.Logging.ConsoleLevel = level.None.String()
		s.Logging.RequestLevel = level.None.String()
		s.Logging.Filepath = ""
	}
}

//////
// Handlers.
//////

// WithReadiness adds server readiness. Multiple readinesses determiners can be
// passed. In this case, only if ALL are ready, the server will be considered
// ready.
//
// NOTE: Use `handler.NewReadinessDeterminer` to bring your own determiner.
func WithReadiness(readinessState ...*handler.ReadinessDeterminer) Option {
	return func(s *Server) {
		s.handlers = append(s.handlers, handler.Readiness(readinessState...))
	}
}

// WithHandlers adds handlers to the list of pre-loaded handlers.
//
// NOTE: Use `handler.New` to bring your own handler.
func WithHandlers(handlers ...handler.Handler) Option {
	return func(s *Server) {
		addHandler(s.GetRouter(), handlers...)
	}
}

// WithoutHandlers disable the default pre-loaded handlers:
// - Liveness handler (`GET /liveness`)
// - OK handler (`GET /`)
// - Stop handler (`GET /stop`).
func WithoutHandlers() Option {
	return func(s *Server) {
		s.handlers = []handler.Handler{}
	}
}
