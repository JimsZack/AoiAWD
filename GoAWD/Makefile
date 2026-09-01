# GoAWD Makefile

# 项目信息
APP_NAME := goawd
VERSION := 1.0.0
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')

# Go 参数
GO := go
GOFLAGS := -v
LDFLAGS := -ldflags "-X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME)"

# 目录
BIN_DIR := bin
CMD_DIR := cmd

# 目标
.PHONY: all build clean test lint fmt vet run docker

all: build

# 编译所有组件
build: build-server build-roundworm build-guardian

# 编译服务器
build-server:
	@echo "Building goawd-server..."
	@mkdir -p $(BIN_DIR)
	$(GO) build $(GOFLAGS) $(LDFLAGS) -o $(BIN_DIR)/goawd-server $(CMD_DIR)/server

# 编译 RoundWorm
build-roundworm:
	@echo "Building goawd-roundworm..."
	@mkdir -p $(BIN_DIR)
	CGO_ENABLED=0 $(GO) build $(GOFLAGS) $(LDFLAGS) -o $(BIN_DIR)/goawd-roundworm $(CMD_DIR)/roundworm

# 编译 Guardian
build-guardian:
	@echo "Building goawd-guardian..."
	@mkdir -p $(BIN_DIR)
	CGO_ENABLED=0 $(GO) build $(GOFLAGS) $(LDFLAGS) -o $(BIN_DIR)/goawd-guardian $(CMD_DIR)/guardian

# 交叉编译
build-linux-amd64:
	@echo "Building for Linux amd64..."
	@mkdir -p $(BIN_DIR)
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 $(GO) build $(GOFLAGS) $(LDFLAGS) -o $(BIN_DIR)/goawd-server-linux-amd64 $(CMD_DIR)/server
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 $(GO) build $(GOFLAGS) $(LDFLAGS) -o $(BIN_DIR)/goawd-roundworm-linux-amd64 $(CMD_DIR)/roundworm
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 $(GO) build $(GOFLAGS) $(LDFLAGS) -o $(BIN_DIR)/goawd-guardian-linux-amd64 $(CMD_DIR)/guardian

build-linux-arm64:
	@echo "Building for Linux arm64..."
	@mkdir -p $(BIN_DIR)
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 $(GO) build $(GOFLAGS) $(LDFLAGS) -o $(BIN_DIR)/goawd-server-linux-arm64 $(CMD_DIR)/server
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 $(GO) build $(GOFLAGS) $(LDFLAGS) -o $(BIN_DIR)/goawd-roundworm-linux-arm64 $(CMD_DIR)/roundworm
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 $(GO) build $(GOFLAGS) $(LDFLAGS) -o $(BIN_DIR)/goawd-guardian-linux-arm64 $(CMD_DIR)/guardian

# 清理
clean:
	@echo "Cleaning..."
	@rm -rf $(BIN_DIR)
	$(GO) clean

# 测试
test:
	@echo "Running tests..."
	$(GO) test -v ./...

# 测试覆盖率
test-coverage:
	@echo "Running tests with coverage..."
	$(GO) test -v -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html

# 格式化
fmt:
	@echo "Formatting..."
	$(GO) fmt ./...

# 静态分析
vet:
	@echo "Running go vet..."
	$(GO) vet ./...

# Lint
lint:
	@echo "Running linter..."
	golangci-lint run

# 依赖管理
deps:
	@echo "Downloading dependencies..."
	$(GO) mod download
	$(GO) mod tidy

# 运行服务器
run: build-server
	@echo "Starting GoAWD server..."
	$(BIN_DIR)/goawd-server \
		-http "0.0.0.0:1337" \
		-tcp "0.0.0.0:8023" \
		-storage memory

# Docker 构建
docker:
	@echo "Building Docker image..."
	docker build -t $(APP_NAME):$(VERSION) .
	docker tag $(APP_NAME):$(VERSION) $(APP_NAME):latest

# Docker 运行
docker-run:
	@echo "Running Docker containers..."
	docker-compose up -d

# Docker 停止
docker-stop:
	@echo "Stopping Docker containers..."
	docker-compose down

# 帮助
help:
	@echo "GoAWD Makefile"
	@echo ""
	@echo "Usage:"
	@echo "  make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  all                - Build all components (default)"
	@echo "  build              - Build all components"
	@echo "  build-server       - Build GoAWD server"
	@echo "  build-roundworm    - Build RoundWorm probe"
	@echo "  build-guardian     - Build Guardian probe"
	@echo "  build-linux-amd64  - Cross compile for Linux amd64"
	@echo "  build-linux-arm64  - Cross compile for Linux arm64"
	@echo "  clean              - Remove build artifacts"
	@echo "  test               - Run tests"
	@echo "  test-coverage      - Run tests with coverage"
	@echo "  fmt                - Format code"
	@echo "  vet                - Run go vet"
	@echo "  lint               - Run linter"
	@echo "  deps               - Download dependencies"
	@echo "  run                - Build and run server"
	@echo "  docker             - Build Docker image"
	@echo "  docker-run         - Run Docker containers"
	@echo "  docker-stop        - Stop Docker containers"
	@echo "  help               - Show this help"
