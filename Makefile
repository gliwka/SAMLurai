.PHONY: build test test-verbose test-coverage clean install lint fmt help update-golden

# Binary name
BINARY_NAME=samlurai
VERSION?=dev
LDFLAGS=-ldflags "-X github.com/gliwka/SAMLurai/cmd.version=$(VERSION)"

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=$(GOCMD) fmt

# Default target
all: build

## build: Build the binary
build:
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME) .

## install: Install the binary to $GOPATH/bin
install:
	$(GOCMD) install $(LDFLAGS) .

## test: Run all tests
test:
	$(GOTEST) -v ./...

## test-short: Run tests without verbose output
test-short:
	$(GOTEST) ./...

## test-coverage: Run tests with coverage report
test-coverage:
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

## test-race: Run tests with race detector
test-race:
	$(GOTEST) -v -race ./...

## update-golden: Update golden test files
update-golden:
	$(GOTEST) -v ./... -update

## lint: Run linter
lint:
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed. Run: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

## fmt: Format code
fmt:
	$(GOFMT) ./...

## tidy: Tidy go.mod
tidy:
	$(GOMOD) tidy

## clean: Clean build artifacts
clean:
	rm -f $(BINARY_NAME)
	rm -f coverage.out coverage.html

## deps: Download dependencies
deps:
	$(GOMOD) download

## run: Run the application
run: build
	./$(BINARY_NAME)

## help: Show this help
help:
	@echo "SAMLurai - SAML Assertion Decoder & Debugger"
	@echo ""
	@echo "Usage:"
	@echo "  make [target]"
	@echo ""
	@echo "Targets:"
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'

# Example targets for development
## example-decode: Run example decode command
example-decode: build
	@echo "PHNhbWw+dGVzdDwvc2FtbD4=" | ./$(BINARY_NAME) decode

## example-inspect: Run example inspect command  
example-inspect: build
	./$(BINARY_NAME) inspect -f testdata/fixtures/assertions/response.xml
