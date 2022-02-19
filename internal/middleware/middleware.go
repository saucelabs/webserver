package middleware

import (
	"net/http"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/saucelabs/sypl"
)

// Log requests in the Apache Combined Log Format.
func Logger(l sypl.ISypl) mux.MiddlewareFunc {
	return func(h http.Handler) http.Handler {
		return handlers.CombinedLoggingHandler(l, h)
	}
}
