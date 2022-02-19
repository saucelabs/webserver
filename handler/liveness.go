// Copyright 2021 The webserver Authors. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package handler

import (
	"fmt"
	"net/http"
)

// Liveness indicates the server is up, and running. It follows the "standard"
// which is send `200` status code, and "OK" in the body.
func Liveness() Handler {
	return Handler{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")

			w.WriteHeader(http.StatusOK)

			fmt.Fprintln(w, http.StatusText(http.StatusOK))
		}),
		Method: http.MethodGet,
		Path:   "/liveness",
	}
}
