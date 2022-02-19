// Copyright 2021 The webserver Authors. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package handler

import "net/http"

// Handler definition.
type Handler struct {
	// Handler function.
	Handler http.HandlerFunc

	// Method to run the `Handler`.
	Method string

	// Path to run the `Handler`.
	Path string
}

// New is `Handler` factory.
func New(method string, path string, handler http.HandlerFunc) Handler {
	return Handler{
		Handler: handler,
		Method:  method,
		Path:    path,
	}
}
