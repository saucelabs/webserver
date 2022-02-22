// Copyright 2021 The webserver Authors. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package telemetry

import (
	"sync"

	"go.opentelemetry.io/otel"
	stdout "go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

//////
// Const, and vars.
//////

const globalTracerName = "global"

//////
// Helpers.
//////

// Initializes the built-in tracer.
func initializeBuiltInTracer() (*sdktrace.TracerProvider, error) {
	exporter, err := stdout.New(stdout.WithPrettyPrint())
	if err != nil {
		return nil, err
	}

	builtInTracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithBatcher(exporter),
	)

	return builtInTracerProvider, nil
}

//////
// Interface.
//////

// ITelemetry defines what a Telemetry does.
type ITelemetry interface {
	// GetGlobalTracer returns the global tracer.
	GetGlobalTracer() trace.Tracer

	// GetTracer retrieves a tracer. If the retrieved tracer doesn't exist, the
	// global tracer is returned.
	GetTracer(name string) trace.Tracer

	// NewTracer creates a tracer from the current provider.
	NewTracer(name string) trace.Tracer
}

//////
// Definition.
//////

// Telemetry definition.
type Telemetry struct {
	// Provider accesses/consumes instrumentation.
	//
	// SEE: https://opentelemetry.io/docs/instrumentation/go/exporting_data/
	Provider trace.TracerProvider

	// TextMapPropagator propagates cross-cutting concerns as key-value text.
	//
	// SEE: // SEE: https://opentelemetry.io/docs/instrumentation/go/manual/#propagators-and-context
	TextMapPropagator []propagation.TextMapPropagator

	// Contains a map of tracers. By default, a global tracer is provided.
	// A tracer creates Spans.
	tracers sync.Map
}

//////
// ITelemetry implementation.
//////

// NewTracer creates a tracer from the current provider.
func (t *Telemetry) NewTracer(name string) trace.Tracer {
	tracer := t.Provider.Tracer(name)

	t.tracers.Store(name, tracer)

	return tracer
}

// GetTracer retrieves a tracer. If the retrieved tracer doesn't exist, the
// global tracer is returned.
func (t *Telemetry) GetTracer(name string) trace.Tracer {
	if tracer, ok := t.tracers.Load(name); ok {
		return tracer.(trace.Tracer)
	}

	return t.GetGlobalTracer()
}

// GetGlobalTracer returns the global tracer.
func (t *Telemetry) GetGlobalTracer() trace.Tracer {
	return t.GetTracer(globalTracerName)
}

//////
// Factory.
//////

// New is Telemetry factory.
func New(
	name string,
	provider trace.TracerProvider,
	textMapPropagators ...propagation.TextMapPropagator,
) (*Telemetry, error) {
	telemetry := &Telemetry{
		Provider:          provider,
		TextMapPropagator: textMapPropagators,
	}

	telemetry.tracers.Store(globalTracerName, otel.Tracer(name))

	otel.SetTracerProvider(telemetry.Provider)

	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(telemetry.TextMapPropagator...))

	return telemetry, nil
}

// NewDefault returns a telemetry with the default tracer, the built-in from the
// SDK which exports to `stdout`, and samples every trace.
func NewDefault(name string) (*Telemetry, error) {
	builtInTracerProvider, err := initializeBuiltInTracer()
	if err != nil {
		return nil, err
	}

	return New(
		name,
		builtInTracerProvider,
		propagation.TraceContext{}, propagation.Baggage{},
	)
}
