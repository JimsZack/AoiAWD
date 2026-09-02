# Snowfind Makefile

APP_NAME = snowfind
VERSION ?= v1.0.0
MAIN_FILE = main.go

# Go环境变量
GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)
CGO_ENABLED = 0

# 构建标志
LDFLAGS = -s -w -X main.version=$(VERSION)
BUILD_FLAGS = -ldflags="$(LDFLAGS)"

# 支持的平台
PLATFORMS = linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64 windows/386 freebsd/amd64

.PHONY: all build build-all clean test deps help release install

# 默认目标
all: deps test build

# 构建当前平台版本
build:
	@echo "构建 $(APP_NAME) $(VERSION) for $(GOOS)/$(GOARCH)"
	@CGO_ENABLED=$(CGO_ENABLED) go build $(BUILD_FLAGS) -o $(APP_NAME) $(MAIN_FILE)
	@echo "✓ 构建完成: $(APP_NAME)"

# 构建所有平台版本
build-all: clean deps
	@echo "构建所有平台版本..."
	@mkdir -p bin release
	@for platform in $(PLATFORMS); do \
		os=$${platform%/*}; \
		arch=$${platform#*/}; \
		output="bin/$(APP_NAME)-$$os-$$arch"; \
		if [ "$$os" = "windows" ]; then output="$$output.exe"; fi; \
		echo "构建 $$platform..."; \
		CGO_ENABLED=$(CGO_ENABLED) GOOS=$$os GOARCH=$$arch go build $(BUILD_FLAGS) -o $$output $(MAIN_FILE); \
		if [ $$? -eq 0 ]; then \
			echo "✓ 成功: $$output"; \
		else \
			echo "✗ 失败: $$platform"; \
			exit 1; \
		fi; \
	done
	@echo "✓ 所有平台构建完成"

# 构建发布包
release: build-all
	@echo "创建发布包..."
	@./build.sh $(VERSION)
	@echo "✓ 发布包创建完成"

# 运行测试
test:
	@echo "运行测试..."
	@go test -v ./...

# 下载依赖
deps:
	@echo "下载依赖..."
	@go mod tidy
	@go mod download

# 清理构建文件
clean:
	@echo "清理构建文件..."
	@rm -rf bin/ release/ $(APP_NAME) $(APP_NAME).exe
	@go clean
	@echo "✓ 清理完成"

# 安装到系统
install: build
	@echo "安装到系统..."
	@sudo cp $(APP_NAME) /usr/local/bin/
	@echo "✓ 安装完成"

# 开发模式运行
dev: build
	@./$(APP_NAME) .

# 显示帮助
help:
	@echo "Snowfind 构建系统"
	@echo ""
	@echo "可用命令:"
	@echo "  all         - 运行 deps, test, build (默认)"
	@echo "  build       - 构建当前平台版本"
	@echo "  build-all   - 构建所有平台版本"
	@echo "  release     - 创建发布包"
	@echo "  test        - 运行测试"
	@echo "  deps        - 下载依赖"
	@echo "  clean       - 清理构建文件"
	@echo "  install     - 安装到系统"
	@echo "  dev         - 构建并在当前目录运行"
	@echo "  help        - 显示此帮助信息"
	@echo ""
	@echo "变量:"
	@echo "  VERSION     - 版本号 (默认: $(VERSION))"
	@echo "  GOOS        - 目标操作系统 (默认: $(GOOS))"
	@echo "  GOARCH      - 目标架构 (默认: $(GOARCH))"
	@echo ""
	@echo "示例:"
	@echo "  make build VERSION=v1.2.0"
	@echo "  make build-all VERSION=v1.2.0"
	@echo "  make release VERSION=v1.2.0"
