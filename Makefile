GO ?= go
TOOLS_BIN ?= $(CURDIR)/.tools/bin
GO_CACHE ?= $(CURDIR)/.tools/cache/go-build
GO_MOD_CACHE ?= $(CURDIR)/.tools/cache/go-mod
GO_ENV = GOCACHE=$(GO_CACHE) GOMODCACHE=$(GO_MOD_CACHE)
HOST_GOOS = $(shell uname -s | tr A-Z a-z)
HOST_GOARCH_RAW = $(shell uname -m)
HOST_GOARCH = $(shell printf '%s' "$(HOST_GOARCH_RAW)" | sed 's/^x86_64$$/amd64/;s/^aarch64$$/arm64/')
GOLANGCI_LINT_VERSION ?= v2.12.2
GOVULNCHECK_VERSION ?= v1.5.0

.PHONY: all fmt test test-race test-build test-gui-deps check-code-size check-docs lint audit tools build build-android check-android-build-env run-built run-built-headless verify

all: verify

fmt:
	$(GO_ENV) $(GO) fmt ./...

test:
	$(GO_ENV) $(GO) test ./...

test-race:
	$(GO_ENV) $(GO) test -race ./...

test-build:
	python3 -m unittest discover -s tests -p '*_test.py'

check-code-size:
	python3 scripts/check_code_file_lines.py

check-docs:
	python3 scripts/check_directory_docs.py

test-gui-deps:
	$(GO_ENV) $(GO) test -tags gui ./internal/adapters/gui

build:
	$(GO_ENV) python3 scripts/build.py

check-android-build-env:
	$(GO_ENV) python3 scripts/build.py

build-android:
	$(GO_ENV) python3 scripts/build.py

run-built:
	python3 scripts/run_artifact.py "releases/$(HOST_GOOS)/$(HOST_GOARCH)/apparat/latest" -- --smoke-test --runtime-dir /tmp/apparat-run-built/apparat

run-built-headless:
	python3 scripts/run_artifact.py "releases/$(HOST_GOOS)/$(HOST_GOARCH)/apparatd/latest" -- --smoke-test --runtime-dir /tmp/apparat-run-built/apparatd

lint:
	$(GO_ENV) $(TOOLS_BIN)/golangci-lint run ./...

audit:
	$(GO_ENV) $(TOOLS_BIN)/govulncheck ./...

tools:
	$(GO_ENV) GOBIN=$(TOOLS_BIN) $(GO) install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)
	$(GO_ENV) GOBIN=$(TOOLS_BIN) $(GO) install golang.org/x/vuln/cmd/govulncheck@$(GOVULNCHECK_VERSION)

verify: fmt test test-race test-build check-code-size check-docs lint audit
