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
	"expvar"
	"time"

	handler "github.com/saucelabs/webserver/handler"
	"github.com/saucelabs/webserver/telemetry"
)

// Option allows to define options for the Server.
type Option func(s *Server)

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

// WithTelemetry sets telemetry.
//
// NOTE: Use `telemetry.New` to bring your own telemetry.
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

// WithReadiness sets server readiness. Returning any non-nil error means server
// isn't ready.
func WithReadiness(readinessFunc handler.ReadinessFunc) Option {
	return func(s *Server) {
		s.preLoadedHandlers = append(s.preLoadedHandlers, handler.Readiness(readinessFunc))
	}
}

// WithHandlers adds a handler to the pre-loaded handlers.
//
// NOTE: Use `handler.New` to add handlers
func WithHandlers(handlers ...handler.Handler) Option {
	return func(s *Server) {
		addHandler(s.GetRouter(), handlers)
	}
}

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
