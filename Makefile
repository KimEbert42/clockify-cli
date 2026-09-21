BINARY_NAME=clockify
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT?=$(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE?=$(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
LDFLAGS=-ldflags "-s -w -X github.com/kim/clockify-cli/internal/cli.buildVersion=$(VERSION) -X github.com/kim/clockify-cli/internal/cli.buildCommit=$(COMMIT) -X github.com/kim/clockify-cli/internal/cli.buildDate=$(DATE)"

.PHONY: build build-all clean test lint check run

build:
	go build $(LDFLAGS) -o $(BINARY_NAME) ./cmd/clockify

build-all:
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BINARY_NAME)-linux-amd64 ./cmd/clockify
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(BINARY_NAME)-linux-arm64 ./cmd/clockify
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BINARY_NAME)-darwin-arm64 ./cmd/clockify
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BINARY_NAME)-darwin-amd64 ./cmd/clockify
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BINARY_NAME)-windows-amd64.exe ./cmd/clockify

clean:
	rm -f $(BINARY_NAME) $(BINARY_NAME)-*

test:
	go test ./...

lint:
	golangci-lint run

check: lint test

run:
	go run $(LDFLAGS) ./cmd/clockify
