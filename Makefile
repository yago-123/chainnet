# Define the output directory for the compiled binary
OUTPUT_DIR := bin

# Define the name of the CLI binary file
CLI_BINARY_NAME := chainnet-cli
NODE_BINARY_NAME := chainnet
WALLET_BINARY_NAME := wallet

# Define the source file for the CLI application
CLI_SOURCE := $(wildcard cmd/cli/*.go)
NODE_SOURCE := $(wildcard cmd/node/*.go)
WALLET_SOURCE := $(wildcard cmd/wallet/*go)

# Define build flags
GCFLAGS := -gcflags "all=-N -l"

.PHONY: all
all: test lint chainnet-cli chainnet-node chainnet-wallet

.PHONY: chainnet-cli
chainnet-cli: output-dir
	@echo "Building chainnet CLI..."
	@go build $(GCFLAGS) -o $(OUTPUT_DIR)/$(CLI_BINARY_NAME) $(CLI_SOURCE)

.PHONY: chainnet-node
chainnet-node: output-dir
	@echo "Building chainnet node..."
	@go build $(GCFLAGS) -o $(OUTPUT_DIR)/$(NODE_BINARY_NAME) $(NODE_SOURCE)

.PHONY: chainnet-wallet
chainnet-wallet: output-dir
	@echo "Building chainnet wallet..."
	@go build $(GCFLAGS) -o $(OUTPUT_DIR)/$(WALLET_BINARY_NAME) $(WALLET_SOURCE)

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

