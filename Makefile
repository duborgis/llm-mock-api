.PHONY: run build test test-contract vet fmt tidy

run:
	go run ./cmd/server

build:
	go build -o bin/llm-mock-api ./cmd/server

test:
	go test ./...

test-contract:
	go test ./test/contract/... -v

vet:
	go vet ./...

fmt:
	gofmt -w .

tidy:
	go mod tidy
