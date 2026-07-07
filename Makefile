GO ?= go
TOOLS_BIN ?= $(CURDIR)/.tools/bin
GO_CACHE ?= $(CURDIR)/.tools/cache/go-build
GO_MOD_CACHE ?= $(CURDIR)/.tools/cache/go-mod
GO_ENV = GOCACHE=$(GO_CACHE) GOMODCACHE=$(GO_MOD_CACHE)
GOLANGCI_LINT_VERSION ?= v2.12.2
GOVULNCHECK_VERSION ?= v1.5.0

.PHONY: all fmt test test-race test-build test-gui-deps lint audit tools build run-built verify

all: verify

fmt:
	$(GO_ENV) $(GO) fmt ./...

test:
	$(GO_ENV) $(GO) test ./...

test-race:
	$(GO_ENV) $(GO) test -race ./...

test-build:
	python3 -m unittest discover -s tests -p '*_test.py'

test-gui-deps:
	$(GO_ENV) $(GO) test -tags gui ./internal/adapters/gui

build:
	$(GO_ENV) python3 scripts/build.py

run-built:
	artifact=$$(python3 scripts/build.py --print-path); python3 scripts/run_artifact.py "$$artifact" -- --smoke-test

lint:
	$(GO_ENV) $(TOOLS_BIN)/golangci-lint run ./...

audit:
	$(GO_ENV) $(TOOLS_BIN)/govulncheck ./...

tools:
	$(GO_ENV) GOBIN=$(TOOLS_BIN) $(GO) install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)
	$(GO_ENV) GOBIN=$(TOOLS_BIN) $(GO) install golang.org/x/vuln/cmd/govulncheck@$(GOVULNCHECK_VERSION)

verify: fmt test test-race test-build lint audit
