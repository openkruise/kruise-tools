.PHONY: build plugin check

LDFLAGS = $(shell ./version.sh)

default: build

build: kubectl-kruise resource-distribution-generator

kubectl-kruise: fmt vet
	GO111MODULE=on CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o bin/kubectl-kruise cmd/plugin/main.go

resource-distribution-generator: fmt vet
	GO111MODULE=on CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o bin/resourcedistributiongenerator cmd/resourcedistributiongenerator/main.go

test:
	find . -iname '*.go' -type f | grep -v /vendor/ | xargs gofmt -l
	GO111MODULE=on go test -v -race ./...
	go vet ./...

fmt: ## Run go fmt against code.
	go fmt ./...

vet: ## Run go vet against code.
	go vet ./...