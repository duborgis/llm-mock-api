.PHONY: run build test test-contract test-bdd vet fmt tidy litellm-up litellm-down litellm-logs

run:
	go run ./cmd/server

build:
	go build -o bin/llm-mock-api ./cmd/server

# Linux/amd64 binary for the Dockerfile — built on the host to avoid depending on
# proxy.golang.org network access from inside the Docker build.
build-linux:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o bin/llm-mock-api-linux-amd64 ./cmd/server
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o bin/openmeter-mock-linux-amd64 ./cmd/openmeter-mock

test:
	go test ./...

test-contract:
	go test ./test/contract/... -v

# BDD (godog/Gherkin) suite that drives real HTTP calls through the already-running
# litellm-up stack and asserts OpenMeter events land for every mock route it covers.
# -count=1 disables the test cache: every run has real side effects against the live stack
# (POSTs through LiteLLM), so a "(cached)" result would silently skip them.
test-bdd:
	go test -tags bdd -count=1 ./test/bdd/... -v

vet:
	go vet ./...

fmt:
	gofmt -w .

tidy:
	go mod tidy

# Full local stack: mock-api + wiremock (fakes Google oauth2) + nginx (front door) + litellm.
# See README "Testing against LiteLLM" for what each service does and how to call it.
litellm-up: build-linux
	python3 scripts/gen-fake-vertex-credentials.py
	docker compose up -d --build

litellm-down:
	docker compose down

litellm-logs:
	docker compose logs -f
