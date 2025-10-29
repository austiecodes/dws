# 任务系统快速开始

本指南演示如何使用 DWS 任务系统提交和管理深度学习训练任务。

## 步骤 1: 准备环境

### 1.1 启动数据库
```bash
# 使用提供的脚本启动 PostgreSQL
./scripts/docker/pg.sh
```

### 1.2 初始化任务表
```bash
# 连接到数据库
psql -h localhost -p 5432 -U dws -d dws

# 执行迁移脚本
\i docs/sql/tasks.sql
```

## 步骤 2: 启动服务

在三个不同的终端窗口中：

```bash
# 终端 1: Platform API
go run ./cmd/platform/main.go

# 终端 2: Scheduler
go run ./cmd/scheduler/main.go

# 终端 3: Worker
go run ./cmd/worker/main.go
```

## 步骤 3: 创建容器

### 3.1 通过前端创建
1. 访问 `http://localhost:5173`
2. 登录账户
3. 进入"容器"页面
4. 选择镜像并创建容器

### 3.2 通过 API 创建
```bash
curl -X POST http://localhost:8080/api/v1/containers \
  -H "Content-Type: application/json" \
  -b cookies.txt \
  -d '{
    "image": "ghcr.io/austiecodes/ubuntu-ssh:latest",
    "password": "mypassword"
  }'
```

## 步骤 4: 提交任务

### 4.1 通过前端提交
1. 进入"任务"页面
2. 选择运行中的容器
3. 输入命令（例如：`python train.py`）
4. 设置预期时长（秒）
5. 设置优先级（可选）
6. 点击"提交任务"

### 4.2 通过 API 提交
```bash
curl -X POST http://localhost:8080/api/v1/tasks \
  -H "Content-Type: application/json" \
  -b cookies.txt \
  -d '{
    "container_id": 1,
    "command": "python -c \"import time; print(\"Hello World\"); time.sleep(5)\"",
    "expected_duration": 300,
    "priority": 0
  }'
```

## 步骤 5: 查看任务状态

### 5.1 通过前端查看
- 任务列表会自动展示所有任务
- 状态徽章显示当前状态
- 点击"查看输出"展开命令输出

### 5.2 通过 API 查看
```bash
# 列出所有任务
curl http://localhost:8080/api/v1/tasks -b cookies.txt

# 查看特定任务
curl http://localhost:8080/api/v1/tasks/1 -b cookies.txt
```

## 典型任务示例

### 示例 1: 简单脚本
```json
{
  "container_id": 1,
  "command": "echo 'Task started' && sleep 10 && echo 'Task completed'",
  "expected_duration": 15,
  "priority": 0
}
```

### 示例 2: Python 训练任务
```json
{
  "container_id": 1,
  "command": "python train.py --epochs 100 --batch-size 32",
  "expected_duration": 3600,
  "priority": 5
}
```

### 示例 3: 数据预处理
```json
{
  "container_id": 1,
  "command": "python preprocess.py --input data/raw --output data/processed",
  "expected_duration": 1800,
  "priority": 3
}
```

## 任务生命周期

```
提交任务 (pending)
    ↓
[Scheduler 5秒后调度]
    ↓
开始执行 (running)
    ↓
[Worker 执行命令]
    ↓
完成 (completed/failed)
```

## 取消任务

### 通过前端
- 点击任务卡片右上角的"取消"按钮

### 通过 API
```bash
curl -X POST http://localhost:8080/api/v1/tasks/1/cancel -b cookies.txt
```

## 超时机制

- **25 分钟**: 系统发送提醒（当前为日志 Mock）
- **30 分钟**: 系统强制终止任务，状态变为 `killed`

**查看超时警告：**
```bash
# 观察 Worker 日志
[worker] MOCK NOTIFICATION: Task 5 (user 1) has 5 minutes remaining before forced termination
```

## 调试技巧

### 查看服务日志
```bash
# Scheduler 日志
[scheduler] found 3 pending task(s)
[scheduler] dispatched task 1 (container=1, priority=0)

# Worker 日志
[worker] executing task 1: container=abc123 command=python train.py
[worker] task 1 completed with exit code 0
```

### 常见问题排查

**问题：任务一直处于 pending 状态**
- 检查 Scheduler 是否运行
- 查看 Scheduler 日志是否有错误

**问题：任务标记为 running 但无输出**
- 检查 Worker 是否运行
- 确认容器状态为 `running`
- 查看 Worker 日志中的错误信息

**问题：任务失败但无明确原因**
- 查看任务的 `output` 字段
- 检查 `exit_code` 是否为非零值
- 在容器内手动执行命令验证

## 性能测试

### 并发提交测试
```bash
# 提交 10 个任务
for i in {1..10}; do
  curl -X POST http://localhost:8080/api/v1/tasks \
    -H "Content-Type: application/json" \
    -b cookies.txt \
    -d "{
      \"container_id\": 1,
      \"command\": \"echo Task $i && sleep 5\",
      \"expected_duration\": 10,
      \"priority\": $i
    }"
done
```

### 观察调度顺序
- 高优先级任务应先被调度
- 相同优先级按提交时间先后执行

## 最佳实践

1. **设置合理的预期时长**：避免过早触发超时警告
2. **使用优先级**：紧急任务设置更高优先级
3. **监控容器资源**：确保容器有足够的 CPU/内存
4. **保存训练检查点**：防止超时导致进度丢失
5. **使用幂等命令**：任务可能被重复执行（未来功能）

## 下一步

- 阅读完整的[任务系统设计文档](TASK_SYSTEM.md)
- 查看[容器管理文档](CONTAINER_LIFECYCLE.md)
- 探索前端任务页面的高级功能

