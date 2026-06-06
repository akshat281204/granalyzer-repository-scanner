# Makefile

.PHONY: build install lint test release

build:
	go build -ldflags="-s -w" -o ./bin/granalyzer ./main.go

install:
	go install ./...

lint:
	golangci-lint run

test:
	go test ./... -v

release:
	goreleaser release --clean
