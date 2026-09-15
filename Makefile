.PHONY: build test lint fmt all

# testdata is excluded everywhere: fixtures are deliberately malformed,
# and gofmt would repair the very violations they exist to prove.
GOFILES = $(shell find cmd internal -name '*.go' -not -path '*/testdata/*') selflint_test.go

all: fmt test lint build

fmt:
	@test -z "$$(gofmt -l $(GOFILES))" || { echo "unformatted:"; gofmt -l $(GOFILES); exit 1; }

build:
	go build -o bin/govanish ./cmd/govanish

test:
	go test ./...

lint: build
	./bin/govanish ./cmd/... ./internal/...
