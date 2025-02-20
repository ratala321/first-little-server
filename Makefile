.PHONY: fmt vet build

all: fmt vet build

fmt:
	go fmt ./...

vet: fmt
	go vet ./...

build: vet
	go build
