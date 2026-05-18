SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

.PHONY: tidy test vet build fmt all

all: vet build test

tidy:
	go mod tidy

vet:
	go vet ./...

build:
	go build ./...

test:
	go test ./... -race -count=1

fmt:
	gofmt -s -w .
