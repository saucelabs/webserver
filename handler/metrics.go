// Copyright 2021 The webserver Authors. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package handler

import (
	"net/http"

	"github.com/saucelabs/webserver/metric"
)

// Metrics serves metrics.
func Metrics() Handler {
	return Handler{
		Handler: metric.Handler().ServeHTTP,
		Method:  http.MethodGet,
		Path:    "/debug/vars",
	}
}
