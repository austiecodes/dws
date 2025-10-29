# 容器生命周期管理

## 概述

平台提供完整的容器生命周期管理功能：创建、停止、删除，并确保用户只能操作自己的容器。

## API 接口

### 1. 创建容器

**Endpoint**: `POST /api/v1/containers`

**Request**:
```json
{
  "image": "dws/ubuntu-ssh:24.04",
  "password": "optional_password"
}
```

**Response**:
```json
{
  "container": {
    "id": 1,
    "uuid": "aff76de1-fd7b-4ccf-8ca4-a25843e5adef",
    "name": "dws-u1-1761676795",
    "image": "dws/ubuntu-ssh:24.04",
    "host_ssh_port": 22006,
    "status": "running",
    "created_at": "2025-10-28T18:39:55Z"
  }
}
```

### 2. 停止容器

**Endpoint**: `POST /api/v1/containers/:uuid/stop`

**Response**:
```json
{
  "message": "container stopped"
}
```

**效果**:
- Docker 容器被停止（但不删除）
- 数据库状态更新为 `stopped`
- 容器可以重新启动（未来功能）

### 3. 删除容器

**Endpoint**: `DELETE /api/v1/containers/:uuid`

**Response**:
```json
{
  "message": "container deleted"
}
```

**效果**:
- Docker 容器被停止并删除
- 数据库记录被删除
- 操作不可逆

### 4. 列出容器

**Endpoint**: `GET /api/v1/containers`

**Response**:
```json
{
  "containers": [
    {
      "id": 1,
      "uuid": "...",
      "name": "dws-u1-xxx",
      "image": "dws/ubuntu-ssh:24.04",
      "host_ssh_port": 22006,
      "status": "running",
      "created_at": "2025-10-28T18:39:55Z"
    }
  ]
}
```

## 权限控制

### 用户隔离

每个用户只能操作自己创建的容器：

```go
// 获取容器并验证所有权
container, err := repository.Containers.GetByUUID(ctx, uuid)
if container.UserID != userID {
    return errors.New("permission denied")
}
```

### 测试用例

**正常操作**:
```bash
# 用户 A 创建容器
POST /api/v1/containers {"image": "..."}
# 返回 UUID: abc123

# 用户 A 停止自己的容器
POST /api/v1/containers/abc123/stop
# 200 OK
```

**权限拒绝**:
```bash
# 用户 B 尝试停止用户 A 的容器
POST /api/v1/containers/abc123/stop
# 403 Forbidden: "permission denied"
```

**未认证**:
```bash
# 没有 session cookie
POST /api/v1/containers/abc123/stop
# 401 Unauthorized: "not authenticated"
```

## 前端界面

### 容器列表

```
┌──────────────────────────────────────────────────────────────┐
│ 名称        | 镜像      | 状态    | SSH 连接 | 操作        │
├──────────────────────────────────────────────────────────────┤
│ dws-u1-xxx | ubuntu... | running | ssh ...  | [停止] [删除]│
│ dws-u1-yyy | ubuntu... | stopped | -        |       [删除] │
│ dws-u1-zzz | ubuntu... | deleted | -        |       [删除] │
└──────────────────────────────────────────────────────────────┘
```

### 按钮行为

- **停止按钮**:
  - 只对 `running` 状态容器显示
  - 点击后调用 API 停止容器
  - 本地状态立即更新为 `stopped`

- **删除按钮**:
  - 所有容器都显示（包括 stopped/deleted）
  - 点击前弹出确认对话框
  - 删除成功后从列表移除

## 状态转换

```
┌─────────┐  Create   ┌─────────┐
│  None   │ ---------> │ running │
└─────────┘            └─────────┘
                            │
                            │ Stop
                            ↓
                       ┌─────────┐
                       │ stopped │
                       └─────────┘
                            │
                            │ Delete
                            ↓
                       ┌─────────┐
                       │ deleted │
                       └─────────┘
```

**说明**:
- `running` → `stopped`: 停止容器（可恢复）
- `stopped` → `deleted`: 删除容器（不可恢复）
- `running` → `deleted`: 直接删除（先停止再删除）

## 错误处理

### 容器不存在

```bash
# 删除不存在的容器
DELETE /api/v1/containers/invalid-uuid

# 响应
{
  "error": "record not found"
}
```

### 容器已停止

```bash
# 停止已停止的容器（幂等操作）
POST /api/v1/containers/:uuid/stop

# 响应（成功，因为容器确实是停止状态）
{
  "message": "container stopped"
}
```

### Docker 错误

如果 Docker daemon 中容器不存在（但 DB 中有记录），操作仍然成功（幂等性）：

```go
if client.IsErrNotFound(err) {
    return nil // 视为成功
}
```

## 实现细节

### 幂等性设计

所有操作都是幂等的：

- **停止**: 如果容器已停止或不存在 → 成功
- **删除**: 如果容器已删除或不存在 → 成功

### 事务一致性

**删除容器**:
```go
1. 停止并删除 Docker 容器
2. 从数据库删除记录
   └─ 失败不回滚（Docker 已删除，DB 记录会被同步任务清理）
```

**停止容器**:
```go
1. 停止 Docker 容器
2. 更新数据库状态为 "stopped"
   └─ 失败不影响 Docker（状态同步任务会修正）
```

### 状态同步

后台同步任务每 5 分钟运行一次，确保 DB 状态与 Docker 一致：

- 容器被外部删除 → DB 状态更新为 `deleted`
- 容器被外部停止 → DB 状态更新为 `stopped`

## 测试

### 单元测试

```bash
go test ./internal/platform/services/... -v
```

**输出**:
```
=== RUN   TestPermissionDenied
=== RUN   TestPermissionDenied/same_user_can_access
=== RUN   TestPermissionDenied/different_user_denied
--- PASS: TestPermissionDenied (0.00s)
PASS
ok      github.com/austiecodes/dws/internal/platform/services  0.302s
```

### 集成测试

```bash
# 创建容器
curl -X POST -d '{"image": "dws/ubuntu-ssh:24.04"}' \
  http://localhost:8080/api/v1/containers

# 停止容器
curl -X POST http://localhost:8080/api/v1/containers/:uuid/stop

# 删除容器
curl -X DELETE http://localhost:8080/api/v1/containers/:uuid
```

## 安全考虑

1. **认证**: 所有操作需要认证（session cookie）
2. **授权**: 用户只能操作自己的容器
3. **幂等性**: 防止重复操作导致错误
4. **输入验证**: UUID 格式验证
5. **确认对话框**: 删除操作需要前端确认

## 后续优化

1. **重启容器**: 添加 `POST /containers/:uuid/start` 重新启动已停止的容器
2. **批量操作**: 支持一次停止/删除多个容器
3. **日志查看**: `GET /containers/:uuid/logs` 查看容器日志
4. **资源限制**: CPU/内存限制配置
5. **定时清理**: 自动删除长期停止的容器

