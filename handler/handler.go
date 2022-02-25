// Copyright 2021 The webserver Authors. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package handler

import (
	"net/http"

	"github.com/saucelabs/webserver/internal/validation"
)

//////
// Definition.
//////

// Handler definition.
type Handler struct {
	// Handler function.
	Handler http.HandlerFunc `json:"handler" validate:"required"`

	// Method to run the `Handler`.
	Method string `json:"method" validate:"required"`

	// Path to run the `Handler`.
	Path string `json:"path" validate:"required"`
}

//////
// Factory.
//////

// New is `Handler` factory.
func New(method string, path string, handler http.HandlerFunc) (Handler, error) {
	h := Handler{}

	h.Handler = handler
	h.Method = method
	h.Path = path

	if err := validation.ValidateStruct(h); err != nil {
		return Handler{}, err
	}

	return h, nil
}
