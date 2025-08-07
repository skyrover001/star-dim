# WebShell 服务器启动指南

WebShell 服务器现在支持通过命令行参数和环境变量来指定监听的IP地址和端口。

## 🚀 启动方式

### 1. 命令行参数

```bash
# 使用默认配置 (0.0.0.0:8080)
./server

# 指定IP和端口
./server -host 192.168.1.100 -port 9090

# 只监听本地连接
./server -host 127.0.0.1 -port 8888

# 显示帮助信息
./server -help
```

### 2. 环境变量

```bash
# Linux/macOS
export WEBSHELL_HOST=192.168.1.100
export WEBSHELL_PORT=9090
./server

# Windows
set WEBSHELL_HOST=192.168.1.100
set WEBSHELL_PORT=9090
server.exe
```

### 3. 使用 Makefile

```bash
# 默认启动 (0.0.0.0:8080)
make run

# 指定IP和端口
make run HOST=192.168.1.100 PORT=9090

# 预设配置
make run-local    # 127.0.0.1:8080
make run-9090     # 0.0.0.0:9090
make run-prod     # 0.0.0.0:80

# 开发模式
make dev HOST=127.0.0.1 PORT=8888
```

### 4. 使用启动脚本

#### Linux/macOS:
```bash
# 默认启动
./start.sh

# 使用环境变量
WEBSHELL_HOST=192.168.1.100 WEBSHELL_PORT=9090 ./start.sh
```

#### Windows:
```cmd
REM 默认启动
start.bat

REM 使用环境变量
set WEBSHELL_HOST=192.168.1.100
set WEBSHELL_PORT=9090
start.bat
```

## 📁 配置文件

### 环境变量配置文件 (.env)

复制 `.env.example` 为 `.env` 并根据需要修改：

```bash
cp .env.example .env
```

编辑 `.env` 文件：
```
WEBSHELL_HOST=0.0.0.0
WEBSHELL_PORT=8080
GIN_MODE=release
```

## 🌐 访问服务

启动服务器后，可以通过以下方式访问：

- **Web界面**: `http://your-host:your-port`
- **API文档**: `http://your-host:your-port/swagger/index.html`
- **登录API**: `http://your-host:your-port/api/v2/login/`

## 🧪 测试配置

更新测试配置文件以匹配服务器地址：

```javascript
// test/login_config.js
const loginConfig = {
    server: {
        baseURL: 'http://192.168.1.100:9090'  // 更新为实际服务器地址
    },
    // ... 其他配置
};
```

## 🐳 Docker 支持

如果使用 Docker，可以通过端口映射来配置：

```bash
# 构建镜像
make docker-build

# 运行容器，映射端口
docker run -p 9090:8080 webshell-server
```

## 📋 命令行选项

| 参数 | 环境变量 | 默认值 | 说明 |
|------|----------|--------|------|
| `-host` | `WEBSHELL_HOST` | `0.0.0.0` | 服务器监听地址 |
| `-port` | `WEBSHELL_PORT` | `8080` | 服务器监听端口 |
| `-help` | - | `false` | 显示帮助信息 |

## 🔧 常见配置

### 开发环境
```bash
./server -host 127.0.0.1 -port 8080
```

### 生产环境
```bash
./server -host 0.0.0.0 -port 80
```

### 内网服务
```bash
./server -host 192.168.1.100 -port 9090
```

### 高可用部署
可以使用不同端口启动多个实例：
```bash
./server -host 0.0.0.0 -port 8080 &
./server -host 0.0.0.0 -port 8081 &
./server -host 0.0.0.0 -port 8082 &
```

## 🛠️ 构建和部署

```bash
# 编译
go build -o server server.go

# 交叉编译
make build-linux    # Linux
make build-windows  # Windows

# 清理
make clean
```

## 🔍 故障排除

