# Define directories
OUTPUT_DIR := bin
NODE_PROTOBUF_DIR := pkg/network/protobuf
OPENAPI_SPEC := api/openapi.yaml
OPENAPI_GENERATED_DIR := pkg/sdk/v1beta/generated
OPENAPI_GENERATED_FILE := $(OPENAPI_GENERATED_DIR)/openapi.gen.go

CLI_BINARY_NAME   := chainnet-cli
MINER_BINARY_NAME := chainnet-miner
NESPV_BINARY_NAME := chainnet-nespv
NODE_BINARY_NAME  := chainnet-node
BOT_BINARY_NAME   := chainnet-bot

# Define the source file for the CLI application
CLI_SOURCE   := $(wildcard cmd/cli/*.go)
MINER_SOURCE := $(wildcard cmd/miner/*.go)
NESPV_SOURCE := $(wildcard cmd/nespv/*go)
NODE_SOURCE  := $(wildcard cmd/node/*.go)
BOT_SOURCE   := $(wildcard cmd/bot/*.go)

# Define the source files for other files
NODE_PROTOBUF_SOURCE    := $(wildcard $(NODE_PROTOBUF_DIR)/*.proto)
NODE_PROTOBUF_PB_SOURCE := $(wildcard $(NODE_PROTOBUF_DIR)/*.pb.go)

# Define build flags
GCFLAGS := -gcflags "all=-N -l"

# Docker image names and paths
DOCKER_IMAGE_MINER := yagoninja/chainnet-miner:latest
DOCKER_IMAGE_NODE  := yagoninja/chainnet-node:latest
DOCKERFILE_MINER   := ./build/docker/miner/Dockerfile
DOCKERFILE_NODE    := ./build/docker/node/Dockerfile

.PHONY: all
all: test lint miner node nespv cli bot

.PHONY: miner
miner: protobuf openapi-generate output-dir
	@echo "Building chainnet miner..."
	@go build $(GCFLAGS) -o $(OUTPUT_DIR)/$(MINER_BINARY_NAME) $(MINER_SOURCE)

.PHONY: node
node: protobuf openapi-generate output-dir
	@echo "Building chainnet node..."
	@go build $(GCFLAGS) -o $(OUTPUT_DIR)/$(NODE_BINARY_NAME) $(NODE_SOURCE)

.PHONY: nespv
nespv: protobuf openapi-generate output-dir
	@echo "Building chainnet nespv..."
	@go build $(GCFLAGS) -o $(OUTPUT_DIR)/$(NESPV_BINARY_NAME) $(NESPV_SOURCE)

.PHONY: cli 
cli: protobuf openapi-generate output-dir
	@echo "Building chainnet CLI..."
	@go build $(GCFLAGS) -o $(OUTPUT_DIR)/$(CLI_BINARY_NAME) $(CLI_SOURCE)

.PHONY: bot
bot: protobuf openapi-generate output-dir
	@echo "Building chainnet bot..."
	@go build $(GCFLAGS) -o $(OUTPUT_DIR)/$(BOT_BINARY_NAME) $(BOT_SOURCE)

.PHONY: protobuf
protobuf:
	@echo "Generating protobuf files..."
	@protoc --go_out=. --go_opt=paths=source_relative $(NODE_PROTOBUF_SOURCE)

.PHONY: openapi-check
openapi-check:
	@echo "Checking OpenAPI spec..."
	@tmp_file="$$(mktemp)"; \
	oapi-codegen -generate types -package generated -o "$$tmp_file" $(OPENAPI_SPEC); \
	rm -f "$$tmp_file"; \
	echo "OpenAPI spec is valid: $(OPENAPI_SPEC)"

.PHONY: openapi-generate
openapi-generate:
	@echo "Generating OpenAPI SDK code..."
	@mkdir -p $(OPENAPI_GENERATED_DIR)
	@oapi-codegen -generate types,client -package generated -o $(OPENAPI_GENERATED_FILE) $(OPENAPI_SPEC)

.PHONY: output-dir
output-dir:
	@mkdir -p $(OUTPUT_DIR)

.PHONY: lint
lint: protobuf openapi-generate
	@echo "Running linter..."
	@golangci-lint run ./...

.PHONY: test
test: protobuf openapi-generate
	@echo "Running tests..."
	@go test -v -cover ./... -tags '!e2e'

.PHONY: e2e
e2e: protobuf openapi-generate
	@echo "Running e2e tests..."
	@go test -v ./tests/e2e -tags e2e

.PHONY: clean
clean:
	@echo "Cleaning up..."
	@rm -rf $(OUTPUT_DIR)
	@rm -f __debug_bin*
	@rm -f _fixture/*
	@rm -f $(NODE_PROTOBUF_PB_SOURCE)
	@rm -f $(OPENAPI_GENERATED_FILE)

.PHONY: imports
imports: 
	@find . -name "*.go" | xargs goimports -w

.PHONY: fmt
fmt:
	@go fmt ./...

.PHONY: debug
debug: node
	dlv --listen=:2345 --headless=true --api-version=2 --accept-multiclient exec ./bin/chainnet-node -- --config default-config.yaml

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

.PHONY: push
push: miner-image node-image
	@echo "Pushing Docker images to Docker Hub..."
	docker push $(DOCKER_IMAGE_MINER)
	docker push $(DOCKER_IMAGE_NODE)
