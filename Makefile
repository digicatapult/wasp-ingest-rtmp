
.PHONY: build lint

build:
	@go build ./...

lint:
	@golangci-lint run .

test:
	@go test ./...