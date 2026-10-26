DOCKER_REGISTRY ?= docker.io
IMAGE ?= bborbe/kafka-topic-reader
BRANCH ?= $(shell git rev-parse --abbrev-ref HEAD)
DIRS += $(shell find */* -maxdepth 0 -name Makefile -exec dirname "{}" \;)
ifeq ($(VERSION),)
	VERSION := $(shell git describe --tags `git rev-list --tags --max-count=1`)
endif

include tools.env

.PHONY: default
default: precommit

# The docker targets live in their own file so the BUILD_TARGET switch sits next to
# the commands it guards. Included after `default`, not before it: make takes its
# default goal from the first target it reads, so an include above this rule would
# silently make `build` what a bare `make` runs — and a bare `make` here must stay
# `precommit`. The include reads DOCKER_REGISTRY, IMAGE, VERSION and DIRS, all
# defined above it.
include Makefile.docker

.PHONY: precommit
precommit: ensure format generate test check addlicense
	@echo "ready to commit"

.PHONY: ensure
ensure:
	go mod tidy -e
	go mod verify
	rm -rf vendor

.PHONY: format
format:
	find . -type f -name 'go.mod' -not -path './vendor/*' -exec go run github.com/shoenig/go-modtool@$(GO_MODTOOL_VERSION) -w fmt "{}" \;
	go run github.com/incu6us/goimports-reviser/v3@$(GOIMPORTS_REVISER_VERSION) -project-name github.com/bborbe/kafka-topic-reader -format -excludes vendor ./...
	find . -type d -name vendor -prune -o -type f -name '*.go' -print0 | xargs -0 -n 10 go run github.com/segmentio/golines@$(GOLINES_VERSION) --max-len=100 -w
	# golines last, then gofmt last so its wrapping is normalized and the gofmt lint check passes
	find . -type f -name '*.go' -not -path './vendor/*' -exec gofmt -w "{}" +

.PHONY: generate
generate:
	rm -rf mocks avro
	mkdir -p mocks
	echo "package mocks" > mocks/mocks.go
	go generate -mod=mod ./...

.PHONY: test
test:
	# -race
	go test -mod=mod -p=$${GO_TEST_PARALLEL:-1} -cover $(shell go list -mod=mod ./... | grep -v /vendor/)

.PHONY: check
check: lint vet vulncheck osv-scanner trivy

.PHONY: lint
lint:
	go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION) run --allow-parallel-runners --config .golangci.yml ./...

.PHONY: vet
vet:
	go vet -mod=mod $(shell go list -mod=mod ./... | grep -v /vendor/)


VULNCHECK_IGNORE ?= GO-2026-4923 GO-2026-4514 GO-2022-0470 GO-2026-4772 GO-2026-4771 GO-2026-5932 GO-2026-6599 GO-2026-6600 GO-2026-6603 GO-2026-6604 GO-2026-6605 GO-2026-6607 GO-2026-6608 GO-2026-6609 GO-2026-6610 GO-2026-6611 GO-2026-6612 GO-2026-6613 GO-2026-6617

.PHONY: vulncheck
vulncheck:
	@PKGS="$(shell go list -mod=mod ./... | grep -v /vendor/)"; \
	IGNORE_JSON=$$(printf '%s\n' $(VULNCHECK_IGNORE) | jq -R . | jq -s .); \
	REMAIN=$$(go run golang.org/x/vuln/cmd/govulncheck@$(GOVULNCHECK_VERSION) -format json $$PKGS 2>/dev/null | \
		jq -rs --argjson ignore "$$IGNORE_JSON" \
			'(map(select(.osv != null)) | map({key: .osv.id, value: (.osv.summary // "")}) | from_entries) as $$sum | \
			 map(select(.finding != null) | .finding) | \
			 map(select(.osv as $$o | $$ignore | index($$o) | not)) | \
			 map("\(.osv)\t\(.trace[-1].module)@\(.trace[-1].version) -> \(.fixed_version)\t\($$sum[.osv] // "")") | \
			 unique | .[]'); \
	if [ -n "$$REMAIN" ]; then \
		echo "Unexpected vulnerabilities (ignored: $(VULNCHECK_IGNORE)):"; \
		printf '%s\n' "$$REMAIN" | column -t -s "$$(printf '\t')"; \
		exit 1; \
	else \
		echo "No unignored vulnerabilities found"; \
	fi

.PHONY: osv-scanner
osv-scanner:
	@if [ -f .osv-scanner.toml ]; then \
		echo "Using .osv-scanner.toml"; \
		go run github.com/google/osv-scanner/v2/cmd/osv-scanner@$(OSV_SCANNER_VERSION) --config .osv-scanner.toml --recursive .; \
	else \
		echo "No config found, running default scan"; \
		go run github.com/google/osv-scanner/v2/cmd/osv-scanner@$(OSV_SCANNER_VERSION) --recursive .; \
	fi

.PHONY: trivy
trivy:
	trivy fs \
	--db-repository ghcr.io/aquasecurity/trivy-db \
	--scanners vuln,secret \
	--quiet \
	--no-progress \
	--disable-telemetry \
	--exit-code 1 .

.PHONY: addlicense
addlicense:
	go run github.com/google/addlicense@$(ADDLICENSE_VERSION) -c "Benjamin Borbe" -y $$(date +'%Y') -l bsd $$(find . -name "*.go" -not -path './vendor/*')

.PHONY: check-version-tag
check-version-tag:
	@if [ -n "$(ALLOW_UNTAGGED_BUILD)" ]; then \
		echo "ALLOW_UNTAGGED_BUILD set — skipping version/tag check"; \
	else \
		head_tag=$$(git describe --tags --exact-match HEAD 2>/dev/null); \
		if [ "$$head_tag" != "$(BUILD_TAG)" ]; then \
			echo "ERROR: refusing to build $(BUILD_TAG) from this tree." >&2; \
			echo "  HEAD is at tag: $${head_tag:-<untagged>}" >&2; \
			echo "  building as:    $(BUILD_TAG)" >&2; \
			echo "  An image stamped vX.Y.Z must be built from the vX.Y.Z tag." >&2; \
			echo "  This applies to BUILD_TARGET too: a remote build would push $(BUILD_TAG)" >&2; \
			echo "  — the newest tag repo-wide — from whatever ref HEAD is at." >&2; \
			echo "  Fix: git checkout $(BUILD_TAG)   (or set ALLOW_UNTAGGED_BUILD=1 for a scratch build)" >&2; \
			exit 1; \
		fi; \
		ref_commit=$$(git rev-parse --verify --quiet "$(BUILD_REF)^{commit}"); \
		tag_commit=$$(git rev-parse --verify --quiet "$(BUILD_TAG)^{commit}"); \
		if [ -n "$$ref_commit" ] && [ -n "$$tag_commit" ] && [ "$$ref_commit" != "$$tag_commit" ]; then \
			echo "ERROR: BUILD_REF $(BUILD_REF) is not $(BUILD_TAG)." >&2; \
			echo "  $(BUILD_TAG) is at: $${tag_commit}" >&2; \
			echo "  $(BUILD_REF) is at: $${ref_commit}" >&2; \
			echo "  A remote build clones BUILD_REF and pushes it as $(BUILD_TAG), so the" >&2; \
			echo "  tag would name a commit it does not contain. Unset BUILD_REF to use HEAD." >&2; \
			exit 1; \
		fi; \
	fi

.PHONY: check-go-mod
check-go-mod:
	@if [ -f "go.mod" ]; then \
		echo "go.mod found, running go mod vendor..."; \
		go mod vendor; \
	else \
		echo "go.mod not found, skipping go mod vendor."; \
	fi


.PHONY: fix
fix:
	@for dir in $$(find `pwd` -type d -name vendor -prune -o -name go.mod -exec dirname "{}" \; | grep -v '^$$'); do \
		cd $${dir}; \
		echo "fix $${dir}"; \
		go get github.com/go-git/go-git/v5@latest; \
		go get github.com/containerd/containerd@latest; \
		go get golang.org/x/crypto@latest; \
		go get golang.org/x/net@latest; \
	done