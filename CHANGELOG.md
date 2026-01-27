# Changelog

All notable changes to this project will be documented in this file.

Please choose versions by [Semantic Versioning](http://semver.org/).

* MAJOR version when you make incompatible API changes,
* MINOR version when you add functionality in a backwards-compatible manner, and
* PATCH version when you make backwards-compatible bug fixes.

## v1.6.4

- improve Docker build with BuildKit and build args for git version, commit, and date
- add BUILD_GIT_COMMIT and BUILD_DATE environment variables to Dockerfile
- refactor Makefile to use VERSION from git tags
- add check-go-mod target to automate vendor updates
- improve Docker cleanup with builder prune and max-used-space limit
- rename REGISTRY to DOCKER_REGISTRY for clarity

## v1.6.3

- refactor build metrics from global state to dependency injection
- add configurable Prometheus namespace via PROMETHEUS_NAMESPACE
- move build-info-metrics.go from metrics package to pkg package

## v1.6.2

- update golang to 1.25.6
- update alpine to 3.23
- update bborbe/* dependencies to latest
- update IBM/sarama to v1.46.3
- update getsentry/sentry-go to v0.40.0

## v1.6.1

- add .PHONY declarations to all Makefile targets
- add go-modtool for go.mod formatting in format target
- add mocks directory creation in generate target
- update dependencies to latest versions

## v1.6.0

- enhance golangci-lint configuration with comprehensive linter settings (funlen, gocognit, nestif, maintidx)
- add readability and style linters (gofmt, goimports, errname, unparam)
- add safety linters (bodyclose, forcetypeassert, asasalint, prealloc)
- refactor changes-provider.go to reduce cognitive complexity (32 → <20) with extracted helper methods
- refactor handler.go to reduce cognitive complexity (22 → <20) with extracted functions
- update containerd dependency to v1.7.29 (security fix for GHSA-m6hq-p25p-ffr2, GHSA-pwhc-rpq9-4c8w)
- update opencontainers/selinux to v1.13.0 (security fix for GHSA-cgrx-mc8f-2prm)
- improve code maintainability by extracting focused, single-responsibility functions
- update CI/CD workflow configuration
- update Dockerfile dependencies

## v1.5.0

- add configurable error preview content length via ERROR_PREVIEW_CONTENT_LENGTH environment variable
- add --error-preview-content-length flag to control error message preview size
- add support for unlimited preview length with -1 value
- add comprehensive tests for configurable preview length (10 bytes, unlimited, zero, larger than content)
- update converter to accept errorPreviewContentLength parameter
- update factory to pass preview length configuration through dependency chain
- maintain backward compatibility with default 100 bytes preview length

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
