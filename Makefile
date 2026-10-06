BINARY_NAME=clockify
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT?=$(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE?=$(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
LDFLAGS=-ldflags "-s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)"
GORELEASER?=goreleaser

.PHONY: build build-all clean test cover lint check run release-snapshot man perf help reviews review-list

help:
	@echo "Targets:"
	@echo "  build              Build $(BINARY_NAME) for current platform"
	@echo "  build-all          Cross-compile for all release platforms"
	@echo "  test               Run unit tests"
	@echo "  lint               Run golangci-lint"
	@echo "  check              lint + test"
	@echo "  release-snapshot   Local goreleaser snapshot build (no publish)"
	@echo "  run                go run with build metadata"
	@echo "  review-list        Show pending PLAN.md §5.7 commit reviews"
	@echo "  reviews            Run pending commit reviews via opencode + kims-team"

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
	rm -rf dist/

test:
	go test -race ./...

review-list:
	@bash scripts/review-commits.sh --list

reviews:
	@REVIEW_MODEL=$${REVIEW_MODEL:-opencode-go/mimo-v2.5} bash scripts/review-commits.sh

cover:
	go test ./... -coverprofile=cover.out -covermode=atomic
	go tool cover -func=cover.out | tail -1
	go run ./tools/coveragegate cover.out
	@echo "--- core packages ---"
	@go tool cover -func=cover.out | awk 'BEGIN{split("internal/config internal/output internal/errors internal/update internal/resolve internal/api pkg/duration pkg/clockify", a, " "); for (i in a) core[a[i]]=1} \
		{ split($$1, p, ":"); path=$$1; sub(/:[0-9]+.*/, "", path); sub(/\/[^\/]*$$/, "", path); \
		  if (path in core) { cov=$$NF; gsub("%","",cov); s[path]+=cov; n[path]++ } } \
		END { for (p in s) printf "%s: %.1f%%\n", p, s[p]/n[p] }' | sort

man:
	go run ./tools/man.go

perf: build
	@echo "Binary size:"; ls -la $(BINARY_NAME)
	@echo "Startup:"; /usr/bin/time -f '%e sec' ./$(BINARY_NAME) version 2>&1 || time ./$(BINARY_NAME) version

lint:
	golangci-lint run || go vet ./...

check: lint test

run:
	go run $(LDFLAGS) ./cmd/clockify

release-snapshot:
	$(GORELEASER) release --snapshot --clean --skip=publish
