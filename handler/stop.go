// Copyright 2021 The webserver Authors. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package handler

import (
	"fmt"
	"net/http"
	"syscall"
)

// Stop allows the server to be remotely, and gracefully stopped.
func Stop() Handler {
	return Handler{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")

			w.WriteHeader(http.StatusOK)

			fmt.Fprintln(w, http.StatusText(http.StatusOK))

			syscall.Kill(syscall.Getpid(), syscall.SIGINT)
		}),
		Method: http.MethodGet,
		Path:   "/stop",
	}
}
