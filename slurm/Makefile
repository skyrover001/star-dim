# Makefile for SLURM JobAcct API Server

.PHONY: build run clean test deps dev help

# 变量定义
APP_NAME = slurm-jobacct
BUILD_DIR = build
MAIN_FILE = main.go
PORT = 8080

# 默认目标
help:
	@echo "Available commands:"
	@echo "  deps     - Download and install dependencies"
	@echo "  build    - Build the application"
	@echo "  run      - Run the application"
	@echo "  dev      - Run in development mode with auto-reload"
	@echo "  test     - Run tests"
	@echo "  clean    - Clean build artifacts"
	@echo "  help     - Show this help message"

# 安装依赖
deps:
	@echo "Installing dependencies..."
	go mod tidy
	go mod download

# 构建应用
build: deps
	@echo "Building $(APP_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(APP_NAME) $(MAIN_FILE)
	@echo "Build completed: $(BUILD_DIR)/$(APP_NAME)"

# 运行应用
run: build
	@echo "Starting $(APP_NAME) on port $(PORT)..."
	./$(BUILD_DIR)/$(APP_NAME)

# 开发模式（直接运行，不构建）
dev: deps
	@echo "Starting $(APP_NAME) in development mode..."
	go run $(MAIN_FILE)

# 运行测试
test:
	@echo "Running tests..."
	go test -v ./...

# 清理构建文件
clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BUILD_DIR)
	go clean

# 检查代码格式
fmt:
	@echo "Formatting code..."
	go fmt ./...

# 代码检查
vet:
	@echo "Running go vet..."
	go vet ./...

# 安装开发工具
install-tools:
	@echo "Installing development tools..."
	go install github.com/air-verse/air@latest

# 使用 air 进行热重载开发
air-dev:
	@echo "Starting with air hot reload..."
	air

# Docker 构建
docker-build:
	@echo "Building Docker image..."
	docker build -t $(APP_NAME):latest .

# Docker 运行
docker-run:
	@echo "Running Docker container..."
	docker run -p $(PORT):$(PORT) $(APP_NAME):latest
