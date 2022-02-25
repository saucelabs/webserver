# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## Roadmap

To be defined.

## [0.0.5] - 2022-02-25
### Added
- Option to set the base router.
- New Metric type.

### Changed
- Setting a feature's option enables the feature.
- `WithMetrics`, `WithHandlers` and `WithReadiness` now accepts multiple params.
- `New`'s params for handler, and metrics are validated.

## [0.0.4] - 2022-02-23
### Changed
- `ReadinessState` is now called `ReadinessDeterminer`.

## [0.0.3] - 2022-02-23
### Changed
- `ReadinessHandler` now provides a `ReadinessState` determiner.
- Readiness determination now support multiple `ReadinessState` determiners.

## [0.0.2] - 2022-02-22
### Added
- Added `GetTelemetry` to the `IServer` interface.
- Added more tests.

### Changed
- Fixed `NewBasic` options processing.

## [0.0.1] - 2022-02-22
### Added
- First release.
