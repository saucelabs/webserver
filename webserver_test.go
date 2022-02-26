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
	"syscall"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/saucelabs/randomness"
	"github.com/saucelabs/webserver/handler"
	"github.com/saucelabs/webserver/metric"
)

const serverName = "test-server"

// Client simulation.
var c = http.Client{Timeout: time.Duration(10) * time.Second}

// Generates random ports.
func generatePort(t *testing.T) int64 {
	t.Helper()

	// Random port.
	r, err := randomness.New(3000, 7000, 10, true)
	if err != nil {
		t.Fatal(err)
	}

	return r.MustGenerate()
}

// Setup a test server.
func setupTestServer(t *testing.T) (IServer, int) {
	t.Helper()

	port := generatePort(t)

	// A classic metric counter.
	counterMetric := metric.NewInt("simple_metric_example_counter")
	counterMetric.Add(1)

	// Router.
	myCustomRouter := mux.NewRouter()
	apiRouter := myCustomRouter.PathPrefix("/api").Subrouter()
	versionedRouter := apiRouter.PathPrefix("/v1").Subrouter()

	// Test server setting many options...
	testServer, err := New(serverName, fmt.Sprintf("0.0.0.0:%d", port),
		WithRouter(versionedRouter),
		// Add a custom handler to the list of pre-loaded handlers.
		WithHandlers(
			handler.Liveness(),
			handler.OK(),
			handler.Stop(),
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
		WithMetrics(metric.Metric{
			Name: "metric_1",
			Value: metric.Func(func() interface{} {
				return struct {
					CustomValue string `json:"custom_value"`
				}{
					CustomValue: "any_value",
				}
			}),
		}),
		WithTimeout(3*time.Second, 1*time.Second, 3*time.Second, 10*time.Second, 3*time.Second),
	)
	if err != nil {
		log.Fatalf("Failed to setup %s, %v", serverName, err)
	}

	// This is how a developer, importing this package would add routers, and
	// routes.
	sr := testServer.GetRouter().PathPrefix("/router2").Subrouter()

	sr.HandleFunc("/counter", func(rw http.ResponseWriter, r *http.Request) {
		rw.WriteHeader(http.StatusOK)

		fmt.Fprintln(rw, http.StatusText(http.StatusOK))

		// Increase metric counter example.
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
				url:                  "/api/v1/liveness",
				sc:                   http.StatusOK,
				expectedBodyContains: http.StatusText(http.StatusOK),
			},
		},
		{
			name: "Should work - /",
			args: args{
				port:                 port,
				url:                  "/api/v1/",
				sc:                   http.StatusOK,
				expectedBodyContains: http.StatusText(http.StatusOK),
			},
		},
		{
			name: "Should work - /ok",
			args: args{
				port:                 port,
				url:                  "/api/v1/ok",
				sc:                   http.StatusOK,
				expectedBodyContains: http.StatusText(http.StatusOK),
			},
		},
		{
			name: "Should work - sub-router - /router2/counter",
			args: args{
				port:                 port,
				url:                  "/api/v1/router2/counter",
				sc:                   http.StatusOK,
				expectedBodyContains: http.StatusText(http.StatusOK),
			},
		},
		{
			name: "Should work - /debug/vars - counter",
			args: args{
				port:                 port,
				url:                  "/api/v1/debug/vars",
				sc:                   http.StatusOK,
				expectedBodyContains: `"simple_metric_example_counter": 2`,
			},
		},
		{
			name: "Should work - /slow",
			args: args{
				port:                 port,
				url:                  "/api/v1/slow",
				sc:                   http.StatusServiceUnavailable,
				expectedBodyContains: ErrRequesTimeout.Error(),
				delay:                3 * time.Second,
			},
		},
		{
			name: "Should work - /stop",
			args: args{
				port:                 port,
				url:                  "/api/v1/stop",
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

func TestNew_address(t *testing.T) {
	type args struct {
		host string
	}
	tests := []struct {
		name           string
		args           args
		needRandomPort bool
		wantErr        bool
	}{
		{
			name: "Should fail - empty",
			args: args{
				host: "",
			},
			wantErr: true,
		},
		{
			name: "Should fail - localhost -> missing port",
			args: args{
				host: "localhost",
			},
			wantErr: true,
		},
		{
			name: "Should work - :NNNN -> localhost:NNNN",
			args: args{
				host: "",
			},
			needRandomPort: true,
			wantErr:        false,
		},
		{
			name: "Should work - localhost:NNNN",
			args: args{
				host: "localhost",
			},
			needRandomPort: true,
			wantErr:        false,
		},
		{
			name: "Should work - 0.0.0.0:NNNN",
			args: args{
				host: "0.0.0.0",
			},
			needRandomPort: true,
			wantErr:        false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Host, and port builder.
			port := ""

			if tt.needRandomPort {
				port = fmt.Sprintf("%d", generatePort(t))
			}

			address := fmt.Sprintf("%s:%s", tt.args.host, port)

			if address == ":" {
				address = ""
			}

			if address == "localhost:" {
				address = "localhost"
			}

			testServer, err := New(serverName, address)
			if err != nil && !tt.wantErr {
				t.Error(err)

				return
			}

			if testServer != nil {
				go func() {
					defer func() {
						if err := testServer.Stop(syscall.SIGKILL); err != nil {
							log.Fatal(err)

							return
						}
					}()
					if err := testServer.Start(); err != nil && !tt.wantErr {
						log.Fatal(err)

						return
					}
				}()
			}
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

	testServer, err := New(serverName, fmt.Sprintf("0.0.0.0:%d", port),
		WithHandlers(
			handler.Liveness(),
		),
		WithLogging("none", "none", ""),
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
