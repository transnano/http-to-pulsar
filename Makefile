GIT_BRANCH := $(shell git rev-parse --abbrev-ref HEAD)
GIT_SHA    := $(shell git rev-parse --short HEAD)

.PHONY: build
build: ## Build go binary
	CGO_ENABLED=0 GOOS=linux go build -installsuffix cgo -o prometheus-pulsar-remote-write

.PHONY: test
test: ## Run all tests
	go test -race ./...

.PHONY: bench
bench: ## Run all benchmarks
	go test -bench . ./...

.PHONY: lint
lint: ## Lint
	golangci-lint run ./...
