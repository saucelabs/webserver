// Copyright 2021 The webserver Authors. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package handler

import (
	"fmt"
	"net/http"
	"syscall"
)

// Stop allows the server to be remotely, and gracefully stopped. Optionally set
// the `hard` query param to `true` to immediately kill the server.
func Stop() Handler {
	return Handler{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			queryParams := r.URL.Query()

			w.Header().Set("Content-Type", "text/plain; charset=utf-8")

			w.WriteHeader(http.StatusOK)

			fmt.Fprintln(w, http.StatusText(http.StatusOK))

			if queryParams.Get("hard") == "true" {
				if err := syscall.Kill(syscall.Getpid(), syscall.SIGKILL); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
				}
			} else {
				if err := syscall.Kill(syscall.Getpid(), syscall.SIGINT); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
				}
			}
		}),
		Method: http.MethodGet,
		Path:   "/stop",
	}
}
