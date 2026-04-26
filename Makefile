MODULE  := github.com/bikramdhoju/ecsctl
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT  := $(shell git rev-parse --short HEAD 2>/dev/null || echo none)
DATE    := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

build:
	@go build -ldflags "-X $(MODULE)/pkg/version.Version=$(VERSION) \
	                    -X $(MODULE)/pkg/version.Commit=$(COMMIT) \
	                    -X $(MODULE)/pkg/version.Date=$(DATE) \
	                    -s -w" -o ecsctl .

deploy: build
	@mv -v ecsctl /usr/local/bin/
	@echo "Deployed Successfully"

test:
	@go test ./...

clean:
	@rm -f ecsctl
