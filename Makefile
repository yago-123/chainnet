# Define the output directory for the compiled binary
OUTPUT_DIR := bin

# Define the name of the CLI binary file
CLI_BINARY_NAME   := chainnet-cli
MINER_BINARY_NAME := chainnet-miner
NESPV_BINARY_NAME := chainnet-nespv
NODE_BINARY_NAME  := chainnet-node

# Define the source file for the CLI application
CLI_SOURCE := $(wildcard cmd/cli/*.go)
MINER_SOURCE := $(wildcard cmd/miner/*.go)
NESPV_SOURCE := $(wildcard cmd/nespv/*go)
NODE_SOURCE := $(wildcard cmd/node/*.go)

# Define build flags
GCFLAGS := -gcflags "all=-N -l"

# Docker image names and paths
DOCKER_IMAGE_MINER := yagoninja/chainnet-miner:latest
DOCKER_IMAGE_NODE  := yagoninja/chainnet-node:latest
DOCKERFILE_MINER   := ./build/docker/miner/Dockerfile
DOCKERFILE_NODE    := ./build/docker/node/Dockerfile

.PHONY: all
all: test lint cli miner node nespv

.PHONY: cli
cli: output-dir
	@echo "Building chainnet CLI..."
	@go build $(GCFLAGS) -o $(OUTPUT_DIR)/$(CLI_BINARY_NAME) $(CLI_SOURCE)

.PHONY: miner
miner: output-dir
	@echo "Building chainnet miner..."
	@go build $(GCFLAGS) -o $(OUTPUT_DIR)/$(MINER_BINARY_NAME) $(MINER_SOURCE)

.PHONY: node
node: output-dir
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
debug: node
	dlv --listen=:2345 --headless=true --api-version=2 --accept-multiclient exec ./bin/chainnet-miner

.PHONY: miner-image
miner-image: miner
	@echo "Building Docker image for chainnet miner..."
	docker build -t $(DOCKER_IMAGE_MINER) -f $(DOCKERFILE_MINER) .

.PHONY: node-image
node-image: node
	@echo "Building Docker image for chainnet node..."
	docker build -t $(DOCKER_IMAGE_NODE) -f $(DOCKERFILE_NODE) .

.PHONY: images
images: miner-image node-image