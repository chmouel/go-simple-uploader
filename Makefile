PROJECT_NAME := "go-simple-uploader"
PKG := "github.com/chmouel/$(PROJECT_NAME)"
PKG_LIST := $(shell go list ${PKG}/... | grep -v /vendor/)
GO_FILES := $(shell find . -name '*.go' | grep -v /vendor/ | grep -v _test.go)

.PHONY: all dep lint vet test test-coverage build clean

all: build

lint: ## Lint Golang files
	@golangci-lint run uploader/

vet: ## Run go vet
	@go vet ${PKG_LIST}

dev: ## Run reflex to dev easily with easy!
	@cd uploader && \
		reflex -v -s -g '*.go' -- sh -c 'go run ../main.go'

test: ## Run unittests
	@go test -short ${PKG_LIST}

test-coverage: ## Run tests with coverage
	@go test -short -coverprofile cover.out -covermode=atomic ${PKG_LIST}
	@cat cover.out >> coverage.txt

build: dep ## Build the binary file
	@go build -o build/$(PROJECT_NAME) $(PKG)

clean: ## Remove previous build
	@rm -f $(PROJECT_NAME)/build

help: ## Display this help screen
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
