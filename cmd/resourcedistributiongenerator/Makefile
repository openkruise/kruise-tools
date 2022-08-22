BINARY="resourcedistributiongenerator"

.PHONY: test fmt vet build clean

default: build

test:
	go test -v ./...

fmt:
	go fmt ./...

vet:
	go vet ./...

build:
	go build -v -o ${BINARY}

clean:
	@if [ -f ${BINARY} ] ; then rm ${BINARY} ; fi