.PHONY: all build test vet fmt check clean

all: build

build:
	go build -o mdt .

test:
	go test -race ./...

vet:
	go vet ./...

fmt:
	gofmt -w .

check: fmt vet test
	mdt -c README.md

clean:
	rm -f mdt
