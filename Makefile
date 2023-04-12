GOCMD		?= go
GOTEST	?= $(GOCMD) test
GOVET		?= $(GOCMD) vet
BINARY	?= tsuru
VERSION	?= $(shell git describe --tags --dirty --match='v*' 2> /dev/null || echo dev)
FILES		?= $(shell find . -type f -name '*.go')

GREEN  := $(shell tput -Txterm setaf 2)
YELLOW := $(shell tput -Txterm setaf 3)
WHITE  := $(shell tput -Txterm setaf 7)
CYAN   := $(shell tput -Txterm setaf 6)
RESET  := $(shell tput -Txterm sgr0)

.PHONY: all test build coverage

default: help

## Build:
build: ## Build your project and put the output binary in build/
	$(GOCMD) build -ldflags "-s -w -X 'main.version=$(VERSION)'" -o build/$(BINARY) .

install: build  ## Build your project and install the binary in $GOPATH/bin/
	rm -f "$(shell go env GOPATH)/bin/$(BINARY)"
	cp build/$(BINARY) "$(shell go env GOPATH)/bin/$(BINARY)"

clean: ## Remove build related file
	rm -fr ./build
	rm -fr ./coverage

clean-all: ## Remove build related file and installed binary
	rm -fr ./build
	rm -fr ./coverage
	rm -f "$(shell go env GOPATH)/bin/$(BINARY)"

## Test:
test: ## Run the tests of the project (fastest)
	$(GOVET) ./...
	$(GOTEST) -v ./...

test-ci: ## Run ALL the tests of the project (+race)
	$(GOVET) ./...
	$(GOTEST) -v -race ./...

test-coverage: ## Run the tests of the project and export the coverage
	rm -fr coverage && mkdir coverage
	$(GOTEST) -cover -covermode=atomic -coverprofile=coverage/coverage.out ./...
	@echo ""
	$(GOCMD) tool cover -func=coverage/coverage.out
	@echo ""
	$(GOCMD) tool cover -func=coverage/coverage.out -o coverage/coverage.txt
	$(GOCMD) tool cover -html=coverage/coverage.out -o coverage/coverage.html

coverage: test-coverage  ## Run test-coverage and open coverage in your browser
	$(GOCMD) tool cover -html=coverage/coverage.out

## Lint:
lint: lint-go  ## Run all available linters

lint-go: ## Use golintci-lint on your project
	echo $(GOCMD) vet ./...
ifeq (, $(shell which staticcheck))
	go install honnef.co/go/tools/cmd/staticcheck@latest
endif
	staticcheck ./...

## Help:
help: ## Show this help.
	@echo ''
	@echo 'Usage:'
	@echo '  ${YELLOW}make${RESET} ${GREEN}<target>${RESET}'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} { \
		if (/^[a-zA-Z_-]+:.*?##.*$$/) {printf "    ${YELLOW}%-20s${GREEN}%s${RESET}\n", $$1, $$2} \
		else if (/^## .*$$/) {printf "  ${CYAN}%s${RESET}\n", substr($$1,4)} \
		}' $(MAKEFILE_LIST)

env:    ## Print useful environment variables to stdout
	@echo '$$(GOCMD)   :' $(GOCMD)
	@echo '$$(GOTEST)  :' $(GOTEST)
	@echo '$$(GOVET)   :' $(GOVET)
	@echo '$$(BINARY)  :' $(BINARY)
	@echo '$$(VERSION) :' $(VERSION)
	@echo '$$(FILES#)  :' $(shell echo $(FILES) | wc -w)
