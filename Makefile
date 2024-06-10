.PHONY: all clean test

# Define the output directory for the compiled binary
OUTPUT_DIR := bin

# Define the name of the binary file
CLI_BINARY_NAME := chainnet-cli

# Define the source file for the CLI application
CLI_SOURCE := cmd/main.go

all: chainnet-cli test

chainnet-cli:
	@mkdir -p $(OUTPUT_DIR)
	@echo "Building chainnet CLI..."
	@go build -o $(OUTPUT_DIR)/$(CLI_BINARY_NAME) $(CLI_SOURCE)

test:
	@echo "Running tests..."
	@go test ./...

clean:
	@echo "Cleaning up..."
	@rm -rf $(OUTPUT_DIR)
	rm -f __debug_bin*
	rm -f _fixture/*
	rm -f main
	rm -f cli/chainnet-cli


.PHONY: debug
debug: 
	dlv debug --headless --listen=:2345 --api-version=2 --accept-multiclient


