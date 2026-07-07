GO ?= go
TOOLS_BIN ?= $(CURDIR)/.tools/bin
GOLANGCI_LINT_VERSION ?= v2.12.2
GOVULNCHECK_VERSION ?= v1.5.0

.PHONY: all fmt test test-race test-gui-deps lint audit tools verify

all: verify

fmt:
	$(GO) fmt ./...

test:
	$(GO) test ./...

test-race:
	$(GO) test -race ./...

test-gui-deps:
	$(GO) test -tags gui ./internal/adapters/gui

lint:
	$(TOOLS_BIN)/golangci-lint run ./...

audit:
	$(TOOLS_BIN)/govulncheck ./...

tools:
	GOBIN=$(TOOLS_BIN) $(GO) install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)
	GOBIN=$(TOOLS_BIN) $(GO) install golang.org/x/vuln/cmd/govulncheck@$(GOVULNCHECK_VERSION)

verify: fmt test test-race lint audit
