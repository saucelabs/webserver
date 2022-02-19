// Copyright 2021 The webserver Authors. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package handler

import (
	"net/http"
)

// ReadinessFunc determines readiness. If error isn't `nil`, it means server
// isn't ready.
//
// NOTE: Be mindful that a readiness probe performs this every N-{s|ms}.
type ReadinessFunc func() error

// Readiness indicates the server is up, running, and ready to work. It follows
// the "standard" which is send "200", and "OK" if ready, or "503" and
// "Service Unavailable" plus the error, if not ready.
func Readiness(readinessFunc ReadinessFunc) Handler {
	return Handler{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if err := readinessFunc(); err != nil {
				http.Error(w, err.Error(), http.StatusServiceUnavailable)

				return
			}

			w.Header().Set("Content-Type", "text/plain; charset=utf-8")

			w.WriteHeader(http.StatusOK)

			w.Write([]byte(http.StatusText(http.StatusOK)))
		}),
		Method: http.MethodGet,
		Path:   "/readiness",
	}
}
