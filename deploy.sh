#!/bin/bash

# GPanel 快速部署脚本
# 使用方法: sudo ./deploy.sh

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 检查是否为 root 用户
if [ "$EUID" -ne 0 ]; then
    echo -e "${RED}错误: 请使用 root 权限运行此脚本${NC}"
    echo "使用: sudo ./deploy.sh"
    exit 1
fi

# 获取服务器 IP 地址
SERVER_IP=$(hostname -I | awk '{print $1}')
if [ -z "$SERVER_IP" ]; then
    SERVER_IP="your-server-ip"
fi

echo -e "${GREEN}=== GPanel 部署脚本 ===${NC}"
echo ""

# 1. 检查依赖
echo -e "${YELLOW}[1/5] 检查依赖...${NC}"
command -v go >/dev/null 2>&1 || { echo -e "${RED}错误: 未安装 Go${NC}"; exit 1; }
command -v npm >/dev/null 2>&1 || { echo -e "${RED}错误: 未安装 Node.js/npm${NC}"; exit 1; }
echo -e "${GREEN}✓ 依赖检查完成${NC}"

# 2. 构建应用
echo -e "${YELLOW}[2/5] 构建应用...${NC}"
make build_linux
echo -e "${GREEN}✓ 构建完成${NC}"

# 3. 安装应用
echo -e "${YELLOW}[3/5] 安装应用...${NC}"
mkdir -p /opt/gpanel
mkdir -p /var/log/gpanel
cp build/linux-amd64/gpanel /opt/gpanel/
chmod +x /opt/gpanel/gpanel
cp build/linux-amd64/gpctl /usr/local/bin/
chmod +x /usr/local/bin/gpctl
cp config.yaml /opt/gpanel/ 2>/dev/null || true
cp gpanel.service /etc/systemd/system/
systemctl daemon-reload
echo -e "${GREEN}✓ 安装完成${NC}"

# 4. 配置防火墙
echo -e "${YELLOW}[4/5] 配置防火墙...${NC}"
if command -v firewall-cmd >/dev/null 2>&1; then
    # CentOS/RHEL/Fedora 使用 firewalld
    firewall-cmd --permanent --add-port=8080/tcp 2>/dev/null || true
    firewall-cmd --reload 2>/dev/null || true
    echo -e "${GREEN}✓ firewalld 防火墙规则已添加${NC}"
elif command -v ufw >/dev/null 2>&1; then
    # Ubuntu/Debian 使用 ufw
    ufw allow 8080/tcp 2>/dev/null || true
    echo -e "${GREEN}✓ ufw 防火墙规则已添加${NC}"
else
    echo -e "${YELLOW}⚠ 未检测到防火墙管理工具，请手动开放 8080 端口${NC}"
fi

# 5. 启动服务
echo -e "${YELLOW}[5/5] 启动服务...${NC}"
systemctl daemon-reload
systemctl enable gpanel
systemctl start gpanel
sleep 2
systemctl status gpanel --no-pager
echo -e "${GREEN}✓ 服务已启动${NC}"

# 6. 显示访问信息
echo ""
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}部署成功！${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo -e "${GREEN}访问地址:${NC}"
echo "  本地访问: http://localhost:8080"
echo "  外部访问: http://${SERVER_IP}:8080"
echo ""
echo -e "${YELLOW}注意:${NC}"
echo "  - 请确保防火墙已开放 8080 端口"
echo "  - 默认登录用户名和密码请查看配置文件"
echo ""
echo -e "${YELLOW}服务管理:${NC}"
echo "  查看状态: sudo systemctl status gpanel"
echo "  启动服务: sudo systemctl start gpanel"
echo "  停止服务: sudo systemctl stop gpanel"
echo "  重启服务: sudo systemctl restart gpanel"
echo ""
echo -e "${YELLOW}查看日志:${NC}"
echo "  sudo journalctl -u gpanel -f"
echo ""
echo -e "${YELLOW}修改配置:${NC}"
echo "  配置文件: /opt/gpanel/config.yaml"
echo "  修改后重启: sudo systemctl restart gpanel"
echo ""