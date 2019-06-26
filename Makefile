SHELL := /bin/bash

.DEFAULT_GOAL: build

# These will be provided to the target
VERSION := $(shell git describe --tags 2>/dev/null || echo "0.0.0")
COMMIT := $(shell git rev-parse --short=8 HEAD)

# Use linker flags to provide version/build settings to the target
LDFLAGS=-ldflags "-X=main.version=$(VERSION) -X=main.commit=$(COMMIT) -extldflags -static -s -w"
# LDFLAGS=-ldflags "-X=main.version=$(VERSION) -X=main.commit=$(COMMIT) -linkmode external -extldflags -static -s -w"

.PHONY: build
build:
	@env GOOS=linux   GOARCH=amd64 go build $(LDFLAGS) -o build/base64.linux.amd64

.PHONY: build-all
build-all:
	env GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o build/base64.windows.amd64
	env GOOS=windows GOARCH=386   go build $(LDFLAGS) -o build/base64.windows.386
	env GOOS=linux   GOARCH=amd64 go build $(LDFLAGS) -o build/base64.linux.amd64
	env GOOS=linux   GOARCH=386   go build $(LDFLAGS) -o build/base64.linux.386
	env GOOS=darwin  GOARCH=amd64 go build $(LDFLAGS) -o build/base64.darwin.amd64
	env GOOS=darwin  GOARCH=386   go build $(LDFLAGS) -o build/base64.darwin.386

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

.PHONY: fuzz-xbase
fuzz-xbase:
	@go-fuzz -bin=xbase/xbase-fuzz.zip -workdir=xbase/fuzz
	@(cd xbase && go-fuzz-build github.com/zemanlx/base64/xbase)

.PHONY: integration-test
integration-test: build
	@./integration_test.sh
