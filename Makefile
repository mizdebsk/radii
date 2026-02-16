VERSION := $(shell git describe --tags --dirty --always 2>/dev/null || echo "dev")

all: build

build:
	go build \
	    -ldflags="-X main.version=$(VERSION)" \
	    -o dist/bin/radii \
	    ./cmd/radii

test:
	go test ./...

archive:
	go mod vendor
	tar czf radii-$(VERSION)-vendor.tar.gz vendor
	git archive HEAD --prefix radii-$(VERSION)/ | gzip -9 >radii-$(VERSION).tar.gz

clean:
	rm -rf dist

.PHONY: all build test archive clean
