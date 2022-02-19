// Copyright 2021 The webserver Authors. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package handler

import (
	"net/http"

	"github.com/saucelabs/webserver/internal/expvar"
)

// ExpVar serves metrics.
func ExpVar() Handler {
	return Handler{
		Handler: expvar.Handler().ServeHTTP,
		Method:  http.MethodGet,
		Path:    "/debug/vars",
	}
}
