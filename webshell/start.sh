#!/bin/bash

# WebShell 服务器启动脚本

# 默认配置
DEFAULT_HOST="0.0.0.0"
DEFAULT_PORT="8080"

# 配置选项
HOST=${WEBSHELL_HOST:-$DEFAULT_HOST}
PORT=${WEBSHELL_PORT:-$DEFAULT_PORT}

echo "=== WebShell 服务器启动脚本 ==="
echo "主机地址: $HOST"
echo "端口: $PORT"
echo "================================"

# 检查是否已编译
if [ ! -f "./server" ] && [ ! -f "./server.exe" ]; then
    echo "正在编译服务器..."
    go build -o server server.go
    if [ $? -ne 0 ]; then
        echo "编译失败!"
        exit 1
    fi
    echo "编译成功!"
fi

# 启动服务器
echo "启动服务器..."
if [ -f "./server.exe" ]; then
    ./server.exe -host $HOST -port $PORT
else
    ./server -host $HOST -port $PORT
fi
