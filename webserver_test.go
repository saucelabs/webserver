// Copyright 2021 The webserver Authors. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package webserver

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/saucelabs/randomness"
	"github.com/saucelabs/webserver/handler"
	"github.com/saucelabs/webserver/internal/expvar"
)

const serverName = "test-server"

// Client simulation.
var c = http.Client{Timeout: time.Duration(10) * time.Second}

// Setup a test server.
func setupTestServer(t *testing.T) (IServer, int) {
	t.Helper()

	// Random port.
	r, err := randomness.New(3000, 7000, 10, true)
	if err != nil {
		t.Fatal(err)
	}

	port := r.MustGenerate()

	// A classic ExpVar counter.
	counterMetric := expvar.NewInt("simple_metric_example_counter")
	counterMetric.Add(1)

	// Test server setting many options...
	testServer, err := New(serverName, fmt.Sprintf("0.0.0.0:%d", port),
		// Add a custom handler to the list of pre-loaded handlers.
		WithPreLoadedHandlers(
			// Simulates a slow operation which should timeout.
			handler.Handler{
				Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					time.Sleep(3 * time.Second)

					w.Header().Set("Content-Type", "text/plain; charset=utf-8")

					w.WriteHeader(http.StatusOK)

					fmt.Fprintln(w, http.StatusText(http.StatusOK))
				}),
				Method: http.MethodGet,
				Path:   "/slow",
			},
			// A `200` handler.
			handler.Handler{
				Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "text/plain; charset=utf-8")

					w.WriteHeader(http.StatusOK)

					fmt.Fprintln(w, http.StatusText(http.StatusOK))
				}),
				Method: http.MethodGet,
				Path:   "/ok",
			},
		),
		// Setting metrics using both the quick, and "raw" way.
		WithMetrics("simple_metric_example_string", "any_value"),
		WithMetrics("simple_metric_example_int", 1),
		WithMetrics("simple_metric_example_bool", true),
		WithMetrics("simple_metric_example_slice", []string{"any_value"}),
		WithMetrics("simple_metric_example_struct", struct {
			CustomValue string `json:"custom_value"`
		}{
			CustomValue: "any_value",
		},
		),
		WithMetricsRaw("raw_metrics_example", expvar.Func(func() interface{} {
			return struct {
				CustomValue string `json:"custom_value"`
			}{
				CustomValue: "any_value",
			}
		})),
		WithTimeout(3*time.Second, 1*time.Second, 3*time.Second, 10*time.Second, 3*time.Second),
		WithoutTelemetry(),
	)
	if err != nil {
		log.Fatalf("Failed to setup %s, %v", serverName, err)
	}

	// This is how a developer, importing this package would add routers, and
	// routes.
	sr := testServer.GetRouter().PathPrefix("/api").Subrouter()

	sr.HandleFunc("/counter", func(rw http.ResponseWriter, r *http.Request) {
		rw.WriteHeader(http.StatusOK)

		fmt.Fprintln(rw, http.StatusText(http.StatusOK))

		// Increase ExpVar counter example.
		counterMetric.Add(1)
	})

	return testServer, int(port)
}

