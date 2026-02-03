.PHONY: build run run-sse inspect clean test tidy docker-build docker-run

BINARY_NAME=mcp-repo-monitor
BUILD_DIR=bin
SSE_PORT=8080
IMAGE_NAME=mcp-repo-monitor

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

docker-build:
	docker build -t $(IMAGE_NAME) .

docker-run: docker-build
	docker run -it --rm --env-file .env $(IMAGE_NAME)

.DEFAULT_GOAL := build
