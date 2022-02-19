# webserver

`webserver` provides a server in which:

- Gracefully handles shutdown
- As built-in logger powered Sypl
- As telemetry powered by Open Telemetry
- As built-in useful handlers such as liveness, and readiness
- As metrics powered by ExpVar
- Applies best practices such as setting up timeouts.

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
