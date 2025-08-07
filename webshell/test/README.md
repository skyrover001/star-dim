# API 测试工具使用说明

这个目录包含了用于测试 webshell API 的工具，特别是登录功能和文件操作API。

## 文件说明

- `test_upload.js` - 主要的API测试类，包含所有文件操作和登录功能
- `login_test.js` - 带配置文件的登录测试脚本
- `login_config.example.js` - 登录配置文件示例
- `package.json` - Node.js 依赖配置

## 安装依赖

```bash
npm install
```

或者使用预定义脚本：

```bash
npm run install-deps
```

## 配置登录信息

1. 复制配置文件示例：
   ```bash
   copy login_config.example.js login_config.js
   ```

2. 编辑 `login_config.js` 文件，填入真实的登录信息：
   ```javascript
   const loginConfig = {
       server: {
           baseURL: 'http://localhost:8080'  // 你的服务器地址
       },
       ssh: {
           cluster: 'default',     // 集群名称
           username: 'your_user',  // SSH用户名
           password: 'your_pass',  // SSH密码
           host: '192.168.1.100',  // SSH服务器IP
           port: '22'              // SSH端口
       },
       testPaths: {
           homeDir: '/home',
           testDir: '/home/test_api'
       }
   };
   ```

## 运行测试

### 1. 连接测试（推荐先运行）
测试服务器连接和登录功能：
```bash
npm run test:connection
```

### 2. 完整测试
运行完整的API测试，包括登录和所有文件操作：
```bash
npm run test:full
```

### 3. 仅登录测试
只测试登录和登出功能：
```bash
npm run test:login
```

### 4. 文件上传测试
专门测试文件上传功能：
```bash
# 完整上传测试（包括大文件和断点续传）
npm run test:upload

# 简单上传测试
npm run test:upload:simple
```

### 5. 原始测试
运行原始的测试脚本：
```bash
npm test
```

## 直接使用FileAPITester类

```javascript
const { FileAPITester } = require('./test_upload.js');

async function customTest() {
    const tester = new FileAPITester('http://localhost:8080');
    
    // 登录
    const result = await tester.login('default', 'username', 'password', '192.168.1.100');
    console.log('Session Key:', result.session_key);
    
    // 执行文件操作
    await tester.listFiles('default', '/home');
    
    // 登出
    await tester.logout();
}
```

## API 方法列表

### 认证相关
- `login(cluster, username, password, host, port)` - 用户登录
- `logout()` - 用户登出
- `updateSessionKey(sessionKey)` - 更新会话密钥

### 文件操作
- `listFiles(cluster, path)` - 列出目录文件
- `uploadFile(cluster, path, filePath, offset, update)` - 上传文件
- `downloadFile(cluster, path, outputPath)` - 下载文件
- `getFileAttr(cluster, path)` - 获取文件属性
- `readFileContent(cluster, path)` - 读取文件内容
- `writeFileContent(cluster, path, content)` - 写入文件内容
- `executeFile(cluster, path)` - 执行脚本文件

### 文件管理
- `createFile(path, type)` - 创建文件/目录
- `deleteFile(path, type)` - 删除文件/目录
- `renameFile(oldPath, newPath)` - 重命名文件
- `copyFile(srcPath, dstPath)` - 复制文件
- `moveFile(srcPath, dstPath)` - 移动文件

### 权限管理
- `changeMode(path, mode)` - 修改文件权限
- `changeOwner(path, owner, group)` - 修改文件所有者

### 其他
- `getQuota(cluster, path)` - 获取配额信息

## 常见问题

1. **连接失败**：检查服务器地址和端口是否正确
2. **登录失败**：检查SSH用户名、密码和主机地址
3. **权限错误**：确保SSH用户有相应的文件操作权限
4. **路径错误**：使用绝对路径，确保目录存在

## 注意事项

- 测试会创建和删除文件，请在安全的环境中运行
- 确保SSH服务器允许密码认证
- 建议先运行连接测试，确认配置正确后再运行完整测试
