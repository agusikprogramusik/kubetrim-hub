VERSION?=0.2.0-dev
GIT_COMMIT?=$(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
BUILD_DATE?=$(shell date -u +'%Y-%m-%dT%H:%M:%SZ')
BUILD_DIR=bin
DOCKER_CACHE_VOLUMES=-v kubetrim-go-cache:/root/.cache/go-build -v kubetrim-go-mod:/go/pkg/mod

.PHONY: all build test test-docker coverage fmt vet clean

all: build

build:
	@mkdir -p $(BUILD_DIR)
	go build -buildvcs=false -o $(BUILD_DIR)/hub-api ./apps/api
	go build -buildvcs=false -o $(BUILD_DIR)/hub-worker ./apps/worker

test:
	go test -v -race ./...

test-docker:
	docker run --rm -v $(shell pwd):/app $(DOCKER_CACHE_VOLUMES) -w /app golang:latest go test -v -race -cover ./...

coverage:
	docker run --rm -v $(shell pwd):/app $(DOCKER_CACHE_VOLUMES) -w /app golang:latest go test -coverprofile=coverage.out ./...

fmt:
	go fmt ./...

vet:
	go vet ./...

clean:
	rm -rf $(BUILD_DIR) coverage.out
