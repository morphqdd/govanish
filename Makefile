.PHONY: build test lint all

all: test lint build

build:
	go build -o bin/govanish ./cmd/govanish

test:
	go test ./...

lint: build
	./bin/govanish ./cmd/... ./internal/...
