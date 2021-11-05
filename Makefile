.PHONY: build plugin check

LDFLAGS = $(shell ./version.sh)

default: build

build: kubectl-kruise

kubectl-kruise:
	GO111MODULE=on CGO_ENABLED=0 GOOS=${GOOS} go build -ldflags "$(LDFLAGS)" -o bin/kubectl-kruise cmd/plugin/main.go

test:
	find . -iname '*.go' -type f | grep -v /vendor/ | xargs gofmt -l
	GO111MODULE=on go test -v -race ./...
	go vet ./...
