all: build test check

modules:
	go mod tidy

build: modules
	go build -o bin/orderservice ./cmd/orderservice

test:
	go test ./...

check:
	golangci-lint run