# SLURM JobAcct & JobSubmit & JobQueue API Server

一个基于 Gin 框架的 SLURM API 服务器，提供作业记账查询、作业提交和作业队列查询功能，通过 SSH 连接到 SLURM 登录节点执行 `sacct`、`sbatch` 和 `squeue` 命令并格式化返回结果。

## 功能特性

- 🚀 RESTful API 接口
- 📊 完整的 SLURM 作业记账查询 (`sacct`)
- 🎯 完整的 SLURM 作业提交功能 (`sbatch`)
- 📋 完整的 SLURM 作业队列查询 (`squeue`)
- 🔐 SSH 连接支持（密码和私钥认证）
- 📁 脚本文件上传和管理
- ⚙️ 灵活的输出格式和参数配置
- 📝 JSON 响应格式
- 🛡️ 认证中间件
- 🌐 CORS 支持
- 📖 完整的 API 文档& JobSubmit API Server

一个基于 Gin 框架的 SLURM API 服务器，提供作业记账查询和作业提交功能，通过 SSH 连接到 SLURM 登录节点执行 `sacct` 和 `sbatch` 命令并格式化返回结果。

## 功能特性

- 🚀 RESTful API 接口
- 📊 完整的 SLURM 作业记账查询 (`sacct`)
- 🎯 完整的 SLURM 作业提交功能 (`sbatch`)
- 🔐 SSH 连接支持（密码和私钥认证）
- � 脚本文件上传和管理
- �️ 灵活的输出格式和参数配置
- 📝 JSON 响应格式
- 🛡️ 认证中间件
- 🌐 CORS 支持
- 📖 完整的 API 文档

## 快速开始

### 安装依赖

```bash
make deps
```

### 构建应用

```bash
make build
```

### 运行服务

```bash
make run
```

或直接运行（开发模式）：

```bash
make dev
```

服务默认运行在端口 8080。

## API 接口

### 基础 URL

```
http://localhost:8080/api/v1
```

### 认证

所有 API 请求需要在 Header 中包含认证信息：

```
sessionKey: your_session_key
```

或

```
Authorization: Bearer your_token
```

### 接口列表

#### 作业记账查询 API

##### 1. 获取作业列表

```http
GET /api/v1/jobacct/jobs
```

**查询参数：**

| 参数 | 类型 | 必需 | 描述 |
|------|------|------|------|
| host | string | 是 | SSH 主机地址 |
| username | string | 是 | SSH 用户名 |
| password | string | 否* | SSH 密码 |
| privatekey | string | 否* | SSH 私钥路径 |
| port | int | 否 | SSH 端口（默认 22） |
| jobids | string | 否 | 作业 ID（逗号分隔） |
| users | string | 否 | 用户名（逗号分隔） |
| accounts | string | 否 | 账户（逗号分隔） |
| partitions | string | 否 | 分区（逗号分隔） |
| states | string | 否 | 状态（逗号分隔） |
| starttime | string | 否 | 开始时间 |
| endtime | string | 否 | 结束时间 |
| format | string | 否 | 输出格式 |
| brief | boolean | 否 | 简洁模式 |
| long | boolean | 否 | 详细模式 |
| parsable | boolean | 否 | 可解析格式 |
| noheader | boolean | 否 | 无表头 |
| allusers | boolean | 否 | 所有用户 |

*注：password 和 privatekey 至少需要提供一个。

**响应示例：**

```json
{
  "success": true,
  "message": "Success",
  "data": [
    {
      "jobid": "12345",
      "jobname": "test_job",
      "partition": "compute",
      "account": "default",
      "alloccpus": 4,
      "state": "COMPLETED",
      "exitcode": "0:0",
      "start": "2024-01-15T10:00:00Z",
      "end": "2024-01-15T11:00:00Z",
      "elapsed": "01:00:00",
      "reqmem": "4Gn",
      "nodelist": "node001",
      "user": "testuser"
    }
  ],
  "total": 1,
  "command": "sacct -o jobid,jobname,partition,account,alloccpus,state,exitcode,start,end,elapsed,reqmem,nodelist,user"
}
```

##### 2. 获取单个作业详情

```http
GET /api/v1/jobacct/jobs/{jobid}
```

**路径参数：**

| 参数 | 类型 | 描述 |
|------|------|------|
| jobid | string | 作业 ID |

**查询参数：**

同获取作业列表接口的 SSH 连接参数。

##### 3. 获取用户作业

```http
GET /api/v1/jobacct/users/{user}/jobs
```

**路径参数：**

| 参数 | 类型 | 描述 |
|------|------|------|
| user | string | 用户名 |

**查询参数：**

同获取作业列表接口。

##### 4. 复杂查询接口

```http
POST /api/v1/jobacct/accounting
```

**请求体：**

