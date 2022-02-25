// Copyright 2021 The webserver Authors. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package handler

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
)

// ReadinessDeterminer definition. It determines if `name` is ready.
type ReadinessDeterminer struct {
	name  string
	ready bool
	m     sync.Mutex
}

// Set state name.
func (t *ReadinessDeterminer) SetName(name string) {
	t.m.Lock()
	defer t.m.Unlock()

	t.name = name
}

// Get state name.
func (t *ReadinessDeterminer) GetName() string {
	t.m.Lock()
	defer t.m.Unlock()

	return t.name
}

// Set readiness state.
func (t *ReadinessDeterminer) SetReadiness(v bool) {
	t.m.Lock()
	defer t.m.Unlock()

	t.ready = v
}

// Get readiness state.
func (t *ReadinessDeterminer) GetReadiness() bool {
	t.m.Lock()
	defer t.m.Unlock()

	return t.ready
}

// NewReadinessDeterminer is the Readiness factory.
func NewReadinessDeterminer(name string) *ReadinessDeterminer {
	return &ReadinessDeterminer{
		name:  name,
		ready: false,
		m:     sync.Mutex{},
	}
}

// Readiness indicates the server is up, running, and ready to work. It follows
// the "standard" which is send `200` status code, and "OK" in the body if it's
// ready, otherwise sends `503`, "Service Unavailable", and the error. Multiple
// readinesses determiners can be passed. In this case, only if ALL are ready,
// the server will be considered ready.
func Readiness(readinessStates ...*ReadinessDeterminer) Handler {
	return Handler{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			readinessStateFinalState := true
			readinessesNames := []string{}

			for _, readinessState := range readinessStates {
				// If any state isn't ready, server isn't ready.
				if !readinessState.GetReadiness() {
					readinessStateFinalState = false
					readinessesNames = append(readinessesNames, readinessState.GetName())
				}
			}

			if !readinessStateFinalState {
				http.Error(
					w,
					fmt.Sprintf("server isn't ready. %s failed readiness", strings.Join(readinessesNames, ", ")),
					http.StatusServiceUnavailable,
				)

				return
			}

			w.Header().Set("Content-Type", "text/plain; charset=utf-8")

			w.WriteHeader(http.StatusOK)

			fmt.Fprintln(w, http.StatusText(http.StatusOK))
		}),
		Method: http.MethodGet,
		Path:   "/readiness",
	}
}
