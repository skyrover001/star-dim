#!/bin/bash

# SLURM JobAcct API Server 启动脚本

# 设置脚本退出时的行为
set -e

# 获取脚本所在目录
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 日志函数
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查是否安装了 Go
check_go() {
    if ! command -v go &> /dev/null; then
        log_error "Go is not installed. Please install Go 1.21 or later."
        exit 1
    fi
    
    GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
    log_info "Go version: $GO_VERSION"
}

# 安装依赖
install_deps() {
    log_info "Installing dependencies..."
    go mod tidy
    go mod download
    log_success "Dependencies installed successfully"
}

# 构建应用
build_app() {
    log_info "Building application..."
    make build
    log_success "Application built successfully"
}

# 启动服务
start_service() {
    log_info "Starting SLURM JobAcct API Server..."
    
    # 检查端口是否被占用
    PORT=${PORT:-8080}
    if lsof -Pi :$PORT -sTCP:LISTEN -t >/dev/null 2>&1; then
        log_warning "Port $PORT is already in use"
        read -p "Do you want to kill the process and continue? (y/N): " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            PID=$(lsof -ti :$PORT)
            kill -9 $PID
            log_info "Killed process $PID"
        else
            log_error "Cannot start server, port $PORT is in use"
            exit 1
        fi
    fi
    
    # 启动服务
    ./build/slurm-jobacct &
    SERVER_PID=$!
    
    # 等待服务启动
    sleep 2
    
    # 检查服务是否正常启动
    if kill -0 $SERVER_PID 2>/dev/null; then
        log_success "Server started successfully (PID: $SERVER_PID)"
        log_info "Server is running on http://localhost:$PORT"
        log_info "Health check: http://localhost:$PORT/health"
        log_info "API documentation: see README.md"
        
        # 保存 PID 到文件
        echo $SERVER_PID > slurm-jobacct.pid
        
        # 等待用户输入停止服务
        echo
        read -p "Press Enter to stop the server..." -r
        
        # 停止服务
        log_info "Stopping server..."
        kill $SERVER_PID
        rm -f slurm-jobacct.pid
        log_success "Server stopped"
    else
        log_error "Failed to start server"
        exit 1
    fi
}

# 开发模式启动
start_dev() {
    log_info "Starting in development mode..."
    make dev
}

# 显示帮助信息
show_help() {
    echo "SLURM JobAcct API Server - 启动脚本"
    echo
    echo "用法: $0 [选项]"
    echo
    echo "选项:"
    echo "  -h, --help     显示此帮助信息"
    echo "  -d, --dev      以开发模式启动（热重载）"
    echo "  -b, --build    仅构建，不启动"
    echo "  -c, --check    检查环境和依赖"
    echo
    echo "环境变量:"
    echo "  PORT           服务端口 (默认: 8080)"
    echo "  GIN_MODE       Gin 模式 (debug/release，默认: release)"
    echo
    echo "示例:"
    echo "  $0             # 正常启动"
    echo "  $0 -d          # 开发模式启动"
    echo "  PORT=9090 $0   # 在端口 9090 启动"
}

# 检查环境
check_env() {
    log_info "Checking environment..."
    
    check_go
    
    # 检查必要的文件
    if [[ ! -f "go.mod" ]]; then
        log_error "go.mod not found. Are you in the correct directory?"
        exit 1
    fi
    
    if [[ ! -f "main.go" ]]; then
        log_error "main.go not found. Are you in the correct directory?"
        exit 1
    fi
    
    log_success "Environment check passed"
}

# 主函数
main() {
    case "${1:-}" in
        -h|--help)
            show_help
            ;;
        -d|--dev)
            check_env
            install_deps
            start_dev
            ;;
        -b|--build)
            check_env
            install_deps
            build_app
            ;;
        -c|--check)
            check_env
            install_deps
            log_success "All checks passed"
            ;;
        "")
            check_env
            install_deps
            build_app
            start_service
            ;;
        *)
            log_error "Unknown option: $1"
            show_help
            exit 1
            ;;
    esac
}

# 捕获中断信号
trap 'log_info "Received interrupt signal"; exit 0' INT TERM

# 运行主函数
main "$@"
