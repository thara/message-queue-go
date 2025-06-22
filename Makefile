.PHONY: all build test lint fmt clean run-example docker-up docker-down

all: build

build:
	go build -o bin/example ./cmd/example

test:
	go test -v -race ./...

test-coverage:
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

lint:
	@which golangci-lint > /dev/null || (echo "Installing golangci-lint..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	golangci-lint run ./...

fmt:
	go fmt ./...
	go mod tidy

clean:
	rm -rf bin/ coverage.out coverage.html

run-example: build docker-up
	./bin/example

docker-up:
	docker-compose up -d redis

docker-down:
	docker-compose down

deps:
	go mod download
	go mod verify