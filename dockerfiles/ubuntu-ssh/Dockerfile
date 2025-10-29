FROM ubuntu:24.04

LABEL maintainer="DWS Platform"
LABEL description="Ubuntu 24.04 with OpenSSH server for remote development"

# 避免交互式提示
ENV DEBIAN_FRONTEND=noninteractive

# 安装 OpenSSH 服务器和常用工具
RUN apt-get update && apt-get install -y \
    openssh-server \
    sudo \
    curl \
    wget \
    vim \
    git \
    build-essential \
    && rm -rf /var/lib/apt/lists/*

# 创建 SSH 运行目录
RUN mkdir -p /var/run/sshd

# SSH 配置优化
RUN sed -i 's/#PermitRootLogin prohibit-password/PermitRootLogin yes/' /etc/ssh/sshd_config \
    && sed -i 's/#PasswordAuthentication yes/PasswordAuthentication yes/' /etc/ssh/sshd_config \
    && sed -i 's/#PubkeyAuthentication yes/PubkeyAuthentication yes/' /etc/ssh/sshd_config \
    && echo "UseDNS no" >> /etc/ssh/sshd_config

# 创建工作目录
RUN mkdir -p /workspace
WORKDIR /workspace

# 复制启动脚本
COPY entrypoint.sh /usr/local/bin/entrypoint.sh
RUN chmod +x /usr/local/bin/entrypoint.sh

# 暴露 SSH 端口
EXPOSE 22

# 使用 entrypoint 脚本处理密码配置
ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]
CMD ["/usr/sbin/sshd", "-D"]

