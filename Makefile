# Define the output directory for the compiled binary
OUTPUT_DIR := bin

# Define the name of the CLI binary file
CLI_BINARY_NAME := chainnet-cli
NODE_BINARY_NAME := chainnet

# Define the source file for the CLI application
CLI_SOURCE := $(wildcard cmd/cli/*.go)
NODE_SOURCE := $(wildcard cmd/node/*.go)

# Define build flags
GCFLAGS := -gcflags "all=-N -l"

.PHONY: all
all: test lint chainnet-cli chainnet-node

.PHONY: chainnet-cli
chainnet-cli:
	@mkdir -p $(OUTPUT_DIR)
	@echo "Building chainnet CLI..."
	@go build $(GCFLAGS) -o $(OUTPUT_DIR)/$(CLI_BINARY_NAME) $(CLI_SOURCE)

.PHONY: chainnet-node
chainnet-node:
	@mkdir -p $(OUTPUT_DIR)
	@echo "Building chainnet node..."
	@go build $(GCFLAGS) -o $(OUTPUT_DIR)/$(NODE_BINARY_NAME) $(NODE_SOURCE)

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
	rm -f __debug_bin*
	rm -f _fixture/*


.PHONY: debug
debug: chainnet-node
	dlv --listen=:2345 --headless=true --api-version=2 --accept-multiclient exec ./bin/chainnet

