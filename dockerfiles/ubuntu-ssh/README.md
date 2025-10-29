# DWS Ubuntu SSH Image

基于 Ubuntu 24.04 的远程开发容器镜像，预装 OpenSSH 服务器和常用开发工具。

## 特性

- **基础镜像**: Ubuntu 24.04 LTS
- **SSH 服务**: 预安装并配置 OpenSSH Server
- **开发工具**: vim, git, curl, wget, build-essential
- **灵活密码**: 支持默认密码或自定义密码
- **生产就绪**: 优化的 SSH 配置，支持密钥和密码认证

## 快速开始

### 1. 构建镜像

```bash
cd dockerfiles/ubuntu-ssh
./build.sh
```

这将构建镜像 `dws/ubuntu-ssh:24.04`

### 2. 运行容器

**使用默认密码 (dws)**:
```bash
docker run -d -p 2222:22 --name dev-container dws/ubuntu-ssh:24.04
```

**使用自定义密码**:
```bash
docker run -d -p 2222:22 -e SSH_PASSWORD=your_password --name dev-container dws/ubuntu-ssh:24.04
```

### 3. SSH 连接

```bash
ssh root@localhost -p 2222
# 默认密码: dws
```

## 环境变量

| 变量名 | 默认值 | 说明 |
|--------|--------|------|
| `SSH_PASSWORD` | `dws` | root 用户的 SSH 密码 |

## 安全建议

1. **首次登录后立即修改密码**:
   ```bash
   passwd
   ```

2. **推荐使用 SSH 密钥认证**:
   ```bash
   # 本地生成密钥对
   ssh-keygen -t ed25519 -C "your_email@example.com"
   
   # 复制公钥到容器
   ssh-copy-id -p 2222 root@localhost
   
   # 禁用密码认证（可选）
   sed -i 's/PasswordAuthentication yes/PasswordAuthentication no/' /etc/ssh/sshd_config
   service ssh restart
   ```

3. **生产环境使用强密码**:
   ```bash
   docker run -d -p 2222:22 \
     -e SSH_PASSWORD=$(openssl rand -base64 32) \
     dws/ubuntu-ssh:24.04
   ```

## 平台集成

在 DWS 平台中使用此镜像：

1. **确保镜像已构建并在本地**:
   ```bash
   docker images | grep dws/ubuntu-ssh
   ```

2. **平台会自动检测此镜像**，用户可在创建容器时选择

3. **密码管理**:
   - 平台默认使用 `dws` 作为初始密码
   - 未来版本将支持在创建时自定义密码
   - 用户可在容器内使用 `passwd` 命令修改

## 包含的工具

- **编辑器**: vim
- **版本控制**: git
- **网络工具**: curl, wget
- **编译工具**: gcc, g++, make
- **系统工具**: sudo

## 扩展开发

如需添加更多工具（如 Python、Node.js、Go），修改 Dockerfile：

```dockerfile
# 在 RUN apt-get install 行添加
RUN apt-get update && apt-get install -y \
    openssh-server \
    python3 \
    python3-pip \
    nodejs \
    npm \
    golang-go \
    && rm -rf /var/lib/apt/lists/*
```

重新构建即可：
```bash
./build.sh
```

## 故障排查

### 无法连接 SSH

1. 检查容器状态:
   ```bash
   docker ps
   docker logs <container-id>
   ```

2. 检查端口映射:
   ```bash
   docker port <container-id>
   ```

3. 检查 SSH 服务:
   ```bash
   docker exec <container-id> service ssh status
   ```

### 密码认证失败

确保使用正确的密码（默认 `dws`），或检查启动时的环境变量。

## 许可证

本镜像用于 DWS 平台内部使用。