```json
{
  "host": "your-slurm-host",
  "username": "your-username",
  "password": "your-password",
  "jobids": ["12345", "12346"],
  "users": ["user1", "user2"],
  "accounts": ["account1"],
  "partitions": ["compute", "gpu"],
  "states": ["COMPLETED", "FAILED"],
  "starttime": "2024-01-01",
  "endtime": "2024-01-31",
  "format": "jobid,jobname,state,elapsed",
  "long": true,
  "allusers": false
}
```

### 5. 健康检查

```http
GET /health
```

**响应：**

```json
{
  "status": "ok",
  "service": "slurm-jobacct"
}
```

#### 作业提交 API

##### 1. 快速提交作业

```http
POST /api/v1/jobsubmit/quick-submit
```

**请求体：**

```json
{
  "host": "your-slurm-host",
  "username": "your-username",
  "password": "your-password",
  "wrap": "echo 'Hello SLURM!'",
  "job_name": "test_quick_job",
  "partition": "compute",
  "time": "00:10:00",
  "ntasks": 1,
  "cpus_per_task": 1,
  "memory": "1G",
  "output": "quick_job_%j.out",
  "error": "quick_job_%j.err"
}
```

**响应示例：**

```json
{
  "success": true,
  "message": "Job submitted successfully",
  "job_id": "12346",
  "cluster": "cluster1",
  "command": "sbatch --job-name=test_quick_job --partition=compute --time=00:10:00 --ntasks=1 --cpus-per-task=1 --mem=1G --output=quick_job_%j.out --error=quick_job_%j.err --wrap='echo Hello SLURM!'"
}
```

##### 2. 完整参数提交作业

```http
POST /api/v1/jobsubmit/submit
```

**请求体：** 支持完整的 sbatch 参数，详见 API 文档。

##### 3. 通过脚本文件提交作业

```http
POST /api/v1/jobsubmit/submit-script
```

**Content-Type:** `multipart/form-data`

**表单字段：**
- `script`: 脚本文件 (file)
- `host`: SSH 主机 (string)
- `username`: SSH 用户名 (string)
- `password`: SSH 密码 (string, 可选)
- `privatekey`: SSH 私钥路径 (string, 可选)
- 其他 sbatch 参数...

##### 4. 支持的 sbatch 参数

本 API 支持完整的 sbatch 命令参数，包括：

**基本参数：**
- `job_name` (-J, --job-name): 作业名称
- `partition` (-p, --partition): 分区
- `time` (-t, --time): 时间限制
- `ntasks` (-n, --ntasks): 任务数量
- `cpus_per_task` (-c, --cpus-per-task): 每任务CPU数
- `memory` (--mem): 内存需求
- `output` (-o, --output): 标准输出文件
- `error` (-e, --error): 标准错误文件

**高级参数：**
- `array` (-a, --array): 数组作业
- `dependency` (-d, --dependency): 作业依赖
- `gres` (--gres): 通用资源
- `constraint` (-C, --constraint): 约束条件
- `exclusive` (--exclusive): 独占模式
- `mail_type` (--mail-type): 邮件通知类型
- `mail_user` (--mail-user): 邮件接收者

详细参数列表请参考 `examples/README.md`。

#### 作业队列查询 API

##### 1. 获取作业队列列表

```http
GET /api/v1/jobqueue/queue
```

**查询参数：**

| 参数 | 类型 | 必需 | 描述 |
|------|------|------|------|
| host | string | 是 | SSH 主机地址 |
| username | string | 是 | SSH 用户名 |
| password | string | 否* | SSH 密码 |
| privatekey | string | 否* | SSH 私钥路径 |
| port | int | 否 | SSH 端口（默认 22） |
| accounts | string | 否 | 账户（逗号分隔） |
| jobs | string | 否 | 作业 ID（逗号分隔） |
| partitions | string | 否 | 分区（逗号分隔） |
| qos | string | 否 | QOS（逗号分隔） |
| states | string | 否 | 状态（逗号分隔） |
| users | string | 否 | 用户名（逗号分隔） |
| names | string | 否 | 作业名称（逗号分隔） |
| clusters | string | 否 | 集群（逗号分隔） |
| format | string | 否 | 输出格式 |
| long | boolean | 否 | 详细模式 |
| noheader | boolean | 否 | 无表头 |
| start | boolean | 否 | 显示预计开始时间 |
| all | boolean | 否 | 显示隐藏分区的作业 |
| sort | string | 否 | 排序字段（逗号分隔） |

*注：password 和 privatekey 至少需要提供一个。

**响应示例：**

```json
{
  "success": true,
  "message": "Success",
  "data": [
    {
      "jobid": "12345",
      "partition": "compute",
      "name": "test_job",
      "user": "testuser",
      "state": "RUNNING",
      "time": "1:30:45",
      "time_left": "28:29:15",
      "nodes": 2,
      "nodelist": "node[001-002]",
      "priority": 1000,
      "qos": "normal",
      "account": "default"
    }
  ],
  "total": 1,
  "command": "squeue -o \"%.10i %.9P %.20j %.8u %.2t %.10M %.10l %.6D %R\""
}
```

