# Define the output directory for the compiled binary
OUTPUT_DIR := bin

# Define the name of the CLI binary file
CLI_BINARY_NAME := chainnet-cli
NODE_BINARY_NAME := chainnet
NESPV_BINARY_NAME := nespv

# Define the source file for the CLI application
CLI_SOURCE := $(wildcard cmd/cli/*.go)
NODE_SOURCE := $(wildcard cmd/node/*.go)
NESPV_SOURCE := $(wildcard cmd/nespv/*go)

# Define build flags
GCFLAGS := -gcflags "all=-N -l"

.PHONY: all
all: test lint chainnet-cli chainnet-node nespv

.PHONY: chainnet-cli
chainnet-cli: output-dir
	@echo "Building chainnet CLI..."
	@go build $(GCFLAGS) -o $(OUTPUT_DIR)/$(CLI_BINARY_NAME) $(CLI_SOURCE)

.PHONY: chainnet-node
chainnet-node: output-dir
	@echo "Building chainnet node..."
	@go build $(GCFLAGS) -o $(OUTPUT_DIR)/$(NODE_BINARY_NAME) $(NODE_SOURCE)

.PHONY: nespv
nespv: output-dir
	@echo "Building chainnet nespv..."
	@go build $(GCFLAGS) -o $(OUTPUT_DIR)/$(NESPV_BINARY_NAME) $(NESPV_SOURCE)

.PHONY: output-dir
output-dir:
	@mkdir -p $(OUTPUT_DIR)
.PHONY: lint
lint:
	@echo "Running linter..."
	@golangci-lint run ./...

.PHONY: test
test:
	@echo "Running tests..."
	@go test -v -cover ./...

.PHONY: clean
clean:
	@echo "Cleaning up..."
	@rm -rf $(OUTPUT_DIR)
	@rm -f __debug_bin*
	@rm -f _fixture/*

.PHONY: imports
imports: 
	@find . -name "*.go" | xargs goimports -w

.PHONY: debug
debug: chainnet-node
	dlv --listen=:2345 --headless=true --api-version=2 --accept-multiclient exec ./bin/chainnet

