# 分布式架构Schema设计 - 完整说明

## 设计原则

### Linus式"好品味"体现

1. **列顺序逻辑清晰**
   - 主键在前
   - 外键紧随其后（表达关系）
   - 核心属性
   - 状态字段
   - 配置/元数据
   - 时间戳在最后

2. **NULL表示"可选/未分配"**
   - `worker_id` NULL = 本地容器或未调度
   - `max_containers` NULL = 无限制
   - 避免魔法值（-1, 0等）

3. **索引为查询服务**
   - 每个外键都有索引
   - 复合索引支持常见查询模式
   - 部分索引优化特定场景

## 表设计详解

### 1. Workers表

```sql
CREATE TABLE workers (
    id VARCHAR(64) PRIMARY KEY,              -- 操作员定义的ID
    name VARCHAR(255) NOT NULL,              -- 人类可读名称
    address VARCHAR(255) NOT NULL,           -- gRPC地址
    status VARCHAR(32) NOT NULL,             -- 状态
    
    -- 容量限制（NULL = 无限）
    max_containers INTEGER,
    max_cpu_cores INTEGER,
    max_memory_gb INTEGER,
    max_gpu_count INTEGER,
    
    -- 运行时状态
    last_heartbeat TIMESTAMP WITH TIME ZONE,
    
    -- 扩展元数据
    metadata JSONB NOT NULL DEFAULT '{}',
    
    -- 时间戳
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

**关键设计决策**：

- **容量限制字段明确**：不依赖metadata，调度器查询效率高
- **ID是字符串**：支持有意义的命名（worker-gpu-01），便于运维
- **Status枚举**：online/offline/maintenance，状态明确
- **Metadata扩展性**：region, zone, tags等可变属性

**索引策略**：
```sql
idx_workers_status             -- 调度器查询在线worker
idx_workers_last_heartbeat     -- 监控守护进程检测宕机
idx_workers_metadata_gin       -- 按region/tags查询
```

### 2. Containers表

```sql
CREATE TABLE containers (
    id SERIAL PRIMARY KEY,
    uuid UUID NOT NULL UNIQUE,               -- 公开API ID
    
    -- 外键（关系优先）
    user_id INTEGER NOT NULL REFERENCES users(id),
    worker_id VARCHAR(64) REFERENCES workers(id),
    
    -- Docker身份
    container_id TEXT NOT NULL UNIQUE,       -- Docker容器ID
    name TEXT NOT NULL,
    image TEXT NOT NULL,
    
    -- 网络
    host_ssh_port INTEGER NOT NULL UNIQUE,
    
    -- 状态
    status TEXT NOT NULL,
    is_deleted BOOLEAN NOT NULL DEFAULT false,
    
    -- 配置
    config JSONB NOT NULL DEFAULT '{}',
    
    -- 时间戳
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);
```

**关键设计决策**：

- **worker_id可空**：NULL表示本地容器或未调度（向后兼容）
- **user_id + worker_id在前**：表达"谁拥有"和"在哪里"
- **config灵活**：资源限制、环境变量、重启策略等放这里

**索引策略**：
```sql
idx_containers_user_id             -- 用户查看自己的容器
idx_containers_worker_id           -- Worker查询自己的容器
idx_containers_worker_status       -- 调度器查询worker的可用容器
```

### 3. Tasks表

```sql
CREATE TABLE tasks (
    id SERIAL PRIMARY KEY,
    
    -- 外键（三重关系）
    user_id INTEGER NOT NULL REFERENCES users(id),
    container_id INTEGER NOT NULL REFERENCES containers(id),
    worker_id VARCHAR(64) REFERENCES workers(id),    -- 重要！性能优化
    
    -- 任务定义
    command TEXT NOT NULL,
    task_type VARCHAR(10) NOT NULL DEFAULT 'cpu',
    expected_duration INTEGER NOT NULL,
    priority INTEGER NOT NULL DEFAULT 0,
    
    -- 状态
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    
    -- 执行追踪
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    
    -- 结果
    output TEXT,
    exit_code INTEGER,
    
    -- 扩展元数据
    metadata JSONB NOT NULL DEFAULT '{}',
    
    -- 时间戳
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

**关键设计决策**：

- **worker_id冗余字段**：虽然可以通过container反查，但：
  - 避免JOIN，查询更快
  - Scheduler分配任务到Worker时直接写入
  - 支持"按worker查询任务"的高频场景

- **三重外键**：user → container → worker，关系清晰
- **时间戳分离**：started_at, completed_at独立字段，便于查询

**索引策略**：
```sql
idx_tasks_worker_status            -- Worker查询自己要执行的任务
idx_tasks_type_status_priority     -- Scheduler按类型和优先级调度
idx_tasks_started_at               -- 超时检测（WHERE started_at < now() - interval '30 minutes'）
```

## 分布式场景支持

### 场景1：调度器分配任务

```sql
-- 1. 找到有capacity的worker
SELECT id, max_containers, 
       (SELECT COUNT(*) FROM containers WHERE worker_id = w.id AND is_deleted = false) AS current
FROM workers w
WHERE status = 'online' 
  AND (max_containers IS NULL OR current < max_containers);

-- 2. 创建容器（已分配worker_id）
INSERT INTO containers (user_id, worker_id, ...) VALUES (...);

-- 3. 创建任务（同时写入container_id和worker_id）
INSERT INTO tasks (user_id, container_id, worker_id, ...) VALUES (...);
```

### 场景2：Worker拉取任务

```sql
-- Worker启动后拉取自己的pending任务（无JOIN）
SELECT * FROM tasks 
WHERE worker_id = 'worker-gpu-01' 
  AND status = 'pending'
ORDER BY priority DESC, created_at ASC
LIMIT 10;
```

### 场景3：监控Worker健康

```sql
-- 检测宕机worker（超过5分钟无心跳）
SELECT id, name, last_heartbeat
FROM workers
WHERE status = 'online' 
  AND last_heartbeat < NOW() - INTERVAL '5 minutes';

-- 将其任务重新入队
UPDATE tasks 
SET status = 'pending', worker_id = NULL 
WHERE worker_id IN (...) AND status = 'running';
```

## GORM模型对应

### Worker模型
```go
type Worker struct {
    ID      string       `gorm:"column:id;primaryKey"`
    Name    string       `gorm:"column:name;not null"`
    Address string       `gorm:"column:address;not null"`
    Status  WorkerStatus `gorm:"column:status;not null;default:'offline'"`

    MaxContainers *int `gorm:"column:max_containers"`  // 指针 = 可NULL
    MaxCPUCores   *int `gorm:"column:max_cpu_cores"`
    MaxMemoryGB   *int `gorm:"column:max_memory_gb"`
    MaxGPUCount   *int `gorm:"column:max_gpu_count"`

    LastHeartbeat *time.Time     `gorm:"column:last_heartbeat"`
    Metadata      datatypes.JSON `gorm:"column:metadata;type:jsonb"`

    CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
    UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime"`
}
```

### Task模型增强
```go
type Task struct {
    ID          uint   `gorm:"column:id;primaryKey;autoIncrement"`
    UserID      uint   `gorm:"column:user_id;index;not null"`
    ContainerID uint   `gorm:"column:container_id;index;not null"`
    WorkerID    *string `gorm:"column:worker_id;index"`  // 新增！

    // ... 其他字段

    // Relations
    User      *User      `gorm:"foreignKey:UserID"`
    Container *Container `gorm:"foreignKey:ContainerID"`
    Worker    *Worker    `gorm:"foreignKey:WorkerID;references:ID"`  // 新增！
}
```

## 迁移路径

### 开发环境（当前）
```bash
# 完整重建
./docs/sql/scripts/rebuild_db.sh
```

### 生产环境（未来）
```sql
-- Migration 1: 增加worker capacity字段
ALTER TABLE workers 
ADD COLUMN max_containers INTEGER,
ADD COLUMN max_cpu_cores INTEGER,
ADD COLUMN max_memory_gb INTEGER,
ADD COLUMN max_gpu_count INTEGER;

