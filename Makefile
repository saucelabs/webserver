# Copyright 2021 The webserver Authors. All rights reserved.
# Use of this source code is governed by a MIT
# license that can be found in the LICENSE file.

BINDIR := $(CURDIR)/bin
HAS_GOLANGCI := $(shell command -v golangci-lint;)
HAS_GODOC := $(shell command -v godoc;)

default: ci

build:
	@GOBIN=$(BINDIR) go install -race ./... && echo "Build OK"

lint:
ifndef HAS_GOLANGCI
	$(error You must install github.com/golangci/golangci-lint)
endif
	@golangci-lint run -v -c .golangci.yml && echo "Lint OK"

test:
	@go test -timeout 30s -short -v -race -cover -coverprofile=coverage.out ./...

coverage:
	@go tool cover -func=coverage.out

doc:
ifndef HAS_GODOC
	$(error You must install godoc, run "go get golang.org/x/tools/cmd/godoc")
endif
	@echo "Open localhost:6060/pkg/github.com/saucelabs/webserver/ in your browser\n"
	@godoc -http :6060

ci: lint test coverage

.PHONY: lint test test-integration coverage doc ci
