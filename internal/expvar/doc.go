// Package expvar is the standard Golang's expvar package with one modification:
// it removes the bad practice of using `init()`, instead exposes:
// - `Start`: does everything the old `init` does
// - `PublishCmdLine`: publishes command line information
// - `PublishMemStats`: publishes memory statistics
// - `RegisterHandler`: registers the standard endpoint: `GET /debug/vars`.
package expvar
