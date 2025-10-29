## Docker 管理器设计

- 统一使用 `internal/lib/docker` 中的 `Manager` 管理容器的创建与销毁。程序启动时由 `cmd/platform/main.go` 调用 `docker.Init(cfg.Docker)` 初始化。
- `internal/platform/services/container_service.go` 通过 `libdocker.MustInstance()` 获取客户端，并负责：
  - 校验镜像是否在 `[docker.allowed_images]` 白名单内；
  - 从 `[docker.ssh_port_range_start, ssh_port_range_end]` 分配未占用的宿主机端口；
  - 调用 Docker API 创建容器并绑定 22 端口到宿主机供 SSH/VSCode 使用；
  - 写入 `containers` 表，记录 `uuid/container_id/host_ssh_port/status` 等字段。
- 失败时会回滚：DB 写入失败则立即清理刚创建的容器，避免遗留。

## 配置说明

`configs/app.toml` 中新增 `[docker]` 段：

```toml
[docker]
host = "unix:///var/run/docker.sock"
api_version = ""
ssh_port_range_start = 22000
ssh_port_range_end = 22999
allowed_images = ["ubuntu:22.04", "nvidia/cuda:12.4.1-base-ubuntu22.04"]
network = ""
```

- `allowed_images` 为空表示不限制；非空则前端只能在白名单中选择。
- `ssh_port_range_*` 定义宿主机可用于 SSH 暴露的端口区间，服务会跳过已被数据库记录或当前系统占用的端口。

## 接口速览

- `POST /api/v1/containers` 创建容器，传入镜像名；
- `GET /api/v1/containers` 查看当前用户创建的容器列表；
- `GET /api/v1/containers/images` 返回可选镜像列表。

返回的 `host_ssh_port` 即宿主机端口，用户可通过 `ssh user@host -p <port>` 或 VSCode Remote SSH 连接。
