# Changelog

All notable changes to this project will be documented in this file.

Please choose versions by [Semantic Versioning](http://semver.org/).

* MAJOR version when you make incompatible API changes,
* MINOR version when you add functionality in a backwards-compatible manner, and
* PATCH version when you make backwards-compatible bug fixes.

## v1.3.1

- refactor MatchesFilter function to separate filter.go file with dedicated tests
- add pkg factory pattern for clean dependency injection and handler creation
- improve code organization by separating filtering logic from changes provider
- add comprehensive unit tests for factory pattern
- update main.go to use factory pattern for handler creation

## v1.3.0

- add binary filtering functionality to `/read` endpoint
- add optional `filter` query parameter for exact binary pattern matching
- filter performs case-sensitive substring search in raw Kafka message values
- split `pkg/changes.go` into separate files by type (`record.go` and `changes-provider.go`)
- add comprehensive tests for binary filtering functionality
- improve test coverage to 52.2%

## v1.2.0

- add tests for all important files

## v1.1.0

- remove vendor
- add github workflow
- go mod update

## v1.0.1

- improve unparseable json output

## v1.0.0

- Initial Version