-- Migration 2: tasks表增加worker_id
ALTER TABLE tasks 
ADD COLUMN worker_id VARCHAR(64) REFERENCES workers(id);

CREATE INDEX idx_tasks_worker_id ON tasks(worker_id);
CREATE INDEX idx_tasks_worker_status ON tasks(worker_id, status) WHERE worker_id IS NOT NULL;
```

## 性能考量

### 查询优化
1. **避免JOIN的热路径**：worker_id冗余避免了tasks → containers → workers的双JOIN
2. **部分索引**：`WHERE is_deleted = false` 减少索引大小
3. **GIN索引**：metadata查询使用倒排索引

### 容量规划
```
假设：
- 100个worker
- 每个worker 50个容器
- 每个容器平均10个任务/天
- 任务保留30天

存储估算：
- workers: 100行 × 1KB = 100KB
- containers: 5000行 × 2KB = 10MB
- tasks: 5000 × 10 × 30 = 1.5M行 × 5KB = 7.5GB

索引大小约等于数据大小，总计 ~15GB
```

## 好品味的体现

1. **"消除特殊情况"**：
   - NULL值的一致语义（未分配/无限制）
   - 不需要"if worker_id == -1"这种丑陋代码

2. **"数据结构决定一切"**：
   - worker_id冗余看似违反范式，但支撑了高频查询
   - 列顺序反映了思考过程（关系→属性→状态→时间）

3. **"向后兼容"**：
   - 新字段都是可空或有默认值
   - 旧代码不写worker_id也能工作（本地模式）

---

**"Talk is cheap. Show me the code."** - 这些表定义已经编译通过，可以直接部署。

