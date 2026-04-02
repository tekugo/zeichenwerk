.PHONY: build test lint fmt vet tidy demo showcase lazymake clean

## Build all packages
build:
	go build ./...

## Run all tests
test:
	go test ./...

## Run tests with coverage report
cover:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

## Run go vet
vet:
	go vet ./...

## Format all source files
fmt:
	gofmt -w .

## Tidy go.mod and go.sum
tidy:
	go mod tidy

## Run the demo app
demo:
	go run ./cmd/demo

## Run the showcase app
showcase:
	go run ./cmd/showcase

## Run lazymake
lazymake:
	go run ./cmd/lazymake

## Remove build artifacts
clean:
	rm -f coverage.out
