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

	"github.com/saucelabs/sypl/level"
	handler "github.com/saucelabs/webserver/handler"
	"github.com/saucelabs/webserver/internal/expvar"
	"github.com/saucelabs/webserver/telemetry"
)

//////
// Const, vars, and types.
//////

// Option allows to define options for the Server.
type Option func(s *Server)

//////
// Timeout.
//////

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

// WithMetricsRaw allows to publishes metrics based on exp vars. It's useful for
// cases such as counters. It gives full control over what's being exposed.
func WithMetricsRaw(name string, metrics expvar.Var) Option {
	return func(s *Server) {
		expvar.Publish(name, metrics)
	}
}

// WithMetrics provides a quick way to publish static metric values.
func WithMetrics(name string, v interface{}) Option {
	return func(s *Server) {
		expvar.Publish(name, expvar.Func(func() interface{} {
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

// WithLoggingOptions sets logging configuration.
//
// NOTE: Set filepath to "" to disabled that.
func WithLoggingOptions(console, request, filepath string) Option {
	return func(s *Server) {
		s.Logging.ConsoleLevel = console
		s.Logging.RequestLevel = request
		s.Logging.Filepath = filepath
	}
}

// WithoutLogging() disables logging.
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

// WithReadiness sets server readiness. Returning any non-nil error means server
// isn't ready.
func WithReadiness(readinessState *handler.ReadinessState) Option {
	return func(s *Server) {
		s.preLoadedHandlers = append(s.preLoadedHandlers, handler.Readiness(readinessState))
	}
}

// WithPreLoadedHandlers adds handlers to the list of pre-loaded handlers.
//
// NOTE: Use `handler.New` to bring your own handler.
func WithPreLoadedHandlers(handlers ...handler.Handler) Option {
	return func(s *Server) {
		addHandler(s.GetRouter(), handlers...)
	}
}

// WithoutPreLoadedHandlers disable the default pre-loaded handlers:
// - OK handler (`GET /`)
// - Liveness handler (`GET /liveness`)
// - Readiness handler (`GET /readiness`)
// - Stop handler (`GET /stop`)
// - Metrics handler (`GET /debug/vars`).
func WithoutPreLoadedHandlers() Option {
	return func(s *Server) {
		s.preLoadedHandlers = []handler.Handler{}
	}
}
