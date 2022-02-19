// Package webserver provides a HTTP server:
// - Gracefully handles shutdown
// - Applies best practices such as setting up timeouts
// - Routing powered by Gorilla Mux
// - HTTP server powered by Go built-in HTTP server
// - Observability is first-class:
//   - Logging powered Sypl
//   - Telemetry powered by Open Telemetry
//   - Metrics powered by ExpVar
//   - Built-in useful handlers such as liveness, and readiness.
package webserver
