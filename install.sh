#!/bin/bash

# GPanel 一键安装脚本
# 使用方法: curl -fsSL https://raw.githubusercontent.com/lveMonsi/GPanel/main/install.sh | sudo bash
# 或者: curl -fsSL https://raw.githubusercontent.com/lveMonsi/GPanel/main/install.sh | sudo bash -s -- v1.0.0
# 或者: sudo ./install.sh [版本号]

set -e

# ============================================================
# 配置区域
# ============================================================

# GitHub 仓库信息
GITHUB_REPO="lveMonsi/GPanel"
GITHUB_API="https://api.github.com/repos"
GITHUB_RELEASES="https://github.com"

# 安装目录
INSTALL_DIR="/opt/gpanel"
DATA_DIR="/var/lib/gpanel"
LOG_DIR="/var/log/gpanel"

# 服务名称
GPANEL_SERVICE="gpanel"
AGENT_SERVICE="gpanel-agent"

# 版本号（可通过参数指定）
VERSION="${1:-}"

# ============================================================
# 颜色定义
# ============================================================
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# ============================================================
# 工具函数
# ============================================================

print_banner() {
    echo -e "${GREEN}"
    echo "   ____   ____                           _ "
    echo "  / ___| |  _ \    __ _   _ __     ___  | |"
    echo " | |  _  | |_) |  / _\` | | '_ \   / _ \ | |"
    echo " | |_| | |  __/  | (_| | | | | | |  __/ | |"
    echo "  \____| |_|      \__,_| |_| |_|  \___| |_|"
    echo -e "${NC}"
    echo -e "${BLUE}GPanel 服务器管理面板 - 一键安装脚本${NC}"
    echo ""
}

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

check_root() {
    if [ "$EUID" -ne 0 ]; then
        log_error "请使用 root 权限运行此脚本"
        echo "使用: curl -fsSL https://raw.githubusercontent.com/lveMonsi/GPanel/main/install.sh | sudo bash"
        echo "或者: sudo ./install.sh [版本号]"
        exit 1
    fi
}

check_os() {
    if [ -f /etc/os-release ]; then
        . /etc/os-release
        OS=$ID
        SYS_VERSION=$VERSION_ID
    elif [ -f /etc/redhat-release ]; then
        OS="rhel"
    else
        OS="unknown"
    fi
    
    log_info "检测到操作系统: $OS"
}

check_arch() {
    ARCH=$(uname -m)
    case $ARCH in
        x86_64|amd64)
            ARCH="amd64"
            ;;
        aarch64|arm64)
            ARCH="arm64"
            ;;
        *)
            log_error "不支持的架构: $ARCH"
            exit 1
            ;;
    esac
    log_info "检测到架构: $ARCH"
}

# ============================================================
# 安装函数
# ============================================================

get_latest_version() {
    log_info "获取最新版本..."
    
    local api_url="${GITHUB_API}/${GITHUB_REPO}/releases/latest"
    local response
    
    # 尝试使用 GitHub API 获取最新版本
    if response=$(curl -fsSL "$api_url" 2>/dev/null); then
        VERSION=$(echo "$response" | grep -oP '"tag_name"\s*:\s*"\K[^"]+')
        if [ -n "$VERSION" ]; then
            log_info "最新版本: $VERSION"
            return 0
        fi
    fi
    
    log_warn "无法从 GitHub API 获取最新版本，尝试从 releases 页面解析..."
    
    # 备用方案：解析 releases 页面
    local releases_url="${GITHUB_RELEASES}/${GITHUB_REPO}/releases"
    if response=$(curl -fsSL "$releases_url" 2>/dev/null); then
        VERSION=$(echo "$response" | grep -oP '/releases/tag/\K[^"]+' | head -1)
        if [ -n "$VERSION" ]; then
            log_info "最新版本: $VERSION"
            return 0
        fi
    fi
    
    log_error "无法获取最新版本，请手动指定版本号"
    echo "使用方法: curl -fsSL https://raw.githubusercontent.com/lveMonsi/GPanel/main/install.sh | sudo bash -s -- v1.0.0"
    exit 1
}

