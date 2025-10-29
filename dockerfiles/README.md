# DWS Docker Images

此目录包含 DWS 平台的标准容器镜像定义。

## 可用镜像

### ubuntu-ssh

基于 Ubuntu 24.04 的远程开发容器，预装 OpenSSH 服务器。

**特性**:
- Ubuntu 24.04 LTS 基础
- 预装 SSH 服务器
- 包含常用开发工具（vim, git, curl, build-essential）
- 支持密码和密钥认证
- 可配置密码

**快速开始**:
```bash
cd ubuntu-ssh
./build.sh
```

详细文档见 [`ubuntu-ssh/README.md`](./ubuntu-ssh/README.md)

## 添加新镜像

1. 创建新目录：`dockerfiles/<image-name>/`
2. 添加必需文件：
   - `Dockerfile` - 镜像定义
   - `build.sh` - 构建脚本
   - `README.md` - 使用文档
3. 如需启动脚本，添加 `entrypoint.sh`
4. 在本文档中添加镜像说明

## 镜像命名规范

- **格式**: `dws/<name>:<tag>`
- **示例**: 
  - `dws/ubuntu-ssh:24.04`
  - `dws/python-dev:3.12`
  - `dws/cuda-torch:12.4`

## 平台集成

构建的镜像会自动被 DWS 平台检测到（通过 `docker images`）。用户可在创建容器时选择这些镜像。

### SSH 容器要求

如果镜像需要支持 SSH 连接（推荐），需满足：

1. **暴露 22 端口**：
   ```dockerfile
   EXPOSE 22
   ```

2. **SSH 服务**：安装并配置 OpenSSH Server

3. **密码配置**：通过环境变量 `SSH_PASSWORD` 支持动态密码
   ```bash
   docker run -e SSH_PASSWORD=mypass dws/your-image
   ```

4. **Entrypoint 脚本**：处理密码设置和 SSH 启动

参考 `ubuntu-ssh` 的实现作为模板。

## 最佳实践

1. **最小化镜像体积**：
   - 使用多阶段构建
   - 清理 apt 缓存：`rm -rf /var/lib/apt/lists/*`
   - 合并 RUN 命令减少层数

2. **安全性**：
   - 不在镜像中硬编码密码
   - 使用环境变量传递敏感信息
   - 定期更新基础镜像

3. **可维护性**：
   - 提供清晰的文档
   - 使用构建脚本标准化流程
   - 添加版本标签

4. **测试**：
   - 构建后测试基本功能
   - 验证 SSH 连接
   - 确保工具可用

## 镜像生命周期

1. **开发**: 在 `dockerfiles/<name>/` 中创建和测试
2. **构建**: 使用 `build.sh` 脚本
3. **验证**: 本地测试功能
4. **部署**: 镜像自动出现在平台镜像列表
5. **维护**: 定期更新和重新构建

## 故障排查

### 镜像未出现在平台

```bash
# 检查镜像是否存在
docker images | grep dws/

# 重启 platform 服务
pkill -f platform && ./tmp/platform/platform
```

### SSH 连接问题

```bash
# 测试容器
docker run -d -p 2222:22 dws/your-image
ssh root@localhost -p 2222

# 查看日志
docker logs <container-id>
```

### 构建失败

- 检查 Dockerfile 语法
- 验证基础镜像可访问
- 确保 Docker daemon 运行

## 贡献指南

添加新镜像时：

1. 遵循命名规范
2. 提供完整文档
3. 测试所有功能
4. 更新本 README

