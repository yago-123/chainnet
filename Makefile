# Define directories
OUTPUT_DIR := bin
NODE_PROTOBUF_DIR := pkg/network/protobuf
OPENAPI_SPEC := api/openapi.yaml

CLI_BINARY_NAME   := chainnet-cli
MINER_BINARY_NAME := chainnet-miner
NESPV_BINARY_NAME := chainnet-nespv
NODE_BINARY_NAME  := chainnet-node
BOT_BINARY_NAME   := chainnet-bot
COMPONENTS        := miner node nespv cli bot

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
DOCKER            ?= docker
IMAGE_REPOSITORY  ?= ghcr.io/yago-123
IMAGE_TAG         ?= latest
DOCKERFILE        := ./build/docker/Dockerfile
DOCKER_IMAGE_MINER := $(IMAGE_REPOSITORY)/chainnet-miner:$(IMAGE_TAG)
DOCKER_IMAGE_NODE  := $(IMAGE_REPOSITORY)/chainnet-node:$(IMAGE_TAG)
DOCKER_IMAGE_NESPV := $(IMAGE_REPOSITORY)/chainnet-nespv:$(IMAGE_TAG)
DOCKER_IMAGE_CLI   := $(IMAGE_REPOSITORY)/chainnet-cli:$(IMAGE_TAG)
DOCKER_IMAGE_BOT   := $(IMAGE_REPOSITORY)/chainnet-bot:$(IMAGE_TAG)

# Container build settings
IMAGE_BUILD_GOOS   ?= linux
IMAGE_BUILD_GOARCH ?= amd64
IMAGE_BUILD_FLAGS  ?= -trimpath -ldflags "-s -w"

.PHONY: all
all: test lint miner node nespv cli bot

.PHONY: miner
miner: protobuf output-dir
	@echo "Building chainnet miner..."
	@go build $(GCFLAGS) -o $(OUTPUT_DIR)/$(MINER_BINARY_NAME) $(MINER_SOURCE)

.PHONY: node
node: protobuf output-dir
	@echo "Building chainnet node..."
	@go build $(GCFLAGS) -o $(OUTPUT_DIR)/$(NODE_BINARY_NAME) $(NODE_SOURCE)

.PHONY: nespv
nespv: protobuf output-dir
	@echo "Building chainnet nespv..."
	@go build $(GCFLAGS) -o $(OUTPUT_DIR)/$(NESPV_BINARY_NAME) $(NESPV_SOURCE)

.PHONY: cli 
cli: protobuf output-dir
	@echo "Building chainnet CLI..."
	@go build $(GCFLAGS) -o $(OUTPUT_DIR)/$(CLI_BINARY_NAME) $(CLI_SOURCE)

.PHONY: bot
bot: protobuf output-dir
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

.PHONY: output-dir
output-dir:
	@mkdir -p $(OUTPUT_DIR)

.PHONY: lint
lint: protobuf
	@echo "Running linter..."
	@golangci-lint run ./...

.PHONY: test
test: protobuf
	@echo "Running tests..."
	@go test -v -cover ./... -tags '!e2e'

.PHONY: e2e
e2e: protobuf
	@echo "Running e2e tests..."
	@go test -v ./tests/e2e -tags e2e

.PHONY: clean
clean:
	@echo "Cleaning up..."
	@rm -rf $(OUTPUT_DIR)
	@rm -f __debug_bin*
	@rm -f _fixture/*
	@rm -f $(NODE_PROTOBUF_PB_SOURCE)

.PHONY: imports
imports: 
	@find . -name "*.go" | xargs goimports -w

.PHONY: fmt
fmt:
	@go fmt ./...

.PHONY: debug
debug: node
	dlv --listen=:2345 --headless=true --api-version=2 --accept-multiclient exec ./bin/chainnet-node -- --config default-config.yaml

.PHONY: image-binaries
image-binaries: $(addprefix image-binary-,$(COMPONENTS))

.PHONY: image-binary-%
image-binary-%: protobuf output-dir
	@echo "Building chainnet $* binary for container image..."
	@CGO_ENABLED=0 GOOS=$(IMAGE_BUILD_GOOS) GOARCH=$(IMAGE_BUILD_GOARCH) go build $(IMAGE_BUILD_FLAGS) -o $(OUTPUT_DIR)/chainnet-$* ./cmd/$*

.PHONY: %-image
%-image: image-binary-%
	@echo "Building Docker image for chainnet $*..."
	$(DOCKER) build --build-arg COMPONENT=$* -t $(IMAGE_REPOSITORY)/chainnet-$*:$(IMAGE_TAG) -f $(DOCKERFILE) .

.PHONY: images
images: $(addsuffix -image,$(COMPONENTS))

.PHONY: push
push: images
	@echo "Pushing container images to GitHub Container Registry..."
	@for component in $(COMPONENTS); do \
		$(DOCKER) push $(IMAGE_REPOSITORY)/chainnet-$$component:$(IMAGE_TAG); \
	done
