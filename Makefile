PROGRAM_NAME = miaou
MAIN_FILE = ./

GIT_COMMIT=$(shell git rev-parse --short HEAD)
GIT_TAG=$(shell git describe --tags |cut -d- -f1)

LDFLAGS = -ldflags "-X github.com/mirwide/${PROGRAM_NAME}/cmd.Version=${GIT_TAG}(${GIT_COMMIT})"

.PHONY: help clean dep build lint generate

.DEFAULT_GOAL := help

help: ## Display this help screen
	@echo "Makefile available targets:"
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  * \033[36m%-15s\033[0m %s\n", $$1, $$2}'

dep: ## Download the dependencies
	go mod tidy
	go mod download

generate: dep
	go get golang.org/x/text/cmd/gotext
	go install golang.org/x/text/cmd/gotext
	go generate ./internal/translations/translations.go

build: dep ## Build for linux
	mkdir -p ./bin
	CGO_ENABLED=0 GOOS=linux go build ${LDFLAGS} -o bin/${PROGRAM_NAME} ${MAIN_FILE}

build-amd64: dep generate ## Build for linux
	mkdir -p ./bin
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build ${LDFLAGS} -o bin/${PROGRAM_NAME}-${GIT_TAG}-linux-amd64 ${MAIN_FILE}

build-riscv: dep generate ## Build for risc-v
	mkdir -p ./bin
	CC=riscv64-linux-gnu-gcc CGO_ENABLED=0 GOOS=linux GOARCH=riscv64 go build ${LDFLAGS} -o bin/${PROGRAM_NAME}-${GIT_TAG}-linux-riscv64 ${MAIN_FILE}

clean: ## Clean build directory
	rm -rf ./bin

lint: dep ## Lint the source files
	GO111MODULE=on golangci-lint run
	gosec -quiet ./...

test: dep ## Run tests
	go test -race -p 1 -timeout 300s -coverprofile=coverage.txt ./... && \
    	go tool cover -func=coverage.txt | tail -n1 | awk '{print "Total test coverage: " $$3}'
	@rm coverage.txt
