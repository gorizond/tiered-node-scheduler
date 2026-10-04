SHELL := /bin/bash
BIN_NAME := tiered-node-scheduler
IMAGE_REPO ?= ghcr.io/gorizond/tiered-node-scheduler
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "v0.1.0")
GIT_COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "dev")
LDFLAGS := -s -w -X k8s.io/component-base/version.gitVersion=$(VERSION) -X k8s.io/component-base/version.gitCommit=$(GIT_COMMIT)

GO_BUILD_ENV := CGO_ENABLED=0

.PHONY: all
all: test build

.PHONY: test
test:
	go vet ./...
	go test -v -race -cover ./pkg/...

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: build
build:
	$(GO_BUILD_ENV) go build -ldflags="$(LDFLAGS)" -o bin/$(BIN_NAME) ./cmd/scheduler

.PHONY: build-linux-amd64
build-linux-amd64:
	$(GO_BUILD_ENV) GOOS=linux GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o bin/$(BIN_NAME)-linux-amd64 ./cmd/scheduler

.PHONY: build-linux-arm64
build-linux-arm64:
	$(GO_BUILD_ENV) GOOS=linux GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o bin/$(BIN_NAME)-linux-arm64 ./cmd/scheduler

.PHONY: docker-build
docker-build:
	docker build \
		--build-arg VERSION=$(VERSION) \
		--build-arg GIT_COMMIT=$(GIT_COMMIT) \
		-t $(IMAGE_REPO):$(VERSION) \
		-t $(IMAGE_REPO):latest .

.PHONY: docker-buildx
docker-buildx:
	docker buildx build \
		--platform linux/amd64,linux/arm64 \
		--build-arg VERSION=$(VERSION) \
		--build-arg GIT_COMMIT=$(GIT_COMMIT) \
		-t $(IMAGE_REPO):$(VERSION) \
		-t $(IMAGE_REPO):latest .

.PHONY: run-local
run-local: build
	./bin/$(BIN_NAME) --config=./config/local-config.yaml --v=4

.PHONY: clean
clean:
	rm -rf bin/
