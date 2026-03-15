BINARY    := bin/ak
PKG       := github.com/ramaxlo/akcli/cmd
COMMIT    := $(shell git rev-parse --short HEAD)
BUILD_DATE := $(shell date -u +%Y-%m-%d)

LDFLAGS := -ldflags "\
	-X '$(PKG).commit=$(COMMIT)' \
	-X '$(PKG).buildDate=$(BUILD_DATE)'"

.PHONY: build clean

build:
	GOOS=linux go build $(LDFLAGS) -o $(BINARY) .

clean:
	rm -f $(BINARY)
