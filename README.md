# DWS - Deep Learning Web Service

多用户远程开发 + 资源隔离平台（轻量版）

一个支持容器管理、任务调度和队列系统的深度学习实验平台。

## 核心功能

### ✅ 已实现

- **用户系统**: 注册、登录、会话管理
- **容器管理**: 创建、启动、停止、删除 Docker 容器（支持 SSH 访问）
- **任务调度**: 
  - 用户提交任务到容器执行
  - 优先级队列调度
  - Scheduler 自动调度 pending 任务
  - Worker 执行任务并捕获输出
  - 30 分钟超时强制终止机制
- **前端界面**: React + TypeScript + Vite + Tailwind CSS

### 🚧 计划中

- RabbitMQ 集成（高并发场景）
- 邮件/短信/浏览器通知
- 任务依赖（DAG）
- 资源配额管理

## 快速开始

### 前置条件

- Go 1.25+
- PostgreSQL 18+
- Docker Engine
- Node.js 18+ (前端)

### 1. 启动数据库

```bash
./scripts/docker/pg.sh
```

### 2. 初始化数据库

```bash
# 使用辅助脚本
./docs/sql/scripts/init_db.sh

# 或手动执行
psql -h localhost -p 5432 -U dws -d dws
\i docs/sql/schema/01_users.sql
\i docs/sql/schema/02_containers.sql
\i docs/sql/schema/03_tasks.sql
\i docs/sql/migrations/20251030_add_soft_delete.sql
\i docs/sql/migrations/20251030_add_task_type.sql
\i docs/sql/migrations/20251031_add_jsonb_fields.sql
```

### 3. 配置应用

```bash
cp configs/app.example.toml configs/app.toml
# 编辑 app.toml 配置数据库连接等
```

### 4. 启动后端服务

```bash
# 终端 1: Platform API
go run ./cmd/platform/main.go

# 终端 2: Scheduler
go run ./cmd/scheduler/main.go

# 终端 3: Worker
go run ./cmd/worker/main.go
```

### 5. 启动前端

```bash
cd web
yarn install
yarn dev
```

访问 `http://localhost:5173`

## 架构

```
┌─────────────┐       ┌─────────────┐       ┌─────────────┐
│   Browser   │──────▶│  Platform   │──────▶│ PostgreSQL  │
│  (React)    │       │   (Gin)     │       │             │
└─────────────┘       └─────────────┘       └─────────────┘
                              │
                              │
                   ┌──────────┴──────────┐
                   ▼                     ▼
            ┌─────────────┐       ┌─────────────┐
            │  Scheduler  │       │   Worker    │
            │  (调度器)   │       │  (执行器)   │
            └─────────────┘       └─────────────┘
                   │                     │
                   └──────────┬──────────┘
                              ▼
                       ┌─────────────┐
                       │   Docker    │
                       │  Containers │
                       └─────────────┘
```

### 服务职责

- **Platform**: 用户 API、容器管理、任务提交
- **Scheduler**: 轮询 pending 任务，标记为 running
- **Worker**: 执行 running 任务，监控超时

## 目录结构

```
dws/
├── cmd/
│   ├── platform/       # Platform 服务入口
│   ├── scheduler/      # Scheduler 服务入口
│   └── worker/         # Worker 服务入口
├── internal/
│   ├── lib/            # 共享库（数据库、Docker、配置）
│   ├── platform/       # Platform 业务逻辑
│   ├── scheduler/      # Scheduler 调度逻辑
│   └── worker/         # Worker 执行逻辑
├── web/                # React 前端
├── docs/               # 文档
├── configs/            # 配置文件
└── scripts/            # 辅助脚本
```

## 文档

- SQL Schema: `docs/sql/schema/`
- SQL Migrations: `docs/sql/migrations/`
- Docker 容器管理: `internal/platform/DOCKER_MANAGE.md`
- Agents 使用指南: `AGENTS.md`

## 技术栈

**后端:**
- Go 1.25
- Gin (HTTP 框架)
- GORM (ORM)
- Docker SDK
- PostgreSQL

**前端:**
- React 18
- TypeScript
- Vite
- Tailwind CSS
- React Router

## 开发规范

### 分层架构

- `internal/lib`: 共享基础设施（禁止跨服务依赖）
- `internal/{platform,scheduler,worker}`: 服务业务逻辑（完全隔离）
- `cmd/*`: 服务入口（仅做依赖注入）

### 代码风格

- 遵循 Go 官方风格指南
- 使用 `gofmt` 格式化代码
- 通过 `golangci-lint` 检查

## 许可

MIT License

---

**注意**: 本项目处于活跃开发中，API 可能随时变更。 
