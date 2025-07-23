REGISTRY ?= docker.io
IMAGE ?= bborbe/kafka-topic-reader
BRANCH ?= $(shell git rev-parse --abbrev-ref HEAD)
DIRS += $(shell find */* -maxdepth 0 -name Makefile -exec dirname "{}" \;)

default: precommit

build:
	docker build --no-cache --rm=true --platform=linux/amd64 -t $(REGISTRY)/$(IMAGE):$(BRANCH) -f Dockerfile .

upload:
	docker push $(REGISTRY)/$(IMAGE):$(BRANCH)

clean:
	docker rmi $(REGISTRY)/$(IMAGE):$(BRANCH) || true

precommit: ensure format generate test check
	@echo "ready to commit"

ensure:
	go mod tidy
	go mod verify
	rm -rf vendor

format:
	go run -mod=mod github.com/incu6us/goimports-reviser/v3 -project-name github.com/bborbe/kafka-topic-reader -format -excludes vendor ./...

generate:
	rm -rf mocks avro
	go generate -mod=mod ./...

test:
	go test -mod=mod -p=$${GO_TEST_PARALLEL:-1} -cover -race $(shell go list -mod=mod ./... | grep -v /vendor/)

check: vet errcheck vulncheck

vet:
	go vet -mod=mod $(shell go list -mod=mod ./... | grep -v /vendor/)

errcheck:
	go run -mod=mod github.com/kisielk/errcheck -ignore '(Close|Write|Fprint)' $(shell go list -mod=mod ./... | grep -v /vendor/)

apply:
	@for i in $(DIRS); do \
		cd $$i; \
		echo "apply $${i}"; \
		make apply; \
		cd ..; \
	done

vulncheck:
	go run -mod=mod golang.org/x/vuln/cmd/govulncheck $(shell go list -mod=mod ./... | grep -v /vendor/)
