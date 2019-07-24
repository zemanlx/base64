SHELL := /bin/bash

.DEFAULT_GOAL: build

# These will be provided to the target
VERSION := $(shell git describe --tags 2>/dev/null || echo "0.0.0")
COMMIT := $(shell git rev-parse --short=8 HEAD)

# Use linker flags to provide version/build settings to the target
LDFLAGS=-ldflags "-s -w -X=main.version=$(VERSION) -X=main.commit=$(COMMIT)"

.PHONY: build
build:
	@env GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o build/base64

.PHONY: clean
clean:
	@rm -f $(TARGET)

.PHONY: install
install:
	@go install $(LDFLAGS)

.PHONY: lint
lint:
	@golangci-lint --color always run

.PHONY: simplify
simplify:
	@gofmt -s -l -w .

.PHONY: test
test:
	@go test -race -coverprofile=coverage.txt -covermode=atomic ./...

.PHONY: go-fuzz
go-fuzz:
	@if ! which go-fuzz &>/dev/null; then \
		go get github.com/dvyukov/go-fuzz/go-fuzz \
		&& go mod tidy \
		; \
	fi

.PHONY: test-fuzzing
test-fuzzing: go-fuzz
	@(cd xbase && go-fuzz-build github.com/zemanlx/base64/xbase)

.PHONY: fuzz-xbase
fuzz-xbase: test-fuzzing
	@go-fuzz -bin=xbase/xbase-fuzz.zip -workdir=xbase/fuzz

.PHONY: integration-test
integration-test: build
	@./integration_test.sh
