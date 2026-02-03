.PHONY: build run run-sse inspect clean test tidy

BINARY_NAME=mcp-repo-monitor
BUILD_DIR=bin
SSE_PORT=8080

build:
	@echo "Building..."
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/server

run: build
	./$(BUILD_DIR)/$(BINARY_NAME) -mode=stdio

run-sse: build
	./$(BUILD_DIR)/$(BINARY_NAME) -mode=sse -addr=:$(SSE_PORT)

inspect: build
	npx @modelcontextprotocol/inspector ./$(BUILD_DIR)/$(BINARY_NAME)

clean:
	rm -rf $(BUILD_DIR)
	go clean

test:
	go test -v ./...

tidy:
	go mod tidy

.DEFAULT_GOAL := build
