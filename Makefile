
.PHONY: all
all: proto

.PHONY: clean
clean: tools-clean

.PHONY: tools
tools: bin/buf bin/mockery bin/protoc-gen-buf-breaking bin/protoc-gen-buf-lint bin/protoc-gen-connect-go bin/protoc-gen-go

.PHONY: tools-clean
tools-clean:
	rm -rf bin
	rm tools/dependencies

tools/dependencies: tools/*
	cd tools && go mod download
	touch tools/dependencies

bin/buf: tools/dependencies
	mkdir -p bin
	cd tools && go build -o ../bin/buf github.com/bufbuild/buf/cmd/buf

bin/protoc-gen-buf-breaking: tools/dependencies
	mkdir -p bin
	cd tools && go build -o ../bin/protoc-gen-buf-breaking github.com/bufbuild/buf/cmd/protoc-gen-buf-breaking

bin/protoc-gen-buf-lint: tools/dependencies
	mkdir -p bin
	cd tools && go build -o ../bin/protoc-gen-buf-lint github.com/bufbuild/buf/cmd/protoc-gen-buf-lint

bin/protoc-gen-go: tools/dependencies
	mkdir -p bin
	cd tools && go build -o ../bin/protoc-gen-go google.golang.org/protobuf/cmd/protoc-gen-go

BIN_DIR = $(shell pwd)/bin

.PHONY: proto
proto: bin/buf bin/protoc-gen-buf-breaking bin/protoc-gen-buf-lint bin/protoc-gen-go
	PATH="$(BIN_DIR):$$PATH" buf lint
	rm -rf internal/testtypes
	PATH="$(BIN_DIR):$$PATH" buf generate
