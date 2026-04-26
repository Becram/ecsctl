BINARY    := ecsctl
MODULE    := github.com/bikramdhoju/ecsctl
VERSION   ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT    := $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE      := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS   := -ldflags "-X $(MODULE)/pkg/version.Version=$(VERSION) \
                        -X $(MODULE)/pkg/version.Commit=$(COMMIT) \
                        -X $(MODULE)/pkg/version.Date=$(DATE) \
                        -s -w"

.PHONY: build install test lint clean tidy

build:
	go build $(LDFLAGS) -o $(BINARY) .

install:
	go install $(LDFLAGS) .

test:
	go test ./...

lint:
	golangci-lint run ./...

tidy:
	go mod tidy

clean:
	rm -f $(BINARY)
