all: build

.PHONY: build
build:
	go build main.go

.PHONY: run
run:
	go run main.go