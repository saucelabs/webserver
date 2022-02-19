// Copyright 2021 The webserver Authors. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package handler

import (
	"fmt"
	"net/http"
)

// OK replies to the request with the "200", and "OK".
func OK() Handler {
	return Handler{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")

			w.WriteHeader(http.StatusOK)

			fmt.Fprintln(w, http.StatusText(http.StatusOK))
		}),
		Method: http.MethodGet,
		Path:   "/",
	}
}
