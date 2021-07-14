.PHONY: build plugin check

LDFLAGS = $(shell ./version.sh)
GOENV  := GO15VENDOREXPERIMENT="1" GO111MODULE=on CGO_ENABLED=0 GOOS=${GOOS} GOARCH=amd64
GO := $(GOENV) go

default: build

build: plugin

plugin:
	GO111MODULE=on CGO_ENABLED=0 GOOS=${GOOS} go build -ldflags '$(LDFLAGS)' -o bin/kubectl-kruise cmd/plugin/main.go


check:
	find . -iname '*.go' -type f | grep -v /vendor/ | xargs gofmt -l
	GO111MODULE=on go test -v -race ./...
	$(GO) vet ./...
