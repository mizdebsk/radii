VERSION := $(shell git describe --tags --dirty --always 2>/dev/null || echo "dev")

all: build

build:
	go build \
	    -ldflags="-X main.version=$(VERSION)" \
	    -o dist/bin/radii \
	    ./cmd/radii

test:
	go test ./...

clean:
	rm -rf dist

.PHONY: all build test clean
