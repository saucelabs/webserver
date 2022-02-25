// Copyright 2021 The webserver Authors. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package webserver

import (
	"context"
	"errors"
	"os"

	"github.com/gorilla/mux"
	handler "github.com/saucelabs/webserver/handler"
)

// Adds a `Handler` to a `Router`.
func addHandler(router *mux.Router, handlers ...handler.Handler) {
	for _, handler := range handlers {
		router.HandleFunc(handler.Path, handler.Handler).Methods(handler.Method)
	}
}

// Verifies is `err` is a timeout.
func isTimeoutError(err error) bool {
	if errors.Is(err, context.DeadlineExceeded) ||
		errors.Is(err, os.ErrDeadlineExceeded) ||
		os.IsTimeout(err) {
		return true
	}

	return false
}
