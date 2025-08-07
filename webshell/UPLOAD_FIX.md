# 文件上传 multipart/form-data 错误修复

## 问题描述

在文件上传时遇到错误：
```
upload file!!!!!!!!!!!,error: unsupported content type: multipart/form-data; boundary=--------------------------93dfe9379c5f64851bc4d1f5
```

## 问题原因

在 `controller/public.go` 的 `GetKeyFromRequest` 函数中，对 `Content-Type` 的处理使用了精确字符串匹配：

```go
// 原来的错误代码
switch contentType {
case "application/json":
    // ...
case "application/x-www-form-urlencoded", "multipart/form-data":
    // ...
}
```

但是 `multipart/form-data` 的完整 Content-Type 包含 boundary 信息，例如：
```
multipart/form-data; boundary=--------------------------93dfe9379c5f64851bc4d1f5
```

因此精确匹配 `"multipart/form-data"` 会失败。

## 解决方案

修改 `controller/public.go` 文件中的内容类型检查逻辑：

### 1. 添加 `strings` 包导入

```go
import (
    // ... 其他导入
    "strings"
)
```

### 2. 修改 switch 语句为条件判断

```go
// 修复后的代码
switch {
case contentType == "application/json":
    if err := c.ShouldBindJSON(&requestInfo); err != nil {
        return "", nil, fmt.Errorf("failed to bind JSON: %v", err)
    }
    // 处理 JSON 参数...
    
case contentType == "application/x-www-form-urlencoded" || strings.HasPrefix(contentType, "multipart/form-data"):
    cluster, _ := c.GetPostForm("cluster")
    systemUsername, _ := c.GetPostForm("systemUsername")
    // 处理表单参数...
    
    // 对于文件上传，也需要获取其他表单字段
    if strings.HasPrefix(contentType, "multipart/form-data") {
        requestInfo.Path, _ = c.GetPostForm("path")
        requestInfo.Type, _ = c.GetPostForm("type")
    }
    
default:
    return "", nil, fmt.Errorf("unsupported content type: %s", contentType)
}
```

## 修复详情

1. **使用 `strings.HasPrefix`**: 检查 Content-Type 是否以 `"multipart/form-data"` 开头，而不是精确匹配
2. **增强表单字段处理**: 为 multipart/form-data 请求额外获取 `path` 和 `type` 字段
3. **保持向后兼容**: 原有的 `application/x-www-form-urlencoded` 处理逻辑保持不变

## 测试验证

创建了专门的文件上传测试文件 `upload_test.js`，包含：

- 简单文件上传测试
- 大文件上传测试
- 文件覆盖测试
- 断点续传测试
- 错误处理测试

运行测试：
```bash
# 简单上传测试
npm run test:upload:simple

# 完整上传测试
npm run test:upload
```

## 相关文件

修改的文件：
- `controller/public.go` - 修复 Content-Type 处理逻辑

新增的测试文件：
- `test/upload_test.js` - 专门的文件上传测试
- 更新了 `test/package.json` 和 `test/README.md`

## 注意事项

1. 这个修复是向后兼容的，不会影响现有的 API 调用
2. 修复后支持所有标准的 multipart/form-data 请求，无论 boundary 值是什么
3. 建议在生产环境部署前先运行上传测试验证功能正常
