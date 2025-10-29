# 软删除与容器重启功能

## 概述

实现了容器重启功能和软删除机制，确保数据保留用于审计，同时对用户隐藏已删除的容器。

## 核心功能

### 1. 容器重启

用户可以停止并重新启动容器，无需重新创建。

**状态转换**:
```
running --[停止]--> stopped --[启动]--> running
```

**API**:
- `POST /api/v1/containers/:uuid/stop` - 停止容器
- `POST /api/v1/containers/:uuid/start` - 启动容器

### 2. 软删除

删除容器时保留数据库记录，仅标记为已删除，Docker 容器被物理删除。

**特点**:
- ✅ 数据保留用于审计
- ✅ 用户看不到已删除容器
- ✅ Docker 容器被物理删除释放资源
- ✅ 事务保证一致性

## 数据库变更

### 新增字段

```sql
ALTER TABLE containers 
ADD COLUMN is_deleted BOOLEAN DEFAULT FALSE NOT NULL;

CREATE INDEX idx_containers_is_deleted ON containers(is_deleted);
```

**字段说明**:
- `is_deleted`: 软删除标记
  - `false` (默认): 活跃容器，用户可见
  - `true`: 已删除容器，用户不可见，仅保留用于审计

### 查询过滤

所有用户查询自动过滤已删除记录：

```go
conn.WithContext(ctx).
    Where("user_id = ? AND is_deleted = ?", userID, false).
    Find(&containers)
```

## 实现细节

### 软删除流程

```go
func (s *ContainerService) Delete(ctx context.Context, userID uint, uuid string) error {
    // 1. 验证权限
    container, err := repository.Containers.GetByUUID(ctx, uuid)
    if container.UserID != userID {
        return errors.New("permission denied")
    }

    // 2. 开启事务
    tx := conn.WithContext(ctx).Begin()

    // 3. 删除 Docker 容器（物理删除）
    if err := manager.RemoveContainer(ctx, container.ContainerID); err != nil {
        tx.Rollback()
        return err
    }

    // 4. 软删除数据库记录（is_deleted = true）
    if err := tx.Model(&Container{}).
        Where("uuid = ?", uuid).
        Update("is_deleted", true).Error; err != nil {
        tx.Rollback()
        return err
    }

    // 5. 提交事务
    return tx.Commit().Error
}
```

### 事务保证

**成功场景**:
1. Docker 容器删除成功
2. 数据库更新成功
3. 提交事务 ✅

**失败场景 1 - Docker 失败**:
1. Docker 容器删除失败 ✗
2. 回滚事务
3. 数据库记录不变 ✅

**失败场景 2 - DB 失败**:
1. Docker 容器删除成功
2. 数据库更新失败 ✗
3. 回滚事务
4. Docker 容器已删除（无法回滚，但会被同步任务标记为 deleted）✅

## API 接口

### 启动容器

**Request**:
```bash
POST /api/v1/containers/:uuid/start
```

**Response**:
```json
{
  "message": "container started"
}
```

**效果**:
- Docker 容器启动
- 状态更新为 `running`

### 停止容器

**Request**:
```bash
POST /api/v1/containers/:uuid/stop
```

**Response**:
```json
{
  "message": "container stopped"
}
```

**效果**:
- Docker 容器停止
- 状态更新为 `stopped`

### 删除容器（软删除）

**Request**:
```bash
DELETE /api/v1/containers/:uuid
```

**Response**:
```json
{
  "message": "container deleted"
}
```

**效果**:
- Docker 容器物理删除
- 数据库记录 `is_deleted = true`
- 用户列表不显示

## 前端界面

### 操作按钮

```
┌──────────────────────────────────────────────┐
│ 状态    | 可用操作                          │
├──────────────────────────────────────────────┤
│ running | [停止] [删除]                     │
│ stopped | [启动] [删除]                     │
│ deleted | 不显示（已被过滤）                 │
└──────────────────────────────────────────────┘
```

