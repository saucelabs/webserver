# webserver

`webserver` provides a web server:

- Gracefully handles shutdown
- Applies best practices such as setting up timeouts
- Routing powered by Gorilla Mux
- Logging powered Sypl
- HTTP server powered by Go built-in HTTP server
- Observability is first-class:
  - Telemetry powered by Open Telemetry
  - Metrics powered by ExpVar
  - Built-in useful handlers such as liveness, and readiness

## Install

`$ go get github.com/saucelabs/webserver`

### Specific version

Example: `$ go get github.com/saucelabs/webserver@v1.2.3`

## Usage

See [`example_test.go`](example_test.go), and [`webserver_test.go`](webserver_test.go) file.

### Documentation

Run `$ make doc` or check out [online](https://pkg.go.dev/github.com/saucelabs/webserver).

## Development

Check out [CONTRIBUTION](CONTRIBUTION.md).

### Release

1. Update [CHANGELOG](CHANGELOG.md) accordingly.
2. Once changes from MR are merged.
3. Tag and release.

## Roadmap

Check out [CHANGELOG](CHANGELOG.md).