##### 2. 获取单个作业的队列信息

```http
GET /api/v1/jobqueue/jobs/{jobid}
```

**路径参数：**

| 参数 | 类型 | 描述 |
|------|------|------|
| jobid | string | 作业 ID |

**查询参数：**

同获取作业队列列表接口的 SSH 连接参数。

##### 3. 获取用户作业队列

```http
GET /api/v1/jobqueue/users/{user}/queue
```

**路径参数：**

| 参数 | 类型 | 描述 |
|------|------|------|
| user | string | 用户名 |

**查询参数：**

同获取作业队列列表接口。

##### 4. 复杂队列查询接口

```http
POST /api/v1/jobqueue/query
```

**请求体：**

```json
{
  "host": "your-slurm-host",
  "username": "your-username",
  "password": "your-password",
  "partitions": ["compute", "gpu"],
  "states": ["RUNNING", "PENDING"],
  "users": ["user1", "user2"],
  "long": true,
  "sort": ["priority", "submittime"],
  "start": true
}
```

##### 5. 获取队列统计信息

```http
GET /api/v1/jobqueue/stats
```

**查询参数：**

同获取作业队列列表接口的 SSH 连接参数。

**响应示例：**

```json
{
  "success": true,
  "message": "Success",
  "total_jobs": 150,
  "statistics": {
    "PENDING": 45,
    "RUNNING": 78,
    "SUSPENDED": 2,
    "COMPLETING": 5,
    "COMPLETED": 0,
    "CANCELLED": 15,
    "FAILED": 3,
    "TIMEOUT": 1,
    "PREEMPTED": 0,
    "NODE_FAIL": 1,
    "OTHER": 0
  },
  "command": "squeue -t all -o \"%.10i %.2t %.9P %.8u\" -h"
}
```

## 支持的 sacct 参数

本 API 支持大部分 sacct 命令参数：

### 基本过滤参数

- `jobids` (-j, --jobs): 作业 ID 列表
- `users` (-u, --user): 用户列表
- `accounts` (-A, --accounts): 账户列表
- `partitions` (-r, --partition): 分区列表
- `states` (-s, --state): 状态列表
- `qos` (-q, --qos): QOS 列表
- `clusters` (-M, --clusters): 集群列表
- `nodelist` (-N, --nodelist): 节点列表
- `jobnames` (--name): 作业名称列表

### 时间范围

- `starttime` (-S, --starttime): 开始时间
- `endtime` (-E, --endtime): 结束时间

### 资源过滤

- `minnodes`/`maxnodes` (-i, --nnodes): 节点数量范围
- `mincpus`/`maxcpus` (-I, --ncpus): CPU 数量范围

### 输出格式控制

- `format` (-o, --format): 自定义输出格式
- `brief` (-b, --brief): 简洁模式
- `long` (-l, --long): 详细模式
- `parsable` (-p, --parsable): 可解析格式
- `noheader` (-n, --noheader): 无表头

### 其他选项

- `allusers` (-a, --allusers): 显示所有用户作业
- `allclusters` (-L, --allclusters): 显示所有集群作业
- `duplicates` (-D, --duplicates): 显示重复作业
- `truncate` (-T, --truncate): 截断时间
- `arrayjobs` (--array): 展开数组作业
- `completion` (-c, --completion): 使用作业完成数据

## 支持的 squeue 参数

本 API 支持大部分 squeue 命令参数：

### 基本过滤参数

- `accounts` (-A, --account): 账户列表
- `jobs` (-j, --job): 作业 ID 列表
- `partitions` (-p, --partition): 分区列表
- `qos` (-q, --qos): QOS 列表
- `states` (-t, --states): 状态列表
- `users` (-u, --user): 用户列表
- `names` (-n, --name): 作业名称列表
- `clusters` (-M, --clusters): 集群列表
- `licenses` (-L, --licenses): 许可证列表
- `nodelist` (-w, --nodelist): 节点列表
- `steps` (-s, --step): 作业步骤列表
- `reservation` (-R, --reservation): 预留名称

### 输出格式控制

- `format` (-o, --format): 自定义输出格式
- `format_long` (-O, --Format): 长格式输出规格
- `noheader` (-h, --noheader): 无表头
- `long` (-l, --long): 详细模式
- `noconvert` (--noconvert): 不转换单位
- `array` (-r, --array): 每行显示一个数组作业元素
- `start` (--start): 显示等待作业的预计开始时间
- `verbose` (-v, --verbose): 详细信息
- `all` (-a, --all): 显示隐藏分区的作业
- `hide` (--hide): 不显示隐藏分区的作业

