# Changelog

All notable changes to this project will be documented in this file.

Please choose versions by [Semantic Versioning](http://semver.org/).

* MAJOR version when you make incompatible API changes,
* MINOR version when you add functionality in a backwards-compatible manner, and
* PATCH version when you make backwards-compatible bug fixes.

## v1.4.1

- improve JSON unmarshal error messages with structured format
- add security fix: limit error preview fields to 100 bytes to prevent DoS
- change error format from string to map with error, valueLength, previewBase64, previewHex
- add hex encoding for binary data preview (safe for non-UTF8 data)
- add comprehensive edge case tests (100 bytes boundary, truncation, binary data)
- add function documentation explaining error structure
- migrate to libhttp.SendJSONResponse with context support
- update dependencies (bborbe/http v1.16.0)
- improve code formatting and organization

## v1.4.0

- add golangci-lint configuration with comprehensive linter settings
- enhance Makefile with additional security checks (gosec, trivy, osv-scanner)
- add code formatting tools (golines for line length enforcement, gofmt)
- add license header automation to precommit workflow
- update build tooling dependencies in tools.go
- improve CI/CD workflow configuration
- refactor code organization and update dependencies
- disable race detection in tests for performance
- enhance error handling and code quality checks

## v1.3.2

- add input validation for filter parameter with maximum length check (1024 bytes)
- add comprehensive tests for filter parameter validation
- update API documentation to include filter parameter limits
- improve security by preventing potential DoS attacks via oversized filter parameters

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
