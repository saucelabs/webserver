// Copyright 2021 The webserver Authors. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package webserver_test

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/saucelabs/randomness"
	"github.com/saucelabs/webserver"
	"github.com/saucelabs/webserver/handler"
)

const serverName = "test-server"

// Logs, and exit.
//nolint:forbidigo
func logAndExit(msg string) {
	fmt.Println(msg)

	os.Exit(1)
}

// Call.
//nolint:noctx,unparam
func callAndExpect(port int, url string, sc int, expectedBodyContains string) (int, string) {
	c := http.Client{Timeout: time.Duration(10) * time.Second}

	resp, err := c.Get(fmt.Sprintf("http://0.0.0.0:%d/%s", port, url))
	if err != nil {
		logAndExit(err.Error())
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logAndExit(err.Error())
	}

	if sc != 0 {
		if resp.StatusCode != sc {
			logAndExit(fmt.Sprintf("Expect %v got %v\n", sc, resp.StatusCode))
		}
	}

	bodyS := string(body)

	if expectedBodyContains != "" {
		if !strings.Contains(bodyS, expectedBodyContains) {
			logAndExit(fmt.Sprintf("Expect %v got %v\n", expectedBodyContains, bodyS))
		}
	}

	return resp.StatusCode, bodyS
}

//////
// Examples.
//////

// Example showing how to use the Web Server.
func ExampleNew() {
	// Probe results accumulator. Golang's example requires some output to be
	// tested against. This accumulator will serve it.
	probeResults := []string{}
	probeResultsLocker := sync.Mutex{}

	// Part of the readiness simulation.
	readinessFlag := handler.NewReadinessDeterminer("tunnel")

	// Golang's example are like tests, it's a bad practice to have a hardcoded
	// port because of the possibility of collision. Generate a random port.
	r, err := randomness.New(3000, 7000, 10, true)
	if err != nil {
		logAndExit(err.Error())
	}

	port := r.MustGenerate()

	// Setup server settings some options.
	testServer, err := webserver.New(serverName, fmt.Sprintf("0.0.0.0:%d", port),
		// Sets server readiness.
		webserver.WithReadiness(readinessFlag),
	)
	if err != nil {
		logAndExit(err.Error())
	}

	// Start server, non-blocking way.
	go func() {
		if err := testServer.Start(); err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				fmt.Println("server stopped")
			} else {
				logAndExit(err.Error())
			}
		}
	}()

	// Ensures enough time for the server to be up, and ready - just for testing.
	time.Sleep(3 * time.Second)

	// Simulates a Readiness probe, for example, Kubernetes.
	go func() {
		for {
			_, body := callAndExpect(int(port), "/readiness", 0, "")

			probeResultsLocker.Lock()
			probeResults = append(probeResults, body)
			probeResultsLocker.Unlock()

			// Probe wait time.
			time.Sleep(500 * time.Millisecond)
		}
	}()

	// Simulates some action which indicates server is ready, example data was
	// loaded from DB, or got updated data from another service.
	go func() {
		time.Sleep(2 * time.Second)

		readinessFlag.SetReadiness(true)
	}()

	// Hold the server online for testing.
	time.Sleep(5 * time.Second)

	// Satisfies Golang example output need.
	probeResultsLocker.Lock()
	fmt.Println(strings.Contains(strings.Join(probeResults, ","), "OK"))
	probeResultsLocker.Unlock()

	// output:
	// true
}

func ExampleNewDefault() {
	// Golang's example are like tests, it's a bad practice to have a hardcoded
	// port because of the possibility of collision. Generate a random port.
	r, err := randomness.New(3000, 7000, 10, true)
	if err != nil {
		logAndExit(err.Error())
	}

	port := r.MustGenerate()

	// Setup server settings some options.
	testServer, err := webserver.NewDefault(serverName, fmt.Sprintf("0.0.0.0:%d", port))
	if err != nil {
		logAndExit(err.Error())
	}

	// Start server, non-blocking way.
	go func() {
		if err := testServer.Start(); err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				fmt.Println("server stopped")
			} else {
				logAndExit(err.Error())
			}
		}
	}()

	// Ensures enough time for the server to be up, and ready - just for testing.
	time.Sleep(3 * time.Second)

	_, body := callAndExpect(int(port), "/api/v1/", 200, "OK")

	fmt.Println(body)

	// output:
	// OK
}
