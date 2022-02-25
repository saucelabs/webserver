// Copyright 2021 The webserver Authors. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package validation

import (
	"github.com/go-playground/validator/v10"
	"github.com/saucelabs/customerror"
)

// Re-usable, cached validator.
// SEE: https://github.com/go-playground/validator/blob/master/_examples/simple/main.go#L27
var validatorSingleton *validator.Validate

// Setup validator.
func setup() *validator.Validate {
	validatorSingleton = validator.New()

	return validatorSingleton
}

// Get safely returns the application validator.
func Get() *validator.Validate {
	if validatorSingleton == nil {
		return setup()
	}

	return validatorSingleton
}

// ValidateStruct allows DRY around the repetitive work of validating structs.
func ValidateStruct(f interface{}) error {
	if err := Get().Struct(f); err != nil {
		return customerror.NewInvalidError("data", customerror.WithError(err))
	}

	return nil
}
