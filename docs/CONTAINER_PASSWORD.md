# 容器 SSH 密码功能

## 概述

平台支持在创建容器时自定义 SSH 密码，增强安全性和灵活性。

## 功能特性

### 1. 默认密码

如果不指定密码，系统使用默认密码 `dws`：

**API 请求**：
```json
{
  "image": "dws/ubuntu-ssh:24.04"
}
```

**SSH 连接**：
```bash
ssh root@localhost -p <port>
# 密码: dws
```

### 2. 自定义密码

创建容器时可指定自定义密码：

**API 请求**：
```json
{
  "image": "dws/ubuntu-ssh:24.04",
  "password": "your_secure_password"
}
```

**SSH 连接**：
```bash
ssh root@localhost -p <port>
# 密码: your_secure_password
```

## 前端界面

### 创建容器表单

```
┌─────────────────────────────────────┐
│ 选择镜像                            │
│ [dws/ubuntu-ssh:24.04 ▼]           │
│                                     │
│ SSH 密码（可选）                    │
│ [••••••••••••]                     │
│ 留空使用默认密码 dws                │
│ 自定义密码后请妥善保管，             │
│ 平台不会存储您的密码                │
│                                     │
│ [创建容器]                          │
└─────────────────────────────────────┘
```

### 容器列表

```
┌────────────────────────────────────────────────────────────────┐
│ 名称              | 镜像            | 状态    | SSH 连接       │
├────────────────────────────────────────────────────────────────┤
│ dws-u1-xxx       | ubuntu-ssh:24.04 | running | ssh root@...  │
│                  |                  |         | 密码: ***      │
└────────────────────────────────────────────────────────────────┘
```

- **运行中容器**：显示 SSH 命令和脱敏密码提示
- **停止/删除容器**：显示 "容器未运行"

## 安全性

### 密码存储

- ✅ **不存储明文密码**：平台不在数据库中存储用户密码
- ✅ **环境变量传递**：通过 Docker 环境变量 `SSH_PASSWORD` 传递
- ✅ **容器内设置**：密码在容器启动时由 entrypoint 脚本设置

### 密码脱敏

前端展示时密码被脱敏为 `***`，避免明文暴露。

### 修改密码

用户可在容器内使用标准 Linux 命令修改密码：

```bash
ssh root@localhost -p <port>
passwd
# 输入新密码
```

## API 接口

### 创建容器

**Endpoint**: `POST /api/v1/containers`

**Request Body**:
```json
{
  "image": "dws/ubuntu-ssh:24.04",
  "password": "optional_custom_password"
}
```

**字段说明**:
- `image` (required): 镜像名称
- `password` (optional): SSH 密码，留空使用默认值 `dws`

**Response**:
```json
{
  "container": {
    "id": 1,
    "uuid": "...",
    "name": "dws-u1-xxx",
    "image": "dws/ubuntu-ssh:24.04",
    "host_ssh_port": 22001,
    "status": "running",
    "created_at": "2025-10-28T18:00:00Z"
  }
}
```

## 实现原理

### 后端流程

1. **Handler** 接收 `password` 参数（可选）
2. **Service** 调用 `CreateWithOptions`，传递密码
3. **Docker Manager** 将密码设置为环境变量 `SSH_PASSWORD`
4. **容器启动** entrypoint.sh 读取环境变量并设置 root 密码

### 代码路径

| 组件 | 文件 | 说明 |
|------|------|------|
| Handler | `internal/platform/handlers/container_handler.go` | 接收 password 参数 |
| Service | `internal/platform/services/container_service.go` | `CreateWithOptions` 方法 |
| Manager | `internal/lib/docker/manager.go` | 设置 `SSH_PASSWORD` 环境变量 |
| Entrypoint | `dockerfiles/ubuntu-ssh/entrypoint.sh` | 读取环境变量设置密码 |

### 环境变量流

```
API Request (password: "test123")
         ↓
Service.CreateWithOptions(opts)
         ↓
Manager.CreateSSHContainer(opts)
         ↓
Docker Container with SSH_PASSWORD=test123
         ↓
entrypoint.sh: echo "root:$SSH_PASSWORD" | chpasswd
         ↓
Container with root password = "test123"
```

## 最佳实践

### 开发环境

使用默认密码 `dws` 快速开发：
```json
{
  "image": "dws/ubuntu-ssh:24.04"
}
```

### 生产环境

1. **强密码策略**：使用至少 12 位混合字符密码
2. **密钥认证**：上传 SSH 公钥，禁用密码认证
3. **定期轮换**：定期修改密码

### 密码强度建议

```bash
# 生成强密码
openssl rand -base64 32

# 或使用密码管理器生成
# 1Password / Bitwarden / KeePass
```

## 故障排查

### 无法连接容器

1. **检查容器状态**：
   ```bash
   docker ps | grep dws-u1
   ```

2. **查看容器日志**：
   ```bash
   docker logs <container-name>
   ```
   应看到：
   ```
   ==========================================
   DWS SSH Container Started
   Default user: root
   SSH Port: 22 (mapped to host port)
   Using custom password
   ==========================================
   ```

3. **测试 SSH 连接**：
   ```bash
   ssh root@localhost -p <port>
   ```

### 密码认证失败

- 确认使用正确的密码（默认 `dws` 或自定义密码）
- 检查容器日志确认密码是否成功设置
- 尝试重新创建容器

## 未来增强

1. **密码复杂度验证**：前端验证密码强度
2. **密钥管理**：支持上传 SSH 公钥
3. **多用户支持**：创建非 root 用户
4. **密码找回**：管理员重置密码功能