download_binaries() {
    log_info "下载二进制文件..."
    
    # 如果没有指定版本，获取最新版本
    if [ -z "$VERSION" ]; then
        get_latest_version
    else
        # 确保版本号以 v 开头
        if [[ ! "$VERSION" =~ ^v ]]; then
            VERSION="v${VERSION}"
        fi
        log_info "安装版本: $VERSION"
    fi
    
    # 创建临时目录
    TEMP_DIR=$(mktemp -d)
    cd "$TEMP_DIR"
    
    # 构建下载文件名和 URL
    local archive_name="gpanel-linux-${ARCH}.tar.gz"
    local download_url="${GITHUB_RELEASES}/${GITHUB_REPO}/releases/download/${VERSION}/${archive_name}"
    local checksum_url="${GITHUB_RELEASES}/${GITHUB_REPO}/releases/download/${VERSION}/checksums.txt"
    
    log_info "下载地址: $download_url"
    
    # 下载压缩包
    if ! curl -fsSL --progress-bar -o "$archive_name" "$download_url"; then
        log_error "下载失败: $archive_name"
        log_error "请检查版本号是否正确: $VERSION"
        log_error "或访问 https://github.com/${GITHUB_REPO}/releases 查看可用版本"
        exit 1
    fi
    
    log_info "下载完成: $archive_name"
    
    # 下载校验和文件（可选）
    if curl -fsSL -o "checksums.txt" "$checksum_url" 2>/dev/null; then
        log_info "验证文件完整性..."
        if command -v sha256sum >/dev/null 2>&1; then
            # 提取当前文件的校验和并验证
            if grep -q "$archive_name" checksums.txt; then
                if sha256sum -c --ignore-missing checksums.txt 2>/dev/null; then
                    log_info "文件校验成功"
                else
                    log_warn "文件校验失败，但继续安装"
                fi
            fi
        fi
    fi
    
    # 解压文件
    log_info "解压文件..."
    if ! tar -xzf "$archive_name"; then
        log_error "解压失败"
        exit 1
    fi
    
    # 检查解压后的文件
    local bin_dir="linux-${ARCH}"
    if [ ! -d "$bin_dir" ]; then
        log_error "解压后未找到目录: $bin_dir"
        exit 1
    fi
    
    # 设置执行权限
    chmod +x "$bin_dir/gpanel" "$bin_dir/gpanel-agent" "$bin_dir/gpctl"
    
    log_info "所有文件下载完成"
}

install_binaries() {
    log_info "安装二进制文件..."
    
    # 创建目录
    mkdir -p "$INSTALL_DIR"
    mkdir -p "$DATA_DIR"
    mkdir -p "$LOG_DIR"
    
    local bin_dir="$TEMP_DIR/linux-${ARCH}"
    
    # 安装 gpanel 和 gpanel-agent 到 /opt/gpanel
    cp "$bin_dir/gpanel" "$INSTALL_DIR/"
    cp "$bin_dir/gpanel-agent" "$INSTALL_DIR/"
    chmod +x "$INSTALL_DIR/gpanel"
    chmod +x "$INSTALL_DIR/gpanel-agent"
    
    # 安装 gpctl 到 /usr/local/bin (PATH 中)
    cp "$bin_dir/gpctl" "/usr/local/bin/"
    chmod +x "/usr/local/bin/gpctl"
    
    log_info "二进制文件安装完成"
}

create_config_files() {
    log_info "创建配置目录..."
    
    # GPanel Core 和 Agent 配置均不需要配置文件
    # GPanel Core: 配置存储于数据库
    # Agent: 使用默认值或环境变量
    
    # 创建数据目录
    mkdir -p "$DATA_DIR"
    mkdir -p "$LOG_DIR"
    
    log_info "配置目录创建完成"
}

create_systemd_services() {
    log_info "创建 systemd 服务..."
    
    # 创建 gpanel-agent.service
    cat > /etc/systemd/system/gpanel-agent.service << 'EOF'
[Unit]
Description=GPanel Agent Service
After=network.target

[Service]
Type=simple
User=root
Group=root
WorkingDirectory=/opt/gpanel
ExecStart=/opt/gpanel/gpanel-agent
Restart=on-failure
RestartSec=5s

# 日志配置
StandardOutput=journal
StandardError=journal

# 安全配置
NoNewPrivileges=true
PrivateTmp=true

# 环境变量
Environment="GIN_MODE=release"

[Install]
WantedBy=multi-user.target
EOF

    # 创建 gpanel.service（依赖 agent）
    cat > /etc/systemd/system/gpanel.service << 'EOF'
[Unit]
Description=GPanel Server Service
After=network.target gpanel-agent.service
Requires=gpanel-agent.service

[Service]
Type=simple
User=root
Group=root
WorkingDirectory=/opt/gpanel
ExecStart=/opt/gpanel/gpanel
Restart=on-failure
RestartSec=5s

# 日志配置
StandardOutput=journal
StandardError=journal

# 安全配置
NoNewPrivileges=true
PrivateTmp=true

# 环境变量
Environment="GIN_MODE=release"

[Install]
WantedBy=multi-user.target
EOF

    # 重载 systemd
    systemctl daemon-reload
    
    log_info "systemd 服务创建完成"
}

