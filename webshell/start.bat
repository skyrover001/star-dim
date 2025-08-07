@echo off
REM WebShell 服务器启动脚本 (Windows)

REM 默认配置
set DEFAULT_HOST=0.0.0.0
set DEFAULT_PORT=8080

REM 配置选项
if "%WEBSHELL_HOST%"=="" set WEBSHELL_HOST=%DEFAULT_HOST%
if "%WEBSHELL_PORT%"=="" set WEBSHELL_PORT=%DEFAULT_PORT%

echo === WebShell 服务器启动脚本 ===
echo 主机地址: %WEBSHELL_HOST%
echo 端口: %WEBSHELL_PORT%
echo ================================

REM 检查是否已编译
if not exist "server.exe" (
    echo 正在编译服务器...
    go build -o server.exe server.go
    if errorlevel 1 (
        echo 编译失败!
        pause
        exit /b 1
    )
    echo 编译成功!
)

REM 启动服务器
echo 启动服务器...
server.exe -host %WEBSHELL_HOST% -port %WEBSHELL_PORT%

pause
