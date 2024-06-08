all: build

.PHONY: build
build:
	go build main.go

.PHONY: run
run:
	go run main.go

.PHONY: debug
debug: 
	dlv debug --headless --listen=:2345 --api-version=2 --accept-multiclient	

