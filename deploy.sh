#!/bin/bash

# GPanel 快速部署脚本
# 使用方法: sudo ./deploy.sh [domain]

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 检查是否为 root 用户
if [ "$EUID" -ne 0 ]; then
    echo -e "${RED}错误: 请使用 root 权限运行此脚本${NC}"
    echo "使用: sudo ./deploy.sh [domain]"
    exit 1
fi

# 获取域名参数
DOMAIN=${1:-""}
if [ -z "$DOMAIN" ]; then
    echo -e "${YELLOW}警告: 未提供域名，将使用默认配置${NC}"
    echo "使用方法: sudo ./deploy.sh your-domain.com"
fi

echo -e "${GREEN}=== GPanel 部署脚本 ===${NC}"
echo ""

# 1. 检查依赖
echo -e "${YELLOW}[1/6] 检查依赖...${NC}"
command -v go >/dev/null 2>&1 || { echo -e "${RED}错误: 未安装 Go${NC}"; exit 1; }
command -v npm >/dev/null 2>&1 || { echo -e "${RED}错误: 未安装 Node.js/npm${NC}"; exit 1; }
command -v nginx >/dev/null 2>&1 || { echo -e "${RED}错误: 未安装 Nginx${NC}"; exit 1; }
echo -e "${GREEN}✓ 依赖检查完成${NC}"

# 2. 构建应用
echo -e "${YELLOW}[2/6] 构建应用...${NC}"
make build_linux
echo -e "${GREEN}✓ 构建完成${NC}"

# 3. 安装应用
echo -e "${YELLOW}[3/6] 安装应用...${NC}"
make install
echo -e "${GREEN}✓ 安装完成${NC}"

# 4. 配置 Nginx
if [ -n "$DOMAIN" ]; then
    echo -e "${YELLOW}[4/6] 配置 Nginx (域名: $DOMAIN)...${NC}"

    NGINX_CONF="/etc/nginx/sites-available/gpanel"
    sed "s/your-domain.com/$DOMAIN/g" nginx.example.conf > "$NGINX_CONF"

    # 创建软链接
    ln -sf "$NGINX_CONF" /etc/nginx/sites-enabled/

    # 测试 Nginx 配置
    nginx -t

    echo -e "${GREEN}✓ Nginx 配置完成${NC}"
    echo -e "${YELLOW}提示: 请手动配置 SSL 证书，参考 nginx.example.conf${NC}"
else
    echo -e "${YELLOW}[4/6] 跳过 Nginx 配置（未提供域名）${NC}"
fi

# 5. 启动服务
echo -e "${YELLOW}[5/6] 启动服务...${NC}"
systemctl start gpanel
systemctl status gpanel --no-pager
echo -e "${GREEN}✓ 服务已启动${NC}"

# 6. 显示访问信息
echo -e "${YELLOW}[6/6] 部署完成！${NC}"
echo ""
echo -e "${GREEN}=== 访问信息 ===${NC}"
echo "应用地址: http://localhost:8080"
if [ -n "$DOMAIN" ]; then
    echo "域名地址: http://$DOMAIN"
    echo ""
    echo -e "${YELLOW}下一步操作:${NC}"
    echo "1. 配置 SSL 证书:"
    echo "   sudo certbot --nginx -d $DOMAIN"
    echo ""
    echo "2. 重启 Nginx:"
    echo "   sudo systemctl restart nginx"
fi
echo ""
echo -e "${YELLOW}查看日志:${NC}"
echo "   sudo journalctl -u gpanel -f"
echo ""
echo -e "${YELLOW}服务管理:${NC}"
echo "   启动: sudo systemctl start gpanel"
echo "   停止: sudo systemctl stop gpanel"
echo "   重启: sudo systemctl restart gpanel"
echo "   状态: sudo systemctl status gpanel"
echo ""
echo -e "${GREEN}部署成功！${NC}"