configure_firewall() {
    log_info "配置防火墙..."
    
    if command -v firewall-cmd >/dev/null 2>&1; then
        # CentOS/RHEL/Fedora 使用 firewalld
        firewall-cmd --permanent --add-port=8080/tcp 2>/dev/null || true
        firewall-cmd --permanent --add-port=9998/tcp 2>/dev/null || true
        firewall-cmd --reload 2>/dev/null || true
        log_info "firewalld 防火墙规则已添加"
    elif command -v ufw >/dev/null 2>&1; then
        # Ubuntu/Debian 使用 ufw
        ufw allow 8080/tcp 2>/dev/null || true
        ufw allow 9998/tcp 2>/dev/null || true
        log_info "ufw 防火墙规则已添加"
    else
        log_warn "未检测到防火墙管理工具，请手动开放 8080 和 9998 端口"
    fi
}

start_services() {
    log_info "启动服务..."
    
    # 启用并启动 Agent 服务
    systemctl enable gpanel-agent
    systemctl start gpanel-agent
    
    # 等待 Agent 启动
    sleep 2
    
    # 启用并启动 GPanel 服务
    systemctl enable gpanel
    systemctl start gpanel
    
    # 等待服务启动
    sleep 3
    
    # 检查服务状态
    if systemctl is-active --quiet gpanel-agent && systemctl is-active --quiet gpanel; then
        log_info "服务启动成功"
    else
        log_error "服务启动失败，请检查日志"
        log_info "查看日志: journalctl -u gpanel-agent -u gpanel -f"
        exit 1
    fi
}

cleanup() {
    log_info "清理临时文件..."
    if [ -d "$TEMP_DIR" ]; then
        rm -rf "$TEMP_DIR"
    fi
}

print_success() {
    # 获取服务器 IP
    SERVER_IP=$(hostname -I | awk '{print $1}')
    if [ -z "$SERVER_IP" ]; then
        SERVER_IP="your-server-ip"
    fi
    
    echo ""
    echo -e "${GREEN}========================================${NC}"
    echo -e "${GREEN}安装成功！${NC}"
    echo -e "${GREEN}========================================${NC}"
    echo ""
    echo -e "${GREEN}安装版本:${NC} $VERSION"
    echo ""
    echo -e "${GREEN}访问地址:${NC}"
    echo "  本地访问: http://localhost:8080"
    echo "  外部访问: http://${SERVER_IP}:8080"
    echo ""
    echo -e "${GREEN}服务管理 (使用 gpctl):${NC}"
    echo "  查看状态: sudo gpctl status"
    echo "  启动服务: sudo gpctl start"
    echo "  停止服务: sudo gpctl stop"
    echo "  重启服务: sudo gpctl restart"
    echo "  获取用户信息: sudo gpctl user-info"
    echo ""
    echo -e "${GREEN}服务管理 (使用 systemctl):${NC}"
    echo "  查看状态: sudo systemctl status gpanel gpanel-agent"
    echo "  启动服务: sudo systemctl start gpanel gpanel-agent"
    echo "  停止服务: sudo systemctl stop gpanel gpanel-agent"
    echo "  重启服务: sudo systemctl restart gpanel gpanel-agent"
    echo ""
    echo -e "${GREEN}查看日志:${NC}"
    echo "  sudo journalctl -u gpanel -u gpanel-agent -f"
    echo ""
    echo -e "${GREEN}配置信息:${NC}"
    echo "  GPanel 配置: 存储于数据库 ($DATA_DIR/gpanel.db)"
    echo "  Agent 配置: 使用默认值或环境变量"
    echo ""
    echo -e "${GREEN}数据目录:${NC}"
    echo "  程序目录: $INSTALL_DIR"
    echo "  数据目录: $DATA_DIR"
    echo "  日志目录: $LOG_DIR"
    echo ""
    echo -e "${YELLOW}卸载:${NC}"
    echo "  sudo gpctl uninstall"
    echo ""
}

# ============================================================
# 主函数
# ============================================================

main() {
    print_banner
    
    # 检查 root 权限
    check_root
    
    # 检查操作系统和架构
    check_os
    check_arch
    
    # 检查依赖
    log_info "检查依赖..."
    if ! command -v curl >/dev/null 2>&1; then
        log_error "未安装 curl，请先安装 curl"
        exit 1
    fi
    
    # 下载二进制文件
    download_binaries
    
    # 安装二进制文件
    install_binaries
    
    # 创建配置文件
    create_config_files
    
    # 创建 systemd 服务
    create_systemd_services
    
    # 配置防火墙
    configure_firewall
    
    # 启动服务
    start_services
    
    # 清理临时文件
    cleanup
    
    # 显示成功信息
    print_success
}

# 运行主函数
main "$@"
