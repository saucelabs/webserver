// Package webserver provides a HTTP server in which:
// - Gracefully handles shutdown
// - As built-in logger powered Sypl
// - As telemetry powered by Open Telemetry
// - As built-in useful handlers such as liveness, and readiness
// - As metrics powered by ExpVar
// - Applies best practices such as setting up timeouts.
package webserver