### 联邦和集群选项

- `federation` (--federation): 报告联邦信息
- `local` (--local): 仅报告本地集群信息
- `sibling` (--sibling): 报告联邦集群中的兄弟作业信息
- `only_job_state` (--only-job-state): 仅查询作业状态

### 输出格式

- `json` (--json): 产生 JSON 输出
- `yaml` (--yaml): 产生 YAML 输出

### 排序和迭代

- `sort` (-S, --sort): 排序字段列表
- `iterate` (-i, --iterate): 迭代周期（秒）

### 支持的状态值

- `PENDING` (PD): 等待中
- `RUNNING` (R): 运行中
- `SUSPENDED` (S): 暂停
- `COMPLETING` (CG): 完成中
- `COMPLETED` (CD): 已完成
- `CANCELLED` (CA): 已取消
- `FAILED` (F): 失败
- `TIMEOUT` (TO): 超时
- `PREEMPTED` (PR): 被抢占
- `NODE_FAIL` (NF): 节点失败
- `REVOKED` (RV): 被撤销
- `SPECIAL_EXIT` (SE): 特殊退出
- `all`: 所有状态

## 错误处理

API 返回统一的错误格式：

```json
{
  "success": false,
  "message": "错误描述",
  "command": "执行的 sacct 命令",
  "raw_output": "原始输出（调试用）"
}
```

常见错误码：

- `400`: 请求参数错误
- `401`: 认证失败
- `404`: 作业不存在
- `500`: 服务器内部错误

## 开发

### 项目结构

```
slurm/
├── main.go              # 主程序入口
├── go.mod               # Go 模块文件
├── Makefile             # 构建脚本
├── controller/          # 控制器
│   └── sacct.go
├── middleware/          # 中间件
│   └── auth.go
├── models/              # 数据模型
│   └── sacct.go
├── utils/               # 工具函数
│   ├── ssh.go
│   └── sacct_parser.go
└── README.md            # 文档
```

### 添加新功能

1. 在 `models/` 中定义数据模型
2. 在 `utils/` 中实现业务逻辑
3. 在 `controller/` 中添加 API 接口
4. 在 `main.go` 中注册路由

### 测试

运行单元测试：

```bash
make test
```

### 代码格式化

```bash
make fmt
```

### 代码检查

```bash
make vet
```

## 部署

### 直接部署

```bash
make build
./build/slurm-jobacct
```

### Docker 部署

```bash
make docker-build
make docker-run
```

## 配置

服务支持以下环境变量：

- `PORT`: 服务端口（默认 8080）
- `GIN_MODE`: Gin 模式（debug/release）

## 安全注意事项

1. **认证**：在生产环境中应实现完善的认证机制
2. **SSH 连接**：建议使用私钥认证而非密码认证
3. **HTTPS**：生产环境应使用 HTTPS
4. **主机密钥验证**：当前跳过了 SSH 主机密钥验证，生产环境应启用

## 许可证

MIT License

## 贡献

欢迎提交 Issue 和 Pull Request！


## 🎉 最新更新 - v1.4.0

### 新增 Sinfo 集群信息 API ⭐
基于 `sinfo` 命令，提供完整的集群资源和节点状态信息：

#### 新增 API 端点：
- **GET** `/api/v1/cluster/info` - 获取集群概览信息
- **GET** `/api/v1/cluster/nodes` - 获取节点信息
- **GET** `/api/v1/cluster/partitions` - 获取分区信息  
- **GET** `/api/v1/cluster/reservations` - 获取预留信息
- **POST** `/api/v1/cluster/query` - 复杂集群查询

#### 支持的查询功能：
- ✅ 节点状态详细信息（CPU、内存、负载等）
- ✅ 分区配置和可用性
- ✅ 资源预留情况
- ✅ 实时集群状态摘要
- ✅ 自定义查询条件和输出格式

#### 使用示例：
```bash
# 获取所有节点信息
curl -X POST http://localhost:8080/api/v1/cluster/nodes \
  -H "Content-Type: application/json" \
  -d '{
    "host": "cluster.example.com",
    "username": "your_username", 
    "password": "your_password"
  }'

# 获取特定分区信息
curl -X POST http://localhost:8080/api/v1/cluster/partitions \
  -H "Content-Type: application/json" \
  -d '{
    "host": "cluster.example.com",
    "username": "your_username",
    "password": "your_password",
    "partitions": ["gpu", "cpu"]
  }'
```

### 健康检查更新
现在健康检查接口显示所有四个可用的 API：
```json
{
  "status": "ok",
  "service": "slurm-api-server", 
  "apis": ["jobacct", "jobsubmit", "jobqueue", "cluster"]
}
```

---

🚀 **SLURM API 服务器现已完整支持四大核心功能：作业记账、作业提交、队列查询和集群信息管理！**

