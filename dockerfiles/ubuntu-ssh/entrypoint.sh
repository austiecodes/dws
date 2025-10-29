#!/bin/bash
set -e

# 默认密码
DEFAULT_PASSWORD="dws"

# 从环境变量获取密码，如果未设置则使用默认值
ROOT_PASSWORD="${SSH_PASSWORD:-$DEFAULT_PASSWORD}"

# 设置 root 密码
echo "root:${ROOT_PASSWORD}" | chpasswd

# 确保 SSH host keys 存在
if [ ! -f /etc/ssh/ssh_host_rsa_key ]; then
    ssh-keygen -A
fi

# 输出启动信息（不输出密码明文）
echo "=========================================="
echo "DWS SSH Container Started"
echo "Default user: root"
echo "SSH Port: 22 (mapped to host port)"
if [ "${SSH_PASSWORD}" = "" ]; then
    echo "Using default password"
else
    echo "Using custom password"
fi
echo "=========================================="

# 执行传入的命令
exec "$@"