1. **端口被占用**: 更换端口或停止占用进程
2. **权限问题**: 使用1024以上端口或以管理员身份运行
3. **防火墙**: 确保防火墙允许相应端口的连接
4. **网络访问**: 检查IP地址和网络配置


## curl 方式接口测试

### test login API
request: curl -X POST -H "Content-Type: application/json" -d "{\"cluster\":\"hpc1\",\"username\":\"root\",\"password\":\"password\",\"host\":\"192.168.1.2\"}" "http://localhost:8080/api/v2/login/"
{"sessionKey":"tsh_a2e932b625c0d598db3800aa91b92016"}
response: {"home_path":"/root","session_key":"tsh_a2e932b625c0d598db3800aa91b92016"}
operation: save session key to local storage
mention: login API must be tested first to get session key for the following APIs and if service is restarted, session key will be invalid.

### test list API
request: curl -X GET -H "Content-Type: application/json" -H "sessionKey: tsh_a2e932b625c0d598db3800aa91b92016" "http://localhost:8080/api/v2/files/?cluster=hpc1&path=/ai"
response: {"listContent":[{"name":"mcp","size":4096,"mode":"drwxr-xr-x","modify":"2025-08-06 20:13:21 +0800 CST","isDir":true},{"name":"star-fire","size":4096,"mode":"drwxr-xr-x","modify":"2025-08-01 09:49:00 +0800 CST","isDir":true}],"listLength":2,"success":"yes"}

## test create API
### folder
request: curl -X POST -H "Content-Type: application/json" -H "sessionKey: tsh_a2e932b625c0d598db3800aa91b92016" -d "{\"path\":\"/ai/new_folder\",\"type\":\"dir\"}" "http://localhost:8080/api/v2/files/"
response: {"success":"yes"}
operation: verify new_folder created in /ai directory
### file
request: curl -X POST -H "Content-Type: application/json" -H "sessionKey: tsh_a2e932b625c0d598db3800aa91b92016" -d "{\"path\":\"/ai/new_folder/test.sh\",\"type\":\"file\"}" "http://localhost:8080/api/v2/files/"
response: {"success":"yes"}
operation: verify new_folder created in /ai directory

## test delete API
### file
request: curl -X DELETE -H "Content-Type: application/json" -H "sessionKey: tsh_a2e932b625c0d598db3800aa91b92016" -d "{\"path\":\"/ai/new_folder/test.sh\",\"type\":\"file\"}" "http://localhost:8080/api/v2/files/"
response: {"success":"yes"}
### folder
request: curl -X DELETE -H "Content-Type: application/json" -H "sessionKey: tsh_a2e932b625c0d598db3800aa91b92016" -d "{\"path\":\"/ai/new_folder\",\"type\":\"dir\"}" "http://localhost:8080/api/v2/files/"
response: {"success":"yes"}

## test rename API
create operation: create a folder named new_folder and create a file named test.sh in /ai directory
curl -X POST -H "Content-Type: application/json" -H "sessionKey: tsh_a2e932b625c0d598db3800aa91b92016" -d "{\"path\":\"/ai/new_folder\",\"type\":\"dir\"}" "http://localhost:8080/api/v2/files/"
curl -X POST -H "Content-Type: application/json" -H "sessionKey: tsh_a2e932b625c0d598db3800aa91b92016" -d "{\"path\":\"/ai/new_folder/test.sh\",\"type\":\"file\"}" "http://localhost:8080/api/v2/files/"
### folder
curl -X PUT -H "Content-Type: application/json" -H "sessionKey: tsh_a2e932b625c0d598db3800aa91b92016" -d "{\"old_path\":\"/ai/new_folder\",\"new_path\":\"/ai/new_folder_rename\"}" "http://localhost:8080/api/v2/files/"
response: {"success":"yes"}
### file
request: curl -X PUT -H "Content-Type: application/json" -H "sessionKey: tsh_a2e932b625c0d598db3800aa91b92016" -d "{\"oldPath\":\"/ai/new_folder_rename/test.sh\",\"newPath\":\"/ai/new_folder_rename/test_renamed.sh\"}" "http://localhost:8080/api/v2/files/"
response: {"success":"yes"}

