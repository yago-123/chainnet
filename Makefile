all: build

.PHONY: build
build:
	go build main.go

.PHONY: run
run:
	go run main.go

.PHONY: test
test:
	go test ./...

.PHONY: debug
debug: 
	dlv debug --headless --listen=:2345 --api-version=2 --accept-multiclient

.PHONY: clean
clean: 
	rm -f __debug_bin*
	rm -f _fixture/*
	rm -f main

