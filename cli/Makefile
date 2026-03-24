VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -ldflags "-X main.Version=$(VERSION)"

build-mac-arm:
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o blueprint-darwin-arm64 .

build-mac-intel:
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o blueprint-darwin-amd64 .

build-linux:
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o blueprint-linux-amd64 .

build-all: build-mac-arm build-mac-intel build-linux

clean:
	rm -f blueprint-darwin-arm64 blueprint-darwin-amd64 blueprint-linux-amd64

.PHONY: build-mac-arm build-mac-intel build-linux build-all clean