## test write content API
request: curl -X POST -H "Content-Type: application/json" -H "sessionKey: tsh_a2e932b625c0d598db3800aa91b92016" -d "{\"cluster\":\"hpc1\",\"path\":\"/ai/new_folder_rename/test_renamed.sh\",\"content\":\"#!/bin/bash\\necho Hello World\"}" "http://localhost:8080/api/v2/files/content/"
response: {"content":"","success":"yes"}
## test read content API
request: curl -X GET -H "Content-Type: application/json" -H "sessionKey: tsh_a2e932b625c0d598db3800aa91b92016" "http://localhost:8080/api/v2/files/content/?cluster=hpc1&path=/ai/new_folder_rename/test_renamed.sh"
#!/bin/bash\necho Hello World

## test copy API
### file
request: curl -X POST -H "Content-Type: application/json" -H "sessionKey: tsh_a2e932b625c0d598db3800aa91b92016" -d "{\"src_path\":\"/ai/new_folder_rename/test_renamed.sh\",\"dst_path\":\"/ai/new_folder_rename/test_copy.sh\"}" "http://localhost:8080/api/v2/files/copy/" 
response: {"success":"yes"}
### folder
request: curl -X POST -H "Content-Type: application/json" -H "sessionKey: tsh_a2e932b625c0d598db3800aa91b92016" -d "{\"src_path\":\"/ai/new_folder_rename\",\"dst_path\":\"/ai/new_folder_copy\"}" "http://localhost:8080/api/v2/files/copy/"
response: {"success":"yes"}

## test move API
### file
request: curl -X POST -H "Content-Type: application/json" -H "sessionKey: tsh_a2e932b625c0d598db3800aa91b92016" -d "{\"src_path\":\"/ai/new_folder_rename/test_copy.sh\",\"dst_path\":\"/ai/new_folder_rename/test_moved.sh\"}" "http://localhost:8080/api/v2/files/move/"
response: {"success":"yes"}
### folder
request: curl -X POST -H "Content-Type: application/json" -H "sessionKey: tsh_a2e932b625c0d598db3800aa91b92016" -d "{\"src_path\":\"/ai/new_folder_copy\",\"dst_path\":\"/ai/new_folder_moved\"}" "http://localhost:8080/api/v2/files/move/"
response: {"success":"yes"}

## test attribute API
request: curl -X GET -H "Content-Type: application/json" -H "sessionKey: tsh_a2e932b625c0d598db3800aa91b92016" "http://localhost:8080/api/v2/files/attr/?cluster=hpc1&path=/ai/new_folder_rename/test_renamed.sh"
response: {"name":"test_renamed.sh","size":29,"mode":"-rw-r--r--","modify":"2025-08-07 10:16:36 +0800 CST","isDir":false}

## test change mode API
request: curl -X POST -H "Content-Type: application/json" -H "sessionKey: tsh_a2e932b625c0d598db3800aa91b92016" -d "{\"path\":\"/ai/new_folder_rename/test_renamed.sh\",\"mode\":\"755\"}" "http://localhost:8080/api/v2/files/chmod/"
response: {"success":"yes"}

## test download API
### folder
curl -X GET -H "Content-Type: application/json" -H "sessionKey: tsh_a2e932b625c0d598db3800aa91b92016" "http://localhost:8080/api/v2/files/download/?path=/ai/mcp" -o mcp.zip
### file 
curl -X GET -H "Content-Type: application/json" -H "sessionKey: tsh_a2e932b625c0d598db3800aa91b92016" "http://localhost:8080/api/v2/files/download/?path=/ai/new_folder_rename/test_renamed.sh" -o test_renamed.sh