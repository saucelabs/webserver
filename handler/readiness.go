// Copyright 2021 The webserver Authors. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package handler

import (
	"fmt"
	"net/http"
)

// ReadinessFunc determines readiness. If error isn't `nil`, it means server
// isn't ready.
//
// NOTE: Be mindful that a readiness probe performs this every N-{s|ms}.
type ReadinessFunc func() error

// Readiness indicates the server is up, running, and ready to work. It follows
// the "standard" which is send `200` status code, and "OK" in the body if it's
// ready, otherwise sends `503`, "Service Unavailable", and the error.
func Readiness(readinessFunc ReadinessFunc) Handler {
	return Handler{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if err := readinessFunc(); err != nil {
				http.Error(w, err.Error(), http.StatusServiceUnavailable)

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