### 按钮逻辑

```tsx
{container.status === "running" && (
  <Button onClick={() => handleStop(uuid)}>停止</Button>
)}
{container.status === "stopped" && (
  <Button onClick={() => handleStart(uuid)}>启动</Button>
)}
<Button variant="destructive" onClick={() => handleDelete(uuid)}>
  删除
</Button>
```

## 测试用例

### 完整流程测试

```bash
# 1. 创建容器
POST /containers → {"status": "running"} ✓

# 2. 停止容器
POST /containers/:uuid/stop → {"message": "container stopped"} ✓

# 3. 重新启动
POST /containers/:uuid/start → {"message": "container started"} ✓

# 4. 软删除
DELETE /containers/:uuid → {"message": "container deleted"} ✓

# 5. 用户不可见
GET /containers → Found in list: False ✓

# 6. 数据库记录保留
SELECT * WHERE uuid=xxx → is_deleted=true ✓
```

### 事务测试

**场景：Docker 删除失败**
```bash
# 模拟 Docker daemon 不可用
systemctl stop docker

# 尝试删除容器
DELETE /containers/:uuid
→ 500 Internal Server Error

# 验证数据库未改变
SELECT is_deleted FROM containers WHERE uuid=xxx
→ false ✓ (未被标记为删除)
```

## 数据审计

### 查询已删除容器

```sql
-- 管理员查询所有已删除容器
SELECT uuid, name, image, user_id, created_at, updated_at
FROM containers
WHERE is_deleted = true
ORDER BY updated_at DESC;
```

### 恢复容器

如需恢复（未来功能）：

```sql
-- 1. 取消软删除标记
UPDATE containers
SET is_deleted = false
WHERE uuid = 'xxx';

-- 2. 需要重新创建 Docker 容器（因为已物理删除）
-- 使用相同的配置创建新容器
```

## 性能优化

### 索引

```sql
CREATE INDEX idx_containers_is_deleted ON containers(is_deleted);
CREATE INDEX idx_containers_user_deleted ON containers(user_id, is_deleted);
```

**查询性能**:
- 用户查询容器列表：使用组合索引 `(user_id, is_deleted)`
- 过滤效率：O(log n) → O(1)

### 数据清理

建议定期清理旧的软删除记录：

```sql
-- 删除 90 天前的软删除记录
DELETE FROM containers
WHERE is_deleted = true
  AND updated_at < NOW() - INTERVAL '90 days';
```

## 对比硬删除

| 特性 | 硬删除 | 软删除 |
|------|--------|--------|
| 数据保留 | ❌ 永久丢失 | ✅ 保留记录 |
| 审计追踪 | ❌ 无法查询 | ✅ 完整历史 |
| 资源占用 | ✅ 立即释放 | ⚠️ DB 占用少量空间 |
| 恢复可能 | ❌ 不可能 | ✅ 可能（需重建容器） |
| 用户体验 | ✅ 简单 | ✅ 相同 |
| 实现复杂度 | ✅ 简单 | ⚠️ 需过滤查询 |

## 最佳实践

### 1. 定期清理

```bash
# 每月清理旧记录
0 0 1 * * /usr/bin/psql -c "DELETE FROM containers WHERE is_deleted = true AND updated_at < NOW() - INTERVAL '90 days';"
```

### 2. 审计日志

记录删除操作：

```go
log.Info("Container soft deleted",
    zap.String("uuid", uuid),
    zap.Uint("user_id", userID),
    zap.String("container_id", containerID))
```

### 3. 监控指标

- 软删除容器数量
- 数据库表大小
- 定期清理效果

## 未来增强

1. **批量操作**: 一次删除多个容器
2. **回收站**: 用户可恢复最近删除的容器
3. **自动清理**: 配置自动清理策略
4. **导出审计**: 导出已删除容器的审计报告

