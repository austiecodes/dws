# 任务调度系统设计与实现

## 概览

本文档描述了 DWS 平台的任务调度系统，该系统允许用户提交任务到他们的容器中执行，并通过 Scheduler 和 Worker 服务进行调度和执行。

## 架构组件

### 1. 数据模型 (`internal/lib/db/task.go`)

**Task 表结构：**

```go
type Task struct {
    ID               uint       // 任务ID
    UserID           uint       // 所属用户
    ContainerID      uint       // 目标容器
    Command          string     // 执行命令
    Status           TaskStatus // 任务状态
    ExpectedDuration int        // 预期时长（秒）
    Priority         int        // 优先级（越大越高）
    StartedAt        *time.Time // 开始时间
    CompletedAt      *time.Time // 完成时间
    Output           string     // 输出结果
    ExitCode         *int       // 退出码
    CreatedAt        time.Time  // 创建时间
    UpdatedAt        time.Time  // 更新时间
}
```

**状态机：**
- `pending` → `running` → `completed` | `failed` | `killed`
- Scheduler 负责 `pending` → `running` 转换
- Worker 负责 `running` → 终态 转换
- 用户或系统可以将任务标记为 `killed`

### 2. Platform API (`internal/platform`)

**REST 端点：**

| Method | Path | 功能 |
|--------|------|------|
| POST | `/api/v1/tasks` | 创建任务 |
| GET | `/api/v1/tasks` | 列出用户的所有任务 |
| GET | `/api/v1/tasks/:id` | 获取任务详情 |
| POST | `/api/v1/tasks/:id/cancel` | 取消任务 |

**创建任务请求：**
```json
{
  "container_id": 1,
  "command": "python train.py",
  "expected_duration": 600,
  "priority": 0
}
```

**验证规则：**
- 容器必须存在且属于当前用户
- 容器必须处于 `running` 状态
- 命令不能为空

### 3. Scheduler 服务 (`cmd/scheduler`, `internal/scheduler`)

**职责：**
- 周期性轮询数据库，查找 `pending` 状态的任务
- 按 `priority DESC, created_at ASC` 排序
- 将任务标记为 `running` 状态

**配置：**
- 默认轮询间隔：5 秒
- 可通过修改 `dispatcher.go` 调整策略

**运行：**
```bash
go run ./cmd/scheduler/main.go
```

### 4. Worker 服务 (`cmd/worker`, `internal/worker`)

**职责：**
- **执行器 (Executor)**: 轮询 `running` 任务，通过 Docker exec 执行命令
- **超时监控 (TimeoutWatcher)**: 监控任务运行时长，强制终止超时任务

**执行流程：**
1. 查询所有 `running` 任务
2. 解析命令为 `[]string`
3. 调用 `docker.ExecCommand(containerID, cmd)`
4. 捕获输出和退出码
5. 更新任务状态：
   - `exit_code == 0` → `completed`
   - `exit_code != 0` → `failed`

**超时机制：**
- 硬性限制：30 分钟
- 通知窗口：25 分钟（Mock 通知）
- 检查间隔：1 分钟

**运行：**
```bash
go run ./cmd/worker/main.go
```

### 5. 前端界面 (`web/src/pages/TasksPage.tsx`)

**功能：**
- 提交新任务表单（选择容器、输入命令、设置时长和优先级）
- 实时任务列表（状态徽章、容器信息、时间戳）
- 展开查看任务输出
- 取消 pending/running 任务

**状态徽章颜色：**
- 🟡 `pending` - 黄色
- 🔵 `running` - 蓝色
- 🟢 `completed` - 绿色
- 🔴 `failed` - 红色
- ⚫ `killed` - 灰色

## 数据流

```
用户提交任务 → Platform API (创建 pending 任务)
                ↓
          Scheduler (pending → running)
                ↓
          Worker Executor (执行 → completed/failed)
                ↓
          Worker TimeoutWatcher (监控 → killed)
```

## 部署与运行

### 前置条件
- PostgreSQL 数据库
- Docker Engine
- Go 1.25+

### 初始化数据库
```sql
-- 执行迁移脚本
\i docs/sql/tasks.sql
```

### 启动服务
```bash
# 1. 启动 Platform API（提供用户接口）
go run ./cmd/platform/main.go

# 2. 启动 Scheduler（调度任务）
go run ./cmd/scheduler/main.go

# 3. 启动 Worker（执行任务）
go run ./cmd/worker/main.go

# 4. 启动前端（开发模式）
cd web && yarn dev
```

## 扩展性设计

### 当前实现（数据库轮询）
- **优点：** 简单、无外部依赖、易于调试
- **缺点：** 高频轮询可能增加数据库负载

### 未来扩展（RabbitMQ）
如需支持高并发场景，可引入消息队列：

1. **Scheduler** 将任务推送到 RabbitMQ（支持优先级队列）
2. **Worker** 订阅队列消费任务
3. **优势：**
   - 减少数据库轮询压力
   - 支持多 Worker 负载均衡
   - 任务优先级原生支持

## 安全与隔离

- ✅ 用户只能操作自己的容器和任务
- ✅ 数据库级别的外键约束（`ON DELETE CASCADE`）
- ✅ 容器运行状态验证
- ⚠️ 命令执行未进行沙箱隔离（依赖容器本身的隔离）

## 监控与运维

### 日志
- `[scheduler]` 前缀：调度器日志
- `[worker]` 前缀：执行器和超时监控日志

### 关键指标
- 待处理任务数（`SELECT COUNT(*) FROM tasks WHERE status = 'pending'`）
- 平均执行时长（`completed_at - started_at`）
- 超时任务比例

## 限制与已知问题

1. **命令解析：** 当前使用简单空格分割，不支持引号包裹的参数
2. **输出大小：** 未限制 `output` 字段大小，长输出可能导致存储问题
3. **并发控制：** Worker 未限制并发执行数，可能导致资源竞争
4. **通知系统：** 当前为 Mock 实现，需要集成邮件/短信/WebSocket

## 后续优化方向

1. 引入任务依赖（DAG）
2. 支持任务重试策略
3. 实现资源配额（CPU/内存限制）
4. 添加任务执行统计和可视化
5. 实现任务模板功能

