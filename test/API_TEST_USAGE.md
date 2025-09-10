# API 测试工具使用指南

## 概述

`all_test.js` 是一个功能完整的 API 测试工具，支持测试 WebShell 项目的各种文件操作接口。现在支持通过配置文件管理登录参数，与 `login_test.js` 保持一致。

## 功能特性

- ✅ 支持测试单个或多个接口
- ✅ 自动处理登录/登出逻辑
- ✅ 提供详细的测试输出
- ✅ **配置文件支持** - 从 `login_config.js` 读取配置
- ✅ 命令行参数支持
- ✅ 智能路径管理

## 配置文件

### 1. 配置文件设置

工具会自动尝试加载 `login_config.js` 文件：

```bash
# 如果存在 login_config.js
✅ 使用配置文件: login_config.js

# 如果不存在，会使用示例配置并提示
⚠️  未找到 login_config.js 文件，使用示例配置
💡 请复制 login_config.example.js 为 login_config.js 并填入真实的登录信息
```

### 2. 配置文件格式

```javascript
const loginConfig = {
    // 服务器配置
    server: {
        baseURL: 'http://localhost:8080',  // WebShell 服务器地址
    },
    
    // SSH连接配置
    ssh: {
        cluster: 'hpc1',           // 集群名称
        username: 'root',          // 用户名
        password: 'your_password', // 密码
        host: '1.94.239.51',       // SSH服务器地址
        port: '22'                 // SSH端口
    },
    
    // 测试路径配置
    testPaths: {
        homeDir: '/home',              // 主目录路径
        testDir: '/home/test_api'      // 测试目录路径
    },
    
    // 测试配置
    test: {
        timeout: 5000,              // 请求超时时间 (毫秒)
        retries: 3,                 // 失败重试次数
        cleanupOnExit: true         // 测试结束后是否清理测试文件
    }
};

module.exports = loginConfig;
```

## 使用方法

### 1. 显示帮助信息
```bash
node all_test.js --help
```

### 2. 完整测试 (默认)
```bash
node all_test.js
```
运行默认的完整测试，包含登录、上传文件、登出。

### 3. 仅登录测试
```bash
node all_test.js --login-only
```
只测试登录、列出文件、登出功能。

### 4. 测试单个接口
```bash
# 测试登录接口
node all_test.js --test login

# 测试文件列表接口
node all_test.js --test list

# 测试文件上传接口
node all_test.js --test upload
```

### 5. 测试多个接口
```bash
# 测试登录、列出文件、上传、登出
node all_test.js --test login,list,upload,logout

# 测试文件操作相关接口
node all_test.js --test createDir,createFile,writeContent,readContent
```

## 可用的测试接口

| 接口名称 | 功能描述 | 备注 |
|---------|---------|------|
| `login` | 用户登录 | 自动获取会话密钥 |
| `logout` | 用户登出 | 清除会话密钥 |
| `list` / `listFiles` | 列出目录文件 | 两个名称等效 |
| `createDir` | 创建目录 | - |
| `createFile` | 创建文件 | - |
| `writeContent` | 写入文件内容 | - |
| `readContent` | 读取文件内容 | - |
| `getAttr` | 获取文件属性 | - |
| `chmod` | 修改文件权限 | - |
| `copy` | 复制文件 | - |
| `rename` | 重命名文件 | - |
| `move` | 移动文件 | - |
| `delete` | 删除文件 | - |
| `upload` | 上传文件 | 需要本地文件 `pic.png` |
| `download` | 下载文件 | - |
| `execute` | 执行脚本文件 | - |
| `chown` | 修改文件所有者 | - |
| `quota` | 获取配额信息 | - |

| `execute` | 执行脚本文件 | - |
| `chown` | 修改文件所有者 | - |
| `quota` | 获取配额信息 | - |

## 配置管理

### 自动配置加载

工具现在支持自动从配置文件加载参数，与 `login_test.js` 保持一致的配置格式：

1. **优先使用** `login_config.js` 文件（如果存在）
2. **备用** `login_config.example.js` 文件（如果主配置文件不存在）

### 配置项说明

- `server.baseURL`: WebShell 服务器地址
- `ssh.cluster`: 集群名称
- `ssh.username`: SSH 用户名
- `ssh.password`: SSH 密码
- `ssh.host`: SSH 服务器地址
- `ssh.port`: SSH 端口
- `testPaths.homeDir`: 主目录路径（用于列出文件测试）
- `testPaths.testDir`: 测试目录路径（用于文件操作测试）

### 创建配置文件

```bash
# 复制示例配置文件
cp login_config.example.js login_config.js

# 编辑配置文件
# 修改其中的用户名、密码、主机地址等信息
```

## 智能登录管理

- 如果测试不包含 `login`，工具会自动先执行登录
- 如果测试不包含 `logout`，工具会在测试结束后自动登出
- 避免重复登录和登出操作

## 示例用法

### 测试文件上传功能
```bash
node all_test.js --test upload
```

### 测试完整的文件操作流程
```bash
node all_test.js --test login,createDir,createFile,writeContent,readContent,upload,download,logout
```

### 测试权限相关操作
```bash
node all_test.js --test chmod,chown,getAttr
```

## 注意事项

1. 确保服务器正在运行并且可以访问
2. 上传测试需要当前目录下有 `pic.png` 文件
3. 某些测试可能需要特定的文件或目录存在
4. 修改 `DEFAULT_CONFIG` 以匹配你的测试环境

## 错误处理

工具会捕获并显示详细的错误信息，包括：
- HTTP 响应错误
- 网络连接错误
- 参数验证错误

每个接口测试都有独立的错误处理，不会因为单个接口失败而中断整个测试流程。