// DRY on calling an endpoint, and checking expectations.
//nolint:noctx,unparam
func callAndExpect(t *testing.T, port int, url string, sc int, expectedBodyContains string) (int, string) {
	t.Helper()

	resp, err := c.Get(fmt.Sprintf("http://0.0.0.0:%d/%s", port, url))
	if err != nil {
		t.Fatal(err)
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	if sc != 0 {
		if resp.StatusCode != sc {
			t.Fatalf("Expect %v got %v", sc, resp.StatusCode)
		}
	}

	var finalBody string

	if body != nil {
		finalBody = string(body)

		if expectedBodyContains != "" {
			if !strings.Contains(finalBody, expectedBodyContains) {
				t.Fatalf("Expect %v got %v", expectedBodyContains, finalBody)
			}
		}
	}

	return resp.StatusCode, finalBody
}

func TestNew(t *testing.T) {
	// Test server.
	testServer, port := setupTestServer(t)

	// Starts in a non-blocking way.
	go func() {
		if err := testServer.Start(); err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				testServer.GetLogger().Infoln("server stopped")
			} else {
				log.Fatal(err)
			}
		}
	}()

	// Ensures enough time for the server to be up, and ready - just for testing.
	time.Sleep(3 * time.Second)

	type args struct {
		port                 int
		url                  string
		sc                   int
		expectedBodyContains string
		delay                time.Duration
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "Should work - liveness",
			args: args{
				port:                 port,
				url:                  "/liveness",
				sc:                   http.StatusOK,
				expectedBodyContains: http.StatusText(http.StatusOK),
			},
		},
		{
			name: "Should work - /",
			args: args{
				port:                 port,
				url:                  "/",
				sc:                   http.StatusOK,
				expectedBodyContains: http.StatusText(http.StatusOK),
			},
		},
		{
			name: "Should work - /ok",
			args: args{
				port:                 port,
				url:                  "/ok",
				sc:                   http.StatusOK,
				expectedBodyContains: http.StatusText(http.StatusOK),
			},
		},
		{
			name: "Should work - sub-router - /api/counter",
			args: args{
				port:                 port,
				url:                  "/api/counter",
				sc:                   http.StatusOK,
				expectedBodyContains: http.StatusText(http.StatusOK),
			},
		},
		{
			name: "Should work - /debug/vars - counter",
			args: args{
				port:                 port,
				url:                  "/debug/vars",
				sc:                   http.StatusOK,
				expectedBodyContains: `"simple_metric_example_counter": 2`,
			},
		},
		{
			name: "Should work - /slow",
			args: args{
				port:                 port,
				url:                  "/slow",
				sc:                   http.StatusServiceUnavailable,
				expectedBodyContains: ErrRequesTimeout.Error(),
				delay:                3 * time.Second,
			},
		},
		{
			name: "Should work - /stop",
			args: args{
				port:                 port,
				url:                  "/stop",
				sc:                   http.StatusOK,
				expectedBodyContains: http.StatusText(http.StatusOK),
				delay:                3 * time.Second,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			callAndExpect(t, tt.args.port, tt.args.url, tt.args.sc, tt.args.expectedBodyContains)
		})
	}
}

func TestNewBasic(t *testing.T) {
	// Random port.
	r, err := randomness.New(3000, 7000, 10, true)
	if err != nil {
		t.Fatal(err)
	}

	port := r.MustGenerate()

	testServer, err := NewBasic(serverName, fmt.Sprintf("0.0.0.0:%d", port),
		WithPreLoadedHandlers(
			handler.Liveness(),
		),
		WithLoggingOptions("none", "none", ""),
	)
	if err != nil {
		log.Fatalf("Failed to setup %s, %v", serverName, err)
	}

	if testServer.GetTelemetry() != nil {
		t.Fatal("Expected no telemetry")
	}

	// Starts in a non-blocking way.
	go func() {
		if err := testServer.Start(); err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				testServer.GetLogger().Infoln("server stopped")
			} else {
				log.Fatal(err)
			}
		}
	}()

	// Ensures enough time for the server to be up, and ready - just for testing.
	time.Sleep(3 * time.Second)

	type args struct {
		port                 int64
		url                  string
		sc                   int
		expectedBodyContains string
	}
	tests := []struct {
		name    string
		args    args
		want    IServer
		wantErr bool
	}{
		{
			name: "Should work - liveness",
			args: args{
				port:                 port,
				url:                  "/liveness",
				sc:                   http.StatusOK,
				expectedBodyContains: http.StatusText(http.StatusOK),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			callAndExpect(t, int(tt.args.port), tt.args.url, tt.args.sc, tt.args.expectedBodyContains)
		})
	}
}
