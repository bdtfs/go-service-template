BIN := "./bin/service"
DOCKER_IMG := "go-service:develop"

RELEASE := "develop"
GIT_HASH := $(shell git log --format="%h" -n 1)
LDFLAGS := -X main.release=$(RELEASE) -X main.buildDate=$(shell date -u +%FT%TZ) -X main.gitHash=$(GIT_HASH)

.PHONY: lint test int-test build run tools codegen dc-reup dc-reb add-component

install-upgrade-lint:
	@curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin

lint: install-upgrade-lint
	golangci-lint run ./...

deps:
	go mod tidy

test:
	go clean --testcache && go test -count 1 -race ./...

int-test:
	go clean --testcache && go test --tags=integration -count 1 -race ./...

build: deps
	go build -v -o $(BIN) -ldflags "$(LDFLAGS)" ./cmd/service

run: build
	./bin/service

tools:
	go generate tools/tools.go

codegen:
	make tools
	go generate ./...

dc-reup:
	docker-compose down
	docker-compose up -d

dc-reb:
	docker-compose down
	docker-compose up --build -d

# Add a new infrastructure component to the project.
# Usage: make add-component COMPONENT=redis
add-component:
	@./scripts/add-component.sh $(COMPONENT)
