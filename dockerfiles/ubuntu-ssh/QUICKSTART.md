# 快速开始指南

## 1. 构建镜像

```bash
cd /Users/austie/codes/dws/dockerfiles/ubuntu-ssh
./build.sh
```

预期输出：
```
Building dws/ubuntu-ssh:24.04...
[+] Building ... done
Build completed: dws/ubuntu-ssh:24.04
```

验证镜像：
```bash
docker images | grep dws/ubuntu-ssh
```

应该看到：
```
dws/ubuntu-ssh    24.04    <image-id>    <time>    <size>
```

## 2. 测试镜像

### 测试 1: 使用默认密码

```bash
docker run -d --name test-ssh -p 2222:22 dws/ubuntu-ssh:24.04
```

连接测试：
```bash
ssh root@localhost -p 2222
# 密码: dws
```

清理：
```bash
docker stop test-ssh && docker rm test-ssh
```

### 测试 2: 使用自定义密码

```bash
docker run -d --name test-ssh -p 2222:22 \
  -e SSH_PASSWORD=mypassword \
  dws/ubuntu-ssh:24.04
```

连接测试：
```bash
ssh root@localhost -p 2222
# 密码: mypassword
```

清理：
```bash
docker stop test-ssh && docker rm test-ssh
```

## 3. 平台集成测试

### 3.1 重启 Platform 服务

```bash
# 停止旧服务
pkill -f 'tmp/platform/platform'

# 重新编译（如果有代码更改）
cd /Users/austie/codes/dws
go build -o tmp/platform/platform ./cmd/platform

# 启动服务
./tmp/platform/platform
```

### 3.2 测试 API

获取镜像列表（应包含 dws/ubuntu-ssh:24.04）：
```bash
curl -s -H "Cookie: <your-session-cookie>" \
  http://localhost:8080/api/v1/containers/images | jq .
```

创建容器（默认密码 "dws"）：
```bash
curl -X POST \
  -H "Cookie: <your-session-cookie>" \
  -H "Content-Type: application/json" \
  -d '{"image": "dws/ubuntu-ssh:24.04"}' \
  http://localhost:8080/api/v1/containers
```

响应示例：
```json
{
  "container": {
    "uuid": "...",
    "name": "dws-u1-1234567890",
    "image": "dws/ubuntu-ssh:24.04",
    "host_ssh_port": 22001,
    "status": "running"
  }
}
```

连接到创建的容器：
```bash
ssh root@localhost -p 22001
# 密码: dws
```

## 4. 前端测试

1. 确保前端 Vite 开发服务器运行在 `http://localhost:5173`
2. 登录平台
3. 进入 "Containers" 页面
4. 应该能看到 `dws/ubuntu-ssh:24.04` 在镜像列表中
5. 选择该镜像并创建容器
6. 使用显示的 SSH 端口和密码 `dws` 连接

## 5. 故障排查

### 镜像未出现在列表中

```bash
# 检查镜像是否存在
docker images | grep dws/ubuntu-ssh

# 如果不存在，重新构建
cd dockerfiles/ubuntu-ssh && ./build.sh
```

### 容器创建失败

查看 platform 服务日志，通常原因：
- Docker daemon 未运行
- 端口范围已用完（检查 `configs/app.toml` 中的 `ssh_port_range_start/end`）
- 镜像不存在或损坏

### SSH 连接失败

```bash
# 检查容器是否运行
docker ps | grep dws-u

# 查看容器日志
docker logs <container-id>

# 检查 SSH 服务
docker exec <container-id> service ssh status

# 测试端口连通性
nc -zv localhost <ssh-port>
```

## 6. 生产部署建议

1. **修改默认密码**：首次登录后立即修改
   ```bash
   passwd
   ```

2. **使用 SSH 密钥**：推荐配置密钥认证
   ```bash
   ssh-copy-id -p <port> root@<host>
   ```

3. **禁用密码认证**（可选）：
   ```bash
   sed -i 's/PasswordAuthentication yes/PasswordAuthentication no/' /etc/ssh/sshd_config
   service ssh restart
   ```

4. **配置防火墙**：限制 SSH 访问来源

5. **定期更新**：保持基础镜像和软件包更新
   ```dockerfile
   # 在 Dockerfile 中
   RUN apt-get update && apt-get upgrade -y
   ```

