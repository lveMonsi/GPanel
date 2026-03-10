#!/bin/bash

# GPanel 管理脚本
# 使用方法:
#   安装最新版本: curl -fsSL https://raw.githubusercontent.com/lveMonsi/GPanel/main/gpanel.sh | sudo bash -s -- install
#   安装指定版本: curl -fsSL https://raw.githubusercontent.com/lveMonsi/GPanel/main/gpanel.sh | sudo bash -s -- install v1.0.0
#   国内镜像安装: curl -fsSL https://gh.llkk.cc/https://raw.githubusercontent.com/lveMonsi/GPanel/main/gpanel.sh | sudo bash -s -- install
#   更新到最新:   sudo ./gpanel.sh update
#   更新到指定:   sudo ./gpanel.sh update v1.0.0
#   卸载:         sudo ./gpanel.sh uninstall
#   查看状态:     sudo ./gpanel.sh status
#   帮助:         sudo ./gpanel.sh help

set -e

# ============================================================
# 配置区域
# ============================================================

# GitHub 仓库信息
GITHUB_REPO="lveMonsi/GPanel"
GITHUB_API="https://api.github.com/repos"
GITHUB_RELEASES="https://github.com"

# GitHub 代理（国内加速）
GITHUB_PROXY="https://gh.llkk.cc/"

# 安装目录
INSTALL_DIR="/opt/gpanel"
DATA_DIR="/var/lib/gpanel"
LOG_DIR="/var/log/gpanel"

# 服务名称
GPANEL_SERVICE="gpanel"
AGENT_SERVICE="gpanel-agent"

# 安装信息文件
INSTALL_INFO_FILE="$INSTALL_DIR/.install_info"

# 版本号
VERSION=""
ACTION=""

# 交互模式标志
INTERACTIVE=true

# 终端输入设备（用于管道运行时从终端读取输入）
TTY_DEVICE="/dev/tty"

# 检测是否可以通过 /dev/tty 读取输入
if [ ! -c "$TTY_DEVICE" ] && [ ! -e "$TTY_DEVICE" ]; then
    INTERACTIVE=false
fi

# ============================================================
# 颜色定义
# ============================================================
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
MAGENTA='\033[0;35m'
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
    echo -e "${BLUE}GPanel 服务器管理面板 - 管理脚本${NC}"
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

log_step() {
    echo -e "${CYAN}[STEP]${NC} $1"
}

check_root() {
    if [ "$EUID" -ne 0 ]; then
        log_error "请使用 root 权限运行此脚本"
        echo "使用: sudo ./gpanel.sh <command> [options]"
        exit 1
    fi
}

check_os() {
    if [ -f /etc/os-release ]; then
        OS=$(grep -oP '^ID=\K.*' /etc/os-release 2>/dev/null | tr -d '"' || echo "unknown")
        OS_VERSION=$(grep -oP '^VERSION_ID=\K.*' /etc/os-release 2>/dev/null | tr -d '"' || echo "unknown")
    elif [ -f /etc/redhat-release ]; then
        OS="rhel"
        OS_VERSION="unknown"
    else
        OS="unknown"
        OS_VERSION="unknown"
    fi
    
    log_info "操作系统: $OS $OS_VERSION"
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
    log_info "系统架构: $ARCH"
}

# ============================================================
# 端口获取函数
# ============================================================

# 获取 Core 服务实际端口
get_core_port() {
    # 默认端口
    local default_port="8080"
    
    # 如果 gpctl 已安装，尝试获取实际端口
    if command -v gpctl >/dev/null 2>&1; then
        local port=$(gpctl get-port core --quiet 2>/dev/null)
        if [ -n "$port" ] && [ "$port" != "" ]; then
            echo "$port"
            return 0
        fi
    fi
    
    echo "$default_port"
}

# 获取 Agent 服务实际端口
get_agent_port() {
    # 默认端口
    local default_port="9998"
    
    # 如果 gpctl 已安装，尝试获取实际端口
    if command -v gpctl >/dev/null 2>&1; then
        local port=$(gpctl get-port agent --quiet 2>/dev/null)
        if [ -n "$port" ] && [ "$port" != "" ]; then
            echo "$port"
            return 0
        fi
    fi
    
    echo "$default_port"
}

# 获取安全入口
get_security_entrance() {
    # 默认安全入口
    local default_entrance="/"
    
    # 如果 gpctl 已安装，尝试获取实际安全入口
    if command -v gpctl >/dev/null 2>&1; then
        local entrance=$(gpctl get-entrance --quiet 2>/dev/null)
        if [ -n "$entrance" ] && [ "$entrance" != "" ]; then
            echo "$entrance"
            return 0
        fi
    fi
    
    echo "$default_entrance"
}

# 缓存端口变量（在脚本开始时调用）
cache_ports() {
    CORE_PORT=$(get_core_port)
    AGENT_PORT=$(get_agent_port)
}

# ============================================================
# 交互式函数
# ============================================================

# 确认操作 (是/否)
confirm() {
    local prompt="$1"
    local default="${2:-n}"
    local reply
    
    if [ "$default" = "y" ]; then
        echo -ne "${YELLOW}${prompt} [Y/n]: ${NC}"
    else
        echo -ne "${YELLOW}${prompt} [y/N]: ${NC}"
    fi
    
    if [ "$INTERACTIVE" = true ]; then
        read -r reply < "$TTY_DEVICE"
    else
        reply="$default"
        echo "$reply"
    fi
    
    reply="${reply:-$default}"
    
    case "$reply" in
        y*|Y*) return 0 ;;
        *) return 1 ;;
    esac
}

# 选择操作 (从选项列表中选择)
select_option() {
    local prompt="$1"
    shift
    local options=("$@")
    local reply
    
    echo -e "${CYAN}${prompt}${NC}"
    echo ""
    
    echo -e "  ${GREEN}0)${NC} 退出"
    local i=1
    for opt in "${options[@]}"; do
        echo -e "  ${GREEN}$i)${NC} $opt"
        ((i++))
    done
    echo ""
    
    while true; do
        echo -ne "${YELLOW}请选择 [0-${#options[@]}]: ${NC}"
        
        if [ "$INTERACTIVE" = true ]; then
            read -r reply < "$TTY_DEVICE"
        else
            echo "1"
            reply=1
        fi
        
        if [ "$reply" = "0" ]; then
            SELECTED_OPTION="__EXIT__"
            return 1
        fi
        
        if [[ "$reply" =~ ^[0-9]+$ ]] && [ "$reply" -ge 1 ] && [ "$reply" -le ${#options[@]} ]; then
            SELECTED_OPTION="${options[$((reply-1))]}"
            return 0
        fi
        
        log_error "无效选择，请输入 0-${#options[@]}"
    done
}

# 选择版本 (从版本列表中选择)
select_version() {
    local prompt="$1"
    local versions=("${@:2}")
    local reply
    
    echo -e "${CYAN}${prompt}${NC}"
    echo ""
    
    echo -e "  ${GREEN}0)${NC} 退出"
    local i=1
    for ver in "${versions[@]}"; do
        echo -e "  ${GREEN}$i)${NC} $ver"
        ((i++))
    done
    echo -e "  ${GREEN}$i)${NC} 手动输入版本号"
    echo ""
    
    while true; do
        echo -ne "${YELLOW}请选择 [0-${#versions[@]}+1]: ${NC}"
        
        if [ "$INTERACTIVE" = true ]; then
            read -r reply < "$TTY_DEVICE"
        else
            echo "1"
            reply=1
        fi
        
        if [ "$reply" = "0" ]; then
            SELECTED_VERSION=""
            return 1
        fi
        
        if [[ "$reply" =~ ^[0-9]+$ ]] && [ "$reply" -ge 1 ] && [ "$reply" -le ${#versions[@]} ]; then
            SELECTED_VERSION="${versions[$((reply-1))]}"
            return 0
        fi
        
        # 手动输入版本号（最后一个选项）
        if [ "$reply" -eq $((${#versions[@]}+1)) ]; then
            echo -ne "${YELLOW}请输入版本号 (如 v1.0.0): ${NC}"
            if [ "$INTERACTIVE" = true ]; then
                read -r SELECTED_VERSION < "$TTY_DEVICE"
            else
                SELECTED_VERSION="${versions[0]}"
                echo "$SELECTED_VERSION"
            fi
            return 0
        fi
        
        log_error "无效选择，请输入 0-$((${#versions[@]}+1))"
    done
}

# 输入确认
input_confirm() {
    local prompt="$1"
    local reply
    
    echo -ne "${YELLOW}${prompt}: ${NC}"
    
    if [ "$INTERACTIVE" = true ]; then
        read -r reply < "$TTY_DEVICE"
    else
        reply="y"
        echo "$reply"
    fi
    
    INPUT_RESULT="$reply"
}

# 按任意键继续
press_any_key() {
    echo -ne "${CYAN}按任意键继续...${NC}"
    if [ "$INTERACTIVE" = true ]; then
        read -n 1 -r -s < "$TTY_DEVICE"
    fi
    echo ""
}

# ============================================================
# 状态检测函数
# ============================================================

# 从数据库获取用户信息
get_db_user_info() {
    local db_file="$DATA_DIR/gpanel.db"
    
    if [ ! -f "$db_file" ]; then
        return 1
    fi
    
    if ! command -v sqlite3 >/dev/null 2>&1; then
        return 1
    fi
    
    DB_USERNAME=$(sqlite3 "$db_file" "SELECT value FROM settings WHERE key='PanelUser' LIMIT 1;" 2>/dev/null)
    DB_PASSWORD=$(sqlite3 "$db_file" "SELECT value FROM settings WHERE key='PanelPassword' LIMIT 1;" 2>/dev/null)
    
    if [ -n "$DB_USERNAME" ] && [ -n "$DB_PASSWORD" ]; then
        return 0
    fi
    
    return 1
}

# 检查是否已安装
is_installed() {
    if [ -f "$INSTALL_DIR/gpanel" ] && [ -f "$INSTALL_DIR/gpanel-agent" ]; then
        return 0
    fi
    return 1
}

# 获取已安装版本
get_installed_version() {
    if [ -f "$INSTALL_INFO_FILE" ]; then
        grep -oP '^version=\K.*' "$INSTALL_INFO_FILE" 2>/dev/null || echo "unknown"
    else
        echo "unknown"
    fi
}

# 获取安装时间
get_install_time() {
    if [ -f "$INSTALL_INFO_FILE" ]; then
        grep -oP '^install_time=\K.*' "$INSTALL_INFO_FILE" 2>/dev/null || echo "unknown"
    else
        echo "unknown"
    fi
}

# 检查是否为预发布版本
is_prerelease_version() {
    if [ -f "$INSTALL_INFO_FILE" ]; then
        local is_pre=$(grep -oP '^is_prerelease=\K.*' "$INSTALL_INFO_FILE" 2>/dev/null || echo "false")
        if [ "$is_pre" = "true" ]; then
            return 0
        fi
    fi
    # 也通过版本号判断
    local version=$(get_installed_version)
    if [[ "$version" =~ ^pre-release- ]]; then
        return 0
    fi
    return 1
}

# 检查服务运行状态
get_service_status() {
    local service=$1
    if systemctl is-active --quiet "$service" 2>/dev/null; then
        echo "running"
    elif systemctl is-enabled --quiet "$service" 2>/dev/null; then
        echo "stopped"
    else
        echo "not_installed"
    fi
}

# 获取服务 PID
get_service_pid() {
    local service=$1
    systemctl show --property MainPID --value "$service" 2>/dev/null || echo ""
}

# 获取服务运行时长
get_service_uptime() {
    local service=$1
    if systemctl is-active --quiet "$service" 2>/dev/null; then
        local start_time
        start_time=$(systemctl show --property ExecMainStartTimestamp --value "$service" 2>/dev/null)
        if [ -n "$start_time" ] && [ "$start_time" != "n/a" ]; then
            echo "$start_time"
        else
            echo "unknown"
        fi
    else
        echo "N/A"
    fi
}

# 打印详细状态信息
print_status() {
    print_banner
    
    echo -e "${CYAN}==================== 安装状态 ====================${NC}"
    echo ""
    
    if is_installed; then
        echo -e "${GREEN}● 安装状态:${NC} 已安装"
        local installed_version=$(get_installed_version)
        
        # 检查是否为预发布版本
        if is_prerelease_version; then
            echo -e "  安装版本: ${YELLOW}$installed_version${NC} ${RED}(预发布版本)${NC}"
        else
            echo -e "  安装版本: ${GREEN}$installed_version${NC}"
        fi
        
        echo -e "  安装时间: $(get_install_time)"
        echo -e "  安装目录: $INSTALL_DIR"
        echo -e "  数据目录: $DATA_DIR"
        echo -e "  日志目录: $LOG_DIR"
    else
        echo -e "${RED}○ 安装状态:${NC} 未安装"
        echo ""
        echo -e "使用以下命令安装:"
        echo -e "  sudo ./gpanel.sh install [version]"
        return 0
    fi
    
    echo ""
    echo -e "${CYAN}==================== 服务状态 ====================${NC}"
    echo ""
    
    # Agent 服务状态
    local agent_status=$(get_service_status "$AGENT_SERVICE")
    local gpanel_status=$(get_service_status "$GPANEL_SERVICE")
    
    case $agent_status in
        running)
            echo -e "${GREEN}● gpanel-agent:${NC} 运行中"
            echo -e "  PID: $(get_service_pid "$AGENT_SERVICE")"
            echo -e "  启动时间: $(get_service_uptime "$AGENT_SERVICE")"
            ;;
        stopped)
            echo -e "${YELLOW}○ gpanel-agent:${NC} 已停止"
            ;;
        not_installed)
            echo -e "${RED}○ gpanel-agent:${NC} 服务未安装"
            ;;
    esac
    
    echo ""
    
    case $gpanel_status in
        running)
            echo -e "${GREEN}● gpanel:${NC} 运行中"
            echo -e "  PID: $(get_service_pid "$GPANEL_SERVICE")"
            echo -e "  启动时间: $(get_service_uptime "$GPANEL_SERVICE")"
            ;;
        stopped)
            echo -e "${YELLOW}○ gpanel:${NC} 已停止"
            ;;
        not_installed)
            echo -e "${RED}○ gpanel:${NC} 服务未安装"
            ;;
    esac
    
    echo ""
    
    # 获取实际端口
    local core_port=$(get_core_port)
    local agent_port=$(get_agent_port)
    local security_entrance=$(get_security_entrance)
    
    # 端口监听状态
    echo -e "${CYAN}==================== 端口状态 ====================${NC}"
    echo ""
    
    if command -v ss >/dev/null 2>&1; then
        local port_core=$(ss -tlnp 2>/dev/null | grep ":${core_port}" || true)
        local port_agent=$(ss -tlnp 2>/dev/null | grep ":${agent_port}" || true)
        
        if [ -n "$port_core" ]; then
            echo -e "${GREEN}● 端口 ${core_port} (Core):${NC} 监听中"
        else
            echo -e "${RED}○ 端口 ${core_port} (Core):${NC} 未监听"
        fi
        
        if [ -n "$port_agent" ]; then
            echo -e "${GREEN}● 端口 ${agent_port} (Agent):${NC} 监听中"
        else
            echo -e "${RED}○ 端口 ${agent_port} (Agent):${NC} 未监听"
        fi
    else
        echo -e "${YELLOW}无法检测端口状态 (ss 命令不可用)${NC}"
    fi
    
    echo ""
    
    # 访问地址
    if [ "$gpanel_status" = "running" ]; then
        local SERVER_IP=$(hostname -I 2>/dev/null | awk '{print $1}' || echo "your-server-ip")
        
        # 获取本地 IP（内网地址）
        local LOCAL_IP=$(ip route get 1 2>/dev/null | awk '{for(i=1;i<=NF;i++) if($i=="src") print $(i+1)}')
        if [ -z "$LOCAL_IP" ]; then
            LOCAL_IP=$(hostname -I 2>/dev/null | tr ' ' '\n' | grep -E '^(192\.168\.|10\.|172\.(1[6-9]|2[0-9]|3[01])\.)' | head -1)
        fi
        if [ -z "$LOCAL_IP" ]; then
            LOCAL_IP="$SERVER_IP"
        fi
        
        echo -e "${CYAN}==================== 访问地址 ====================${NC}"
        echo ""
        
        # 根据安全入口显示不同的访问地址
        if [ "$security_entrance" != "/" ]; then
            echo -e "  安全入口: ${GREEN}${security_entrance}${NC}"
            echo ""
            echo -e "  本地访问: ${GREEN}http://${LOCAL_IP}:${core_port}${security_entrance}${NC}"
            echo -e "  外部访问: ${GREEN}http://${SERVER_IP}:${core_port}${security_entrance}${NC}"
        else
            echo -e "  本地访问: ${GREEN}http://${LOCAL_IP}:${core_port}${NC}"
            echo -e "  外部访问: ${GREEN}http://${SERVER_IP}:${core_port}${NC}"
        fi
        echo ""
    fi
}

# ============================================================
# 版本获取函数
# ============================================================

get_latest_version() {
    log_info "获取最新版本..." >&2
    
    local api_url="${GITHUB_API}/${GITHUB_REPO}/releases/latest"
    local response
    
    # 下载函数（支持 wget 和 curl）
    local fetch_cmd=""
    if command -v wget >/dev/null 2>&1; then
        fetch_cmd="wget -q -T 30 -O -"
    elif command -v curl >/dev/null 2>&1; then
        fetch_cmd="curl -fsSL --connect-timeout 30 --max-time 60"
    else
        echo ""
        return 1
    fi
    
    # 尝试使用 GitHub API 获取最新版本
    if response=$($fetch_cmd "$api_url" 2>/dev/null); then
        local latest_version=$(echo "$response" | grep -oP '"tag_name"\s*:\s*"\K[^"]+')
        if [ -n "$latest_version" ]; then
            echo "$latest_version"
            return 0
        fi
    fi
    
    # 备用方案：解析 releases 页面
    local releases_url="${GITHUB_RELEASES}/${GITHUB_REPO}/releases"
    if response=$($fetch_cmd "$releases_url" 2>/dev/null); then
        local latest_version=$(echo "$response" | grep -oP '/releases/tag/\K[^"]+' | head -1)
        if [ -n "$latest_version" ]; then
            echo "$latest_version"
            return 0
        fi
    fi
    
    echo ""
    return 1
}

# 获取所有可用版本
get_available_versions() {
    local api_url="${GITHUB_API}/${GITHUB_REPO}/releases"
    local response
    
    # 支持 wget 和 curl
    local fetch_cmd=""
    if command -v wget >/dev/null 2>&1; then
        fetch_cmd="wget -q -T 30 -O -"
    elif command -v curl >/dev/null 2>&1; then
        fetch_cmd="curl -fsSL --connect-timeout 30 --max-time 60"
    else
        return 1
    fi
    
    if response=$($fetch_cmd "$api_url" 2>/dev/null); then
        echo "$response" | grep -oP '"tag_name"\s*:\s*"\K[^"]+'
        return 0
    fi
    
    return 1
}

# 获取预发布版本
get_prerelease_version() {
    log_info "获取预发布版本..." >&2
    
    local api_url="${GITHUB_API}/${GITHUB_REPO}/releases"
    local response
    
    # 支持 wget 和 curl
    local fetch_cmd=""
    if command -v wget >/dev/null 2>&1; then
        fetch_cmd="wget -q -T 30 -O -"
    elif command -v curl >/dev/null 2>&1; then
        fetch_cmd="curl -fsSL --connect-timeout 30 --max-time 60"
    else
        echo ""
        return 1
    fi
    
    # 尝试使用 GitHub API 获取预发布版本
    if response=$($fetch_cmd "$api_url" 2>/dev/null); then
        # 查找 pre-release 标记的版本
        local prerelease_version=$(echo "$response" | grep -oP '"tag_name"\s*:\s*"\K[^"]+' | grep "^pre-release-" | head -1)
        if [ -n "$prerelease_version" ]; then
            echo "$prerelease_version"
            return 0
        fi
    fi
    
    echo ""
    return 1
}

# 获取预发布版本信息
get_prerelease_info() {
    local version=$1
    local api_url="${GITHUB_API}/${GITHUB_REPO}/releases/tags/${version}"
    local response
    
    # 支持 wget 和 curl
    local fetch_cmd=""
    if command -v wget >/dev/null 2>&1; then
        fetch_cmd="wget -q -T 30 -O -"
    elif command -v curl >/dev/null 2>&1; then
        fetch_cmd="curl -fsSL --connect-timeout 30 --max-time 60"
    else
        return 1
    fi
    
    if response=$($fetch_cmd "$api_url" 2>/dev/null); then
        echo "$response"
        return 0
    fi
    
    return 1
}

# 获取可用版本数组
get_available_versions_array() {
    local versions
    mapfile -t versions < <(get_available_versions | head -10)
    echo "${versions[@]}"
}

# 验证版本是否存在
# 规范化版本号（添加 v 前缀，但 pre-release 版本除外）
normalize_version() {
    local version=$1
    
    # pre-release 版本格式为 pre-release-xxx，不需要 v 前缀
    if [[ "$version" =~ ^pre-release- ]]; then
        echo "$version"
    elif [[ ! "$version" =~ ^v ]]; then
        echo "v${version}"
    else
        echo "$version"
    fi
}

validate_version() {
    local version=$1
    
    # 规范化版本号
    version=$(normalize_version "$version")
    
    local download_url="${GITHUB_PROXY}${GITHUB_RELEASES}/${GITHUB_REPO}/releases/download/${version}/gpanel-linux-${ARCH}.tar.gz"
    
    # 支持 wget 和 curl
    if command -v wget >/dev/null 2>&1; then
        if wget --spider -q -T 30 "$download_url" 2>/dev/null; then
        return 0
    fi
    elif command -v curl >/dev/null 2>&1; then
        if curl --output /dev/null --silent --head --fail --connect-timeout 30 --max-time 60 "$download_url" 2>/dev/null; then
            return 0
        fi
    fi
    
    return 1
}

# ============================================================
# 安装信息管理
# ============================================================

save_install_info() {
    local version=$1
    local install_time=$(date '+%Y-%m-%d %H:%M:%S')
    
    cat > "$INSTALL_INFO_FILE" << EOF
version=$version
install_time=$install_time
install_dir=$INSTALL_DIR
data_dir=$DATA_DIR
log_dir=$LOG_DIR
arch=$ARCH
os=$OS
EOF
    
    chmod 600 "$INSTALL_INFO_FILE"
}

# ============================================================
# 安装函数
# ============================================================

# 下载文件（优先 wget，备选 curl）
# 参数: $1=URL, $2=输出文件名, $3=超时秒数(可选,默认60), $4=重试次数(可选,默认3)
download_file() {
    local url="$1"
    local output="$2"
    local timeout="${3:-60}"
    local retries="${4:-3}"
    local attempt=1
    
    while [ $attempt -le $retries ]; do
        if [ $attempt -gt 1 ]; then
            log_warn "重试下载 ($attempt/$retries)..."
            sleep 2
        fi
        
        # 优先使用 wget
        if command -v wget >/dev/null 2>&1; then
            log_info "使用 wget 下载..."
            if wget --timeout="$timeout" --tries=1 -q --show-progress -O "$output" "$url" 2>&1; then
                return 0
            fi
        # 备选 curl
        elif command -v curl >/dev/null 2>&1; then
            log_info "使用 curl 下载..."
            if curl -fsSL --progress-bar --connect-timeout "$timeout" --max-time $((timeout * 10)) -o "$output" "$url" 2>&1; then
                return 0
            fi
        else
            log_error "未找到 wget 或 curl，请先安装其中一个"
            return 1
        fi
        
        ((attempt++))
    done
    
    return 1
}

download_binaries() {
    local target_version=$1
    
    log_step "下载二进制文件 (版本: $target_version)..."
    
    # 检查下载工具
    if ! command -v wget >/dev/null 2>&1 && ! command -v curl >/dev/null 2>&1; then
        log_error "未找到 wget 或 curl，请先安装其中一个"
        log_info "安装方法: apt install wget 或 yum install wget"
        exit 1
    fi
    
    # 创建临时目录
    TEMP_DIR=$(mktemp -d)
    cd "$TEMP_DIR"
    
    # 构建下载文件名和 URL（使用 GitHub 代理加速）
    local archive_name="gpanel-linux-${ARCH}.tar.gz"
    local download_url="${GITHUB_PROXY}${GITHUB_RELEASES}/${GITHUB_REPO}/releases/download/${target_version}/${archive_name}"
    local checksum_url="${GITHUB_PROXY}${GITHUB_RELEASES}/${GITHUB_REPO}/releases/download/${target_version}/checksums.txt"
    
    log_info "下载地址: $download_url"
    
    # 下载压缩包（超时120秒，重试3次）
    if ! download_file "$download_url" "$archive_name" 120 3; then
        log_error "下载失败: $archive_name"
        log_error "请检查版本号是否正确: $target_version"
        log_error "或访问 https://github.com/${GITHUB_REPO}/releases 查看可用版本"
        exit 1
    fi
    
    log_info "下载完成: $archive_name"
    
    # 下载校验和文件（可选）
    if download_file "$checksum_url" "checksums.txt" 30 1 2>/dev/null; then
        log_info "验证文件完整性..."
        if command -v sha256sum >/dev/null 2>&1; then
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
    log_step "安装二进制文件..."
    
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

create_systemd_services() {
    log_step "创建 systemd 服务..."
    
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
    log_step "检测防火墙状态..."
    
    # 获取实际端口
    local core_port=$(get_core_port)
    local agent_port=$(get_agent_port)
    
    # 检测防火墙类型并保存到全局变量
    FIREWALL_TYPE="none"
    FIREWALL_AUTO_CONFIGURED=false
    FIREWALL_CORE_PORT="$core_port"
    FIREWALL_AGENT_PORT="$agent_port"
    
    if command -v firewall-cmd >/dev/null 2>&1; then
        FIREWALL_TYPE="firewalld"
        log_info "检测到 firewalld 防火墙"
        
        # 自动放行端口
        log_info "正在自动放行端口 ${core_port} 和 ${agent_port}..."
        if firewall-cmd --permanent --add-port=${core_port}/tcp 2>/dev/null; then
            firewall-cmd --permanent --add-port=${agent_port}/tcp 2>/dev/null || true
            firewall-cmd --reload 2>/dev/null || true
            FIREWALL_AUTO_CONFIGURED=true
            log_info "已自动放行端口 ${core_port} 和 ${agent_port}"
        else
            log_warn "自动放行失败，可能需要手动配置"
        fi
        
    elif command -v ufw >/dev/null 2>&1; then
        FIREWALL_TYPE="ufw"
        log_info "检测到 ufw 防火墙"
        
        # 自动放行端口
        log_info "正在自动放行端口 ${core_port} 和 ${agent_port}..."
        if ufw allow ${core_port}/tcp 2>/dev/null; then
            ufw allow ${agent_port}/tcp 2>/dev/null || true
            FIREWALL_AUTO_CONFIGURED=true
            log_info "已自动放行端口 ${core_port} 和 ${agent_port}"
        else
            log_warn "自动放行失败，可能需要手动配置"
        fi
        
    elif command -v iptables >/dev/null 2>&1; then
        FIREWALL_TYPE="iptables"
        log_info "检测到 iptables 防火墙"
        
        # 自动放行端口
        log_info "正在自动放行端口 ${core_port} 和 ${agent_port}..."
        if iptables -I INPUT -p tcp --dport ${core_port} -j ACCEPT 2>/dev/null; then
            iptables -I INPUT -p tcp --dport ${agent_port} -j ACCEPT 2>/dev/null || true
            FIREWALL_AUTO_CONFIGURED=true
            log_info "已自动放行端口 ${core_port} 和 ${agent_port}"
            # 尝试持久化规则
            if command -v iptables-save >/dev/null 2>&1; then
                mkdir -p /etc/iptables 2>/dev/null || true
                iptables-save > /etc/iptables/rules.v4 2>/dev/null || true
            fi
        else
            log_warn "自动放行失败，可能需要手动配置"
        fi
        
    else
        FIREWALL_TYPE="none"
        log_info "未检测到防火墙管理工具"
    fi
}

start_services() {
    log_step "启动服务..."
    
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

stop_services() {
    log_step "停止服务..."
    
    systemctl stop gpanel 2>/dev/null || true
    systemctl stop gpanel-agent 2>/dev/null || true
    
    log_info "服务已停止"
}

cleanup_temp() {
    if [ -d "$TEMP_DIR" ]; then
        rm -rf "$TEMP_DIR"
    fi
}

print_install_success() {
    echo ""
    echo -e "${GREEN}========================================${NC}"
    echo -e "${GREEN}安装成功！${NC}"
    echo -e "${GREEN}========================================${NC}"
    echo ""
    echo -e "${GREEN}安装版本:${NC} $VERSION"
    echo ""
    
    # 使用 gpctl 显示连接信息
    if command -v gpctl >/dev/null 2>&1; then
        gpctl init-security --show
    else
        # 备用方案：手动获取 IP 和端口
        local core_port=$(get_core_port)
        local security_entrance=$(get_security_entrance)
        
        SERVER_IP=$(hostname -I 2>/dev/null | awk '{print $1}')
        if [ -z "$SERVER_IP" ]; then
            SERVER_IP="your-server-ip"
        fi
        
        LOCAL_IP=$(ip route get 1 2>/dev/null | awk '{for(i=1;i<=NF;i++) if($i=="src") print $(i+1)}')
        if [ -z "$LOCAL_IP" ]; then
            LOCAL_IP=$(hostname -I 2>/dev/null | tr ' ' '\n' | grep -E '^(192\.168\.|10\.|172\.(1[6-9]|2[0-9]|3[01])\.)' | head -1)
        fi
        if [ -z "$LOCAL_IP" ]; then
            LOCAL_IP="$SERVER_IP"
        fi
        
        echo -e "${GREEN}访问地址:${NC}"
        if [ "$security_entrance" != "/" ]; then
            echo "  安全入口: ${security_entrance}"
            echo "  本地访问: http://${LOCAL_IP}:${core_port}${security_entrance}"
            echo "  外部访问: http://${SERVER_IP}:${core_port}${security_entrance}"
        else
            echo "  本地访问: http://${LOCAL_IP}:${core_port}"
            echo "  外部访问: http://${SERVER_IP}:${core_port}"
        fi
    fi
    
    # 显示账号密码
    echo ""
    echo -e "${GREEN}登录信息:${NC}"
    if [ -n "$RESET_USERNAME" ] && [ -n "$RESET_PASSWORD" ]; then
        echo -e "  用户名: ${CYAN}${RESET_USERNAME}${NC}"
        echo -e "  密码: ${CYAN}${RESET_PASSWORD}${NC}"
    elif get_db_user_info; then
        echo -e "  用户名: ${CYAN}${DB_USERNAME}${NC}"
        echo -e "  密码: ${CYAN}${DB_PASSWORD}${NC}"
    else
        echo -e "  用户名: ${CYAN}admin${NC}"
        echo -e "  密码: ${CYAN}admin123${NC}"
    fi
    
    echo ""
    echo -e "${GREEN}服务管理:${NC}"
    echo "  sudo ./gpanel.sh update    # 更新版本"
    echo "  sudo ./gpanel.sh uninstall # 卸载"
    echo "  sudo gpctl status          # 使用 gpctl 管理"
    echo ""
    echo -e "${GREEN}查看日志:${NC}"
    echo "  sudo journalctl -u gpanel -u gpanel-agent -f"
    
    # 防火墙端口提示
    if [ -n "$FIREWALL_TYPE" ] && [ "$FIREWALL_TYPE" != "none" ]; then
        echo ""
        echo -e "${YELLOW}========================================${NC}"
        echo -e "${YELLOW}防火墙提示${NC}"
        echo -e "${YELLOW}========================================${NC}"
        echo ""
        
        local core_port="${FIREWALL_CORE_PORT:-8080}"
        local agent_port="${FIREWALL_AGENT_PORT:-9998}"
        
        if [ "$FIREWALL_AUTO_CONFIGURED" = true ]; then
            echo -e "${GREEN}已自动在 ${FIREWALL_TYPE} 防火墙放行以下端口:${NC}"
            echo ""
            echo -e "  ${GREEN}●${NC} ${core_port} (Core Web 服务)"
            echo -e "  ${GREEN}●${NC} ${agent_port} (Agent 服务)"
            echo ""
            if [ "$FIREWALL_TYPE" = "iptables" ]; then
                echo -e "${YELLOW}注意: iptables 规则已保存到 /etc/iptables/rules.v4${NC}"
            fi
        else
            echo -e "${YELLOW}检测到系统已安装 ${FIREWALL_TYPE} 防火墙，请手动放行以下端口:${NC}"
            echo ""
            
            case "$FIREWALL_TYPE" in
                firewalld)
                    echo -e "  ${CYAN}firewalld 放行命令:${NC}"
                    echo "    sudo firewall-cmd --permanent --add-port=${core_port}/tcp"
                    echo "    sudo firewall-cmd --permanent --add-port=${agent_port}/tcp"
                    echo "    sudo firewall-cmd --reload"
                    ;;
                ufw)
                    echo -e "  ${CYAN}ufw 放行命令:${NC}"
                    echo "    sudo ufw allow ${core_port}/tcp"
                    echo "    sudo ufw allow ${agent_port}/tcp"
                    ;;
                iptables)
                    echo -e "  ${CYAN}iptables 放行命令:${NC}"
                    echo "    sudo iptables -I INPUT -p tcp --dport ${core_port} -j ACCEPT"
                    echo "    sudo iptables -I INPUT -p tcp --dport ${agent_port} -j ACCEPT"
                    echo "    sudo iptables-save > /etc/iptables/rules.v4  # 持久化规则"
                    ;;
            esac
            echo ""
            echo -e "${YELLOW}需要放行的端口:${NC}"
            echo "  - ${core_port} (Core Web 服务)"
            echo "  - ${agent_port} (Agent 服务)"
        fi
        echo ""
    fi
}

# ============================================================
# 命令处理函数
# ============================================================

do_install_pre() {
    print_banner
    
    # 显示预发布版本警告
    echo -e "${YELLOW}========================================${NC}"
    echo -e "${YELLOW}    预发布版本安装警告${NC}"
    echo -e "${YELLOW}========================================${NC}"
    echo ""
    echo -e "${RED}警告: 您即将安装预发布版本！${NC}"
    echo ""
    echo -e "${YELLOW}预发布版本可能存在以下问题:${NC}"
    echo -e "  ${RED}●${NC} 可能包含未完成的功能"
    echo -e "  ${RED}●${NC} 可能存在未发现的 Bug"
    echo -e "  ${RED}●${NC} 可能存在稳定性问题"
    echo -e "  ${RED}●${NC} 可能与正式版不兼容"
    echo -e "  ${RED}●${NC} 不建议在生产环境中使用"
    echo ""
    echo -e "${GREEN}预发布版本包含:${NC}"
    echo -e "  ${GREEN}●${NC} 最新的功能特性"
    echo -e "  ${GREEN}●${NC} 最新的 Bug 修复"
    echo -e "  ${GREEN}●${NC} 开发中的新功能预览"
    echo ""
    
    if ! confirm "确认安装预发布版本?"; then
        log_info "安装已取消"
        exit 0
    fi
    
    # 检查是否已安装
    if is_installed; then
        local installed_version=$(get_installed_version)
        log_warn "GPanel 已安装 (当前版本: $installed_version)"
        echo ""
        
        # 检查是否已经是预发布版本
        if [[ "$installed_version" =~ ^pre-release- ]]; then
            log_info "当前已安装预发布版本"
            if ! confirm "是否更新到最新的预发布版本?"; then
                log_info "安装已取消"
                exit 0
            fi
        else
            if ! confirm "是否覆盖当前安装的正式版本?"; then
                log_info "安装已取消"
                exit 0
            fi
        fi
    fi
    
    # 检查操作系统和架构
    check_os
    check_arch
    
    # 检查依赖
    log_info "检查依赖..."
    if ! command -v curl >/dev/null 2>&1; then
        log_error "未安装 curl，请先安装 curl"
        exit 1
    fi
    
    # 获取预发布版本
    local prerelease_version=$(get_prerelease_version)
    
    if [ -z "$prerelease_version" ]; then
        log_error "未找到预发布版本"
        log_info "请确保仓库中有预发布版本，或者使用 'sudo ./gpanel.sh install' 安装正式版本"
        exit 1
    fi
    
    VERSION="$prerelease_version"
    
    # 获取预发布版本信息
    local release_info=$(get_prerelease_info "$VERSION")
    local build_time=""
    local commit_short=""
    local commit_msg=""
    
    if [ -n "$release_info" ]; then
        # 尝试从 release body 中解析信息
        build_time=$(echo "$release_info" | grep -oP '\*\*构建时间\*\*\s*\|\s*\K[^|]+' | xargs || echo "未知")
        commit_short=$(echo "$release_info" | grep -oP '\*\*Commit\*\*\s*\|\s*`\K[^`]+' || echo "未知")
        commit_msg=$(echo "$release_info" | grep -oP '\*\*Commit Message\*\*\s*\|\s*\K.*' | xargs || echo "未知")
    fi
    
    # 显示安装信息
    echo ""
    echo -e "${CYAN}==================== 预发布版本信息 ====================${NC}"
    echo ""
    echo -e "  版本: ${GREEN}$VERSION${NC}"
    if [ -n "$build_time" ] && [ "$build_time" != "未知" ]; then
        echo -e "  构建时间: ${GREEN}$build_time${NC}"
    fi
    if [ -n "$commit_short" ] && [ "$commit_short" != "未知" ]; then
        echo -e "  Commit: ${GREEN}$commit_short${NC}"
    fi
    if [ -n "$commit_msg" ] && [ "$commit_msg" != "未知" ]; then
        echo -e "  Commit Message: ${GREEN}$commit_msg${NC}"
    fi
    echo -e "  架构: ${GREEN}$ARCH${NC}"
    echo -e "  安装目录: ${GREEN}$INSTALL_DIR${NC}"
    echo -e "  数据目录: ${GREEN}$DATA_DIR${NC}"
    echo -e "  日志目录: ${GREEN}$LOG_DIR${NC}"
    echo ""
    echo -e "${YELLOW}注意: 这是一个预发布版本，可能不稳定！${NC}"
    echo ""
    
    if ! confirm "确认开始安装预发布版本?"; then
        log_info "安装已取消"
        exit 0
    fi
    
    # 停止服务（避免覆盖正在运行的二进制文件）
    stop_services
    
    # 下载二进制文件
    download_binaries "$VERSION"
    
    # 安装二进制文件
    install_binaries
    
    # 创建 systemd 服务
    create_systemd_services
    
    # 检测数据库是否存在
    local DB_EXISTS=false
    if [ -f "$DATA_DIR/gpanel.db" ]; then
        DB_EXISTS=true
        log_info "检测到已存在的数据库，保留原有配置"
    fi
    
    # 启动服务
    start_services
    
    # 检测是否为全新安装
    if [ "$DB_EXISTS" = false ]; then
        log_step "检测到全新安装，正在初始化安全配置..."
        
        log_info "等待服务初始化..."
        sleep 5
        
        local wait_count=0
        while [ ! -f "$DATA_DIR/gpanel.db" ] && [ $wait_count -lt 30 ]; do
            sleep 1
            ((wait_count++))
        done
        
        if [ -f "$DATA_DIR/gpanel.db" ]; then
            log_info "数据库已初始化"
            
            if command -v gpctl >/dev/null 2>&1; then
                log_info "正在生成随机端口和安全入口..."
                gpctl init-security
                
                log_info "正在重置管理员密码..."
                RESET_OUTPUT=$(gpctl reset password --quiet 2>/dev/null)
                if [ -n "$RESET_OUTPUT" ]; then
                    RESET_USERNAME=$(echo "$RESET_OUTPUT" | awk '{print $1}')
                    RESET_PASSWORD=$(echo "$RESET_OUTPUT" | awk '{print $2}')
                    log_info "管理员密码已重置"
                fi
            else
                log_warn "gpctl 命令不可用，跳过安全配置初始化"
            fi
        else
            log_warn "数据库初始化超时，跳过安全配置"
        fi
    fi
    
    # 配置防火墙
    configure_firewall
    
    # 保存安装信息，标记为预发布版本
    save_install_info "$VERSION"
    echo "is_prerelease=true" >> "$INSTALL_INFO_FILE"
    
    # 清理临时文件
    cleanup_temp
    
    # 显示成功信息
    print_install_pre_success
}

print_install_pre_success() {
    echo ""
    echo -e "${GREEN}========================================${NC}"
    echo -e "${GREEN}预发布版本安装成功！${NC}"
    echo -e "${GREEN}========================================${NC}"
    echo ""
    echo -e "${GREEN}安装版本:${NC} $VERSION ${YELLOW}(预发布版本)${NC}"
    echo ""
    
    # 使用 gpctl 显示连接信息
    if command -v gpctl >/dev/null 2>&1; then
        gpctl init-security --show
    else
        # 备用方案
        local core_port=$(get_core_port)
        local security_entrance=$(get_security_entrance)
        
        SERVER_IP=$(hostname -I 2>/dev/null | awk '{print $1}')
        if [ -z "$SERVER_IP" ]; then
            SERVER_IP="your-server-ip"
        fi
        
        LOCAL_IP=$(ip route get 1 2>/dev/null | awk '{for(i=1;i<=NF;i++) if($i=="src") print $(i+1)}')
        if [ -z "$LOCAL_IP" ]; then
            LOCAL_IP=$(hostname -I 2>/dev/null | tr ' ' '\n' | grep -E '^(192\.168\.|10\.|172\.(1[6-9]|2[0-9]|3[01])\.)' | head -1)
        fi
        if [ -z "$LOCAL_IP" ]; then
            LOCAL_IP="$SERVER_IP"
        fi
        
        echo -e "${GREEN}访问地址:${NC}"
        if [ "$security_entrance" != "/" ]; then
            echo "  安全入口: ${security_entrance}"
            echo "  本地访问: http://${LOCAL_IP}:${core_port}${security_entrance}"
            echo "  外部访问: http://${SERVER_IP}:${core_port}${security_entrance}"
        else
            echo "  本地访问: http://${LOCAL_IP}:${core_port}"
            echo "  外部访问: http://${SERVER_IP}:${core_port}"
        fi
    fi
    
    # 显示账号密码
    echo ""
    echo -e "${GREEN}登录信息:${NC}"
    if [ -n "$RESET_USERNAME" ] && [ -n "$RESET_PASSWORD" ]; then
        echo -e "  用户名: ${CYAN}${RESET_USERNAME}${NC}"
        echo -e "  密码: ${CYAN}${RESET_PASSWORD}${NC}"
    elif get_db_user_info; then
        echo -e "  用户名: ${CYAN}${DB_USERNAME}${NC}"
        echo -e "  密码: ${CYAN}${DB_PASSWORD}${NC}"
    else
        echo -e "  用户名: ${CYAN}admin${NC}"
        echo -e "  密码: ${CYAN}admin123${NC}"
    fi
    
    echo ""
    echo -e "${YELLOW}========================================${NC}"
    echo -e "${YELLOW}预发布版本注意事项:${NC}"
    echo -e "${YELLOW}========================================${NC}"
    echo ""
    echo -e "  ${RED}●${NC} 此版本可能不稳定，请谨慎使用"
    echo -e "  ${RED}●${NC} 如遇问题，可通过 'sudo ./gpanel.sh uninstall' 卸载"
    echo -e "  ${RED}●${NC} 建议安装正式版: 'sudo ./gpanel.sh install'"
    echo ""
    echo -e "${GREEN}服务管理:${NC}"
    echo "  sudo ./gpanel.sh update        # 更��到正式版本"
    echo "  sudo ./gpanel.sh install-pre   # 更新到最新预发布版本"
    echo "  sudo ./gpanel.sh uninstall     # 卸载"
    echo ""
    echo -e "${GREEN}查看日志:${NC}"
    echo "  sudo journalctl -u gpanel -u gpanel-agent -f"
    
    # 防火墙端口提示
    if [ -n "$FIREWALL_TYPE" ] && [ "$FIREWALL_TYPE" != "none" ]; then
        echo ""
        echo -e "${YELLOW}========================================${NC}"
        echo -e "${YELLOW}防火墙提示${NC}"
        echo -e "${YELLOW}========================================${NC}"
        echo ""
        
        local core_port="${FIREWALL_CORE_PORT:-8080}"
        local agent_port="${FIREWALL_AGENT_PORT:-9998}"
        
        if [ "$FIREWALL_AUTO_CONFIGURED" = true ]; then
            echo -e "${GREEN}已自动在 ${FIREWALL_TYPE} 防火墙放行以下端口:${NC}"
            echo ""
            echo -e "  ${GREEN}●${NC} ${core_port} (Core Web 服务)"
            echo -e "  ${GREEN}●${NC} ${agent_port} (Agent 服务)"
            echo ""
            if [ "$FIREWALL_TYPE" = "iptables" ]; then
                echo -e "${YELLOW}注意: iptables 规则已保存到 /etc/iptables/rules.v4${NC}"
            fi
        else
            echo -e "${YELLOW}检测到系统已安装 ${FIREWALL_TYPE} 防火墙，请手动放行以下端口:${NC}"
            echo ""
            
            case "$FIREWALL_TYPE" in
                firewalld)
                    echo -e "  ${CYAN}firewalld 放行命令:${NC}"
                    echo "    sudo firewall-cmd --permanent --add-port=${core_port}/tcp"
                    echo "    sudo firewall-cmd --permanent --add-port=${agent_port}/tcp"
                    echo "    sudo firewall-cmd --reload"
                    ;;
                ufw)
                    echo -e "  ${CYAN}ufw 放行命令:${NC}"
                    echo "    sudo ufw allow ${core_port}/tcp"
                    echo "    sudo ufw allow ${agent_port}/tcp"
                    ;;
                iptables)
                    echo -e "  ${CYAN}iptables 放行命令:${NC}"
                    echo "    sudo iptables -I INPUT -p tcp --dport ${core_port} -j ACCEPT"
                    echo "    sudo iptables -I INPUT -p tcp --dport ${agent_port} -j ACCEPT"
                    echo "    sudo iptables-save > /etc/iptables/rules.v4  # 持久化规则"
                    ;;
            esac
            echo ""
            echo -e "${YELLOW}需要放行的端口:${NC}"
            echo "  - ${core_port} (Core Web 服务)"
            echo "  - ${agent_port} (Agent 服务)"
        fi
        echo ""
    fi
}

do_install() {
    local target_version="${1:-}"
    
    print_banner
    
    # 检查是否已安装
    if is_installed; then
        local installed_version=$(get_installed_version)
        log_warn "GPanel 已安装 (当前版本: $installed_version)"
        echo ""
        
        select_option "请选择操作:" "更新到新版本" "卸载后重新安装"
        
        # 用户选择退出
        if [ $? -ne 0 ]; then
            log_info "操作已取消"
            return 0
        fi
        
        case "$SELECTED_OPTION" in
            "更新到新版本")
                do_update
                return $?
                ;;
            "卸载后重新安装")
                do_uninstall
                if [ $? -ne 0 ]; then
                    return 1
                fi
                echo ""
                log_info "继续安装..."
                ;;
        esac
    fi
    
    # 检查操作系统和架构
    check_os
    check_arch
    
    # 检查依赖
    log_info "检查依赖..."
    if ! command -v curl >/dev/null 2>&1; then
        log_error "未安装 curl，请先安装 curl"
        exit 1
    fi
    
    # 确定安装版本
    if [ -n "$target_version" ]; then
        # 规范化版本号
        target_version=$(normalize_version "$target_version")
        
        # 验证版本是否存在
        log_info "验证版本 $target_version ..."
        if ! validate_version "$target_version"; then
            log_error "版本 $target_version 不存在"
            echo ""
            log_info "获取可用版本列表..."
            local versions
            mapfile -t versions < <(get_available_versions | head -10)
            
            if [ ${#versions[@]} -eq 0 ]; then
                log_error "无法获取版本列表"
                exit 1
            fi
            
            select_version "请选择要安装的版本:" "${versions[@]}"
            if [ $? -ne 0 ]; then
                log_info "安装已取消"
                exit 0
            fi
            target_version="$SELECTED_VERSION"
        fi
        VERSION="$target_version"
    else
        # 获取最新版本
        local latest_version=$(get_latest_version)
        
        if [ -z "$latest_version" ]; then
            log_error "无法获取最新版本"
            echo ""
            log_info "获取可用版本列表..."
            local versions
            mapfile -t versions < <(get_available_versions | head -10)
            
            if [ ${#versions[@]} -eq 0 ]; then
                log_error "无法获取版本列表，请手动指定版本"
                log_info "使用方法: sudo ./gpanel.sh install v1.0.0"
                exit 1
            fi
            
            select_version "请选择要安装的版本:" "${versions[@]}"
            if [ $? -ne 0 ]; then
                log_info "安装已取消"
                exit 0
            fi
            VERSION="$SELECTED_VERSION"
        else
            echo ""
            echo -e "${CYAN}可用版本:${NC}"
            echo -e "  ${GREEN}1)${NC} $latest_version (最新版)"
            echo -e "  ${GREEN}2)${NC} 选择其他版本"
            echo -e "  ${GREEN}3)${NC} 手动输入版本号"
            echo -e "  ${GREEN}0)${NC} 退出"
            echo ""
            
            local choice
            echo -ne "${YELLOW}请选择 [0-3]: ${NC}"
            if [ "$INTERACTIVE" = true ]; then
                read -r choice < "$TTY_DEVICE"
            else
                choice=1
                echo "$choice"
            fi
            
            case "$choice" in
                0)
                    log_info "安装已取消"
                    exit 0
                    ;;
                1)
                    VERSION="$latest_version"
                    ;;
                2)
                    log_info "获取可用版本列表..."
                    local versions
                    mapfile -t versions < <(get_available_versions | head -10)
                    
                    if [ ${#versions[@]} -eq 0 ]; then
                        log_error "无法获取版本列表"
                        exit 1
                    fi
                    
                    select_version "请选择要安装的版本:" "${versions[@]}"
                    if [ $? -ne 0 ]; then
                        log_info "安装已取消"
                        exit 0
                    fi
                    VERSION="$SELECTED_VERSION"
                    ;;
                3)
                    echo -ne "${YELLOW}请输入版本号 (如 v1.0.0): ${NC}"
                    if [ "$INTERACTIVE" = true ]; then
                        read -r VERSION < "$TTY_DEVICE"
                    else
                        VERSION="$latest_version"
                        echo "$VERSION"
                    fi
                    
                    VERSION=$(normalize_version "$VERSION")
                    ;;
                *)
                    log_info "使用最新版本: $latest_version"
                    VERSION="$latest_version"
                    ;;
            esac
        fi
    fi
    
    # 验证最终版本
    VERSION=$(normalize_version "$VERSION")
    
    if ! validate_version "$VERSION"; then
        log_error "版本 $VERSION 不存在"
        exit 1
    fi
    
    log_info "安装版本: $VERSION"
    
    # 确认安装
    echo ""
    echo -e "${CYAN}==================== 安装信息 ====================${NC}"
    echo ""
    echo -e "  版本: ${GREEN}$VERSION${NC}"
    echo -e "  架构: ${GREEN}$ARCH${NC}"
    echo -e "  安装目录: ${GREEN}$INSTALL_DIR${NC}"
    echo -e "  数据目录: ${GREEN}$DATA_DIR${NC}"
    echo -e "  日志目录: ${GREEN}$LOG_DIR${NC}"
    echo ""
    
    if ! confirm "确认开始安装?"; then
        log_info "安装已取消"
        exit 0
    fi
    
    # 下载二进制文件
    download_binaries "$VERSION"
    
    # 安装二进制文件
    install_binaries
    
    # 创建 systemd 服务
    create_systemd_services
    
    # 在启动服务之前检测数据库是否存在（避免服务启动后自动初始化数据库导致检测失效）
    local DB_EXISTS=false
    if [ -f "$DATA_DIR/gpanel.db" ]; then
        DB_EXISTS=true
        log_info "检测到已存在的数据库，保留原有配置"
    fi
    
    # 启动服务
    start_services
    
    # 检测是否为全新安装（数据库不存在），如果是则初始化安全配置
    if [ "$DB_EXISTS" = false ]; then
        log_step "检测到全新安装，正在初始化安全配置..."
        
        # 等待服务完全启动并生成数据库
        log_info "等待服务初始化..."
        sleep 5
        
        # 检查数据库是否生成
        local wait_count=0
        while [ ! -f "$DATA_DIR/gpanel.db" ] && [ $wait_count -lt 30 ]; do
            sleep 1
            ((wait_count++))
        done
        
        if [ -f "$DATA_DIR/gpanel.db" ]; then
            log_info "数据库已初始化"
            
            # 调用 gpctl 初始化安全配置（随机端口和安全入口）
            if command -v gpctl >/dev/null 2>&1; then
                log_info "正在生成随机端口和安全入口..."
                gpctl init-security
                
                # 重置 admin 密码并保存
                log_info "正在重置管理员密码..."
                RESET_OUTPUT=$(gpctl reset password --quiet 2>/dev/null)
                if [ -n "$RESET_OUTPUT" ]; then
                    # 格式: username password
                    RESET_USERNAME=$(echo "$RESET_OUTPUT" | awk '{print $1}')
                    RESET_PASSWORD=$(echo "$RESET_OUTPUT" | awk '{print $2}')
                    log_info "管理员密码已重置"
                fi
            else
                log_warn "gpctl 命令不可用，跳过安全配置初始化"
            fi
        else
            log_warn "数据库初始化超时，跳过安全配置"
        fi
    fi
    
    # 配置防火墙（在安全配置之后，使用实际端口）
    configure_firewall
    
    # 保存安装信息
    save_install_info "$VERSION"
    
    # 清理临时文件
    cleanup_temp
    
    # 显示成功信息
    print_install_success
}

do_update() {
    local target_version="${1:-}"
    
    print_banner
    
    # 检查是否已安装
    if ! is_installed; then
        log_error "GPanel 未安装"
        log_info "请先安装: sudo ./gpanel.sh install [version]"
        exit 1
    fi
    
    local current_version=$(get_installed_version)
    log_info "当前版本: $current_version"
    
    # 检查操作系统和架构
    check_os
    check_arch
    
    # 确定目标版本
    if [ -n "$target_version" ]; then
        # 规范化版本号
        target_version=$(normalize_version "$target_version")
        VERSION="$target_version"
    else
        # 获取最新版本
        local latest_version=$(get_latest_version)
        
        if [ -z "$latest_version" ]; then
            log_error "无法获取最新版本"
            echo ""
            log_info "获取可用版本列表..."
            local versions
            mapfile -t versions < <(get_available_versions | head -10)
            
            if [ ${#versions[@]} -eq 0 ]; then
                log_error "无法获取版本列表，请手动指定版本"
                log_info "使用方法: sudo ./gpanel.sh update v1.0.0"
                exit 1
            fi
            
            select_version "请选择要更新的版本:" "${versions[@]}"
            if [ $? -ne 0 ]; then
                log_info "更新已取消"
                exit 0
            fi
            VERSION="$SELECTED_VERSION"
        else
            echo ""
            echo -e "${CYAN}可用版本:${NC}"
            echo -e "  ${GREEN}1)${NC} $latest_version (最新版)"
            echo -e "  ${GREEN}2)${NC} 选择其他版本"
            echo -e "  ${GREEN}3)${NC} 手动输入版本号"
            echo -e "  ${GREEN}0)${NC} 退出"
            echo ""
            
            local choice
            echo -ne "${YELLOW}请选择 [0-3]: ${NC}"
            if [ "$INTERACTIVE" = true ]; then
                read -r choice < "$TTY_DEVICE"
            else
                choice=1
                echo "$choice"
            fi
            
            case "$choice" in
                0)
                    log_info "更新已取消"
                    exit 0
                    ;;
                1)
                    VERSION="$latest_version"
                    ;;
                2)
                    log_info "获取可用版本列表..."
                    local versions
                    mapfile -t versions < <(get_available_versions | head -10)
                    
                    if [ ${#versions[@]} -eq 0 ]; then
                        log_error "无法获取版本列表"
                        exit 1
                    fi
                    
                    select_version "请选择要更新的版本:" "${versions[@]}"
                    if [ $? -ne 0 ]; then
                        log_info "更新已取消"
                        exit 0
                    fi
                    VERSION="$SELECTED_VERSION"
                    ;;
                3)
                    echo -ne "${YELLOW}请输入版本号 (如 v1.0.0): ${NC}"
                    if [ "$INTERACTIVE" = true ]; then
                        read -r VERSION < "$TTY_DEVICE"
                    else
                        VERSION="$latest_version"
                        echo "$VERSION"
                    fi
                    
                    VERSION=$(normalize_version "$VERSION")
                    ;;
                *)
                    log_info "使用最新版本: $latest_version"
                    VERSION="$latest_version"
                    ;;
            esac
        fi
    fi
    
    # 检查是否已是最新版本
    if [ "$VERSION" = "$current_version" ]; then
        log_info "当前已是最新版本: $VERSION"
        
        if ! confirm "是否强制重新安装当前版本?"; then
            return 0
        fi
    fi
    
    log_info "更新到版本: $VERSION"
    
    # 验证版本是否存在
    if ! validate_version "$VERSION"; then
        log_error "版本 $VERSION 不存在"
        echo ""
        log_info "获取可用版本列表..."
        local versions
        mapfile -t versions < <(get_available_versions | head -10)
        
        if [ ${#versions[@]} -gt 0 ]; then
            select_version "请选择要更新的版本:" "${versions[@]}"
            if [ $? -ne 0 ]; then
                log_info "更新已取消"
                exit 0
            fi
            VERSION="$SELECTED_VERSION"
        else
            exit 1
        fi
    fi
    
    # 确认更新
    echo ""
    echo -e "${CYAN}==================== 更新信息 ====================${NC}"
    echo ""
    echo -e "  当前版本: ${YELLOW}$current_version${NC}"
    echo -e "  目标版本: ${GREEN}$VERSION${NC}"
    echo -e "  架构: ${GREEN}$ARCH${NC}"
    echo ""
    
    if ! confirm "确认开始更新?"; then
        log_info "更新已取消"
        exit 0
    fi
    
    # 停止服务
    stop_services
    
    # 下载新版本
    download_binaries "$VERSION"
    
    # 安装二进制文件
    install_binaries
    
    # 保存安装信息
    save_install_info "$VERSION"
    
    # 清理临时文件
    cleanup_temp
    
    # 启动服务
    start_services
    
    echo ""
    echo -e "${GREEN}========================================${NC}"
    echo -e "${GREEN}更新成功！${NC}"
    echo -e "${GREEN}========================================${NC}"
    echo ""
    echo -e "${GREEN}旧版本:${NC} $current_version"
    echo -e "${GREEN}新版本:${NC} $VERSION"
    echo ""
    
    # 显示连接信息
    if command -v gpctl >/dev/null 2>&1; then
        gpctl init-security --show
    fi
}

do_uninstall() {
    print_banner
    
    # 检查是否已安装
    if ! is_installed; then
        log_warn "GPanel 未安装"
        return 0
    fi
    
    local installed_version=$(get_installed_version)
    log_info "当前安装版本: $installed_version"
    
    # 显示安装信息
    echo ""
    echo -e "${CYAN}==================== 卸载信息 ====================${NC}"
    echo ""
    echo -e "  版本: ${YELLOW}$installed_version${NC}"
    echo -e "  安装目录: ${YELLOW}$INSTALL_DIR${NC}"
    echo -e "  数据目录: ${YELLOW}$DATA_DIR${NC}"
    echo -e "  日志目录: ${YELLOW}$LOG_DIR${NC}"
    echo ""
    
    # 确认卸载
    if ! confirm "确定要卸载 GPanel 吗?"; then
        log_info "卸载已取消"
        return 1
    fi
    
    log_warn "开始卸载 GPanel..."
    
    # 停止服务
    stop_services
    
    # 禁用并删除服务
    log_step "删除 systemd 服务..."
    systemctl disable gpanel 2>/dev/null || true
    systemctl disable gpanel-agent 2>/dev/null || true
    rm -f /etc/systemd/system/gpanel.service
    rm -f /etc/systemd/system/gpanel-agent.service
    systemctl daemon-reload
    log_info "systemd 服务已删除"
    
    # 删除二进制文件
    log_step "删除程序文件..."
    rm -rf "$INSTALL_DIR"
    rm -f /usr/local/bin/gpctl
    log_info "程序文件已删除"
    
    # 询问是否删除数据和日志
    echo ""
    if confirm "是否同时删除数据目录和日志目录?"; then
        log_step "删除数据和日志..."
        rm -rf "$DATA_DIR"
        rm -rf "$LOG_DIR"
        log_info "数据和日志已删除"
    else
        log_info "数据和日志已保留"
        log_info "数据目录: $DATA_DIR"
        log_info "日志目录: $LOG_DIR"
    fi
    
    echo ""
    echo -e "${GREEN}========================================${NC}"
    echo -e "${GREEN}卸载完成！${NC}"
    echo -e "${GREEN}========================================${NC}"
    echo ""
    
    return 0
}

do_status() {
    print_status
}

do_help() {
    print_banner
    echo -e "${CYAN}使用方法:${NC}"
    echo ""
    echo "  sudo ./gpanel.sh <command> [options]"
    echo ""
    echo -e "${CYAN}命令:${NC}"
    echo ""
    echo -e "  ${GREEN}install [version]${NC}    安装 GPanel 正式版本"
    echo "                      不指定版本则交互式选择版本"
    echo "                      示例: sudo ./gpanel.sh install v1.0.0"
    echo ""
    echo -e "  ${GREEN}install-pre${NC}          安装 GPanel 预发布版本"
    echo "                      ${YELLOW}警告: 预发布版本可能不稳定${NC}"
    echo "                      包含最新功能和 Bug 修复"
    echo "                      不建议在生产环境使用"
    echo ""
    echo -e "  ${GREEN}update [version]${NC}     更新 GPanel"
    echo "                      不指定版本则交互式选择版本"
    echo "                      示例: sudo ./gpanel.sh update v1.0.0"
    echo ""
    echo -e "  ${GREEN}uninstall${NC}            卸载 GPanel"
    echo "                      交互式确认，可选择保留数据"
    echo ""
    echo -e "  ${GREEN}status${NC}               查看状态"
    echo "                      显示安装状态、服务状态、端口状态"
    echo ""
    echo -e "  ${GREEN}help${NC}                 显示帮助信息"
    echo ""
    echo -e "${CYAN}版本说明:${NC}"
    echo ""
    echo -e "  ${GREEN}正式版本${NC} (推荐)"
    echo "      - 经过测试的稳定版本"
    echo "      - 适合生产环境使用"
    echo "      - 格式: v1.0.0, v1.1.0 等"
    echo ""
    echo -e "  ${YELLOW}预发布版本${NC} (开发版)"
    echo "      - 包含最新的功能和修复"
    echo "      - 可能存在 Bug 和不稳定因素"
    echo "      - 仅用于测试和体验新功能"
    echo "      - 格式: pre-release-YYYYMMDD-HHMMSS"
    echo ""
    echo -e "${CYAN}交互式操作:${NC}"
    echo ""
    echo "  脚本支持交互式操作，包括:"
    echo "  - 版本选择: 从可用版本列表中选择"
    echo "  - 安装确认: 安装前确认安装信息"
    echo "  - 更新确认: 更新前确认版本信息"
    echo "  - 卸载确认: 卸载前确认并选择是否保留数据"
    echo ""
    echo -e "${CYAN}示例:${NC}"
    echo ""
    echo "  # 交互式安装正式版 (推荐)"
    echo "  sudo ./gpanel.sh install"
    echo ""
    echo "  # 安装指定正式版本"
    echo "  sudo ./gpanel.sh install v1.0.0"
    echo ""
    echo "  # 安装预发布版本 (开发测试用)"
    echo "  sudo ./gpanel.sh install-pre"
    echo ""
    echo "  # 交互式更新"
    echo "  sudo ./gpanel.sh update"
    echo ""
    echo "  # 查看状态"
    echo "  sudo ./gpanel.sh status"
    echo ""
    echo -e "${CYAN}目录:${NC}"
    echo ""
    echo "  程序目录: $INSTALL_DIR"
    echo "  数据目录: $DATA_DIR"
    echo "  日志目录: $LOG_DIR"
    echo ""
}

# ============================================================
# 主函数
# ============================================================

main() {
    # 解析参数
    if [ $# -eq 0 ]; then
        # 无参数时默认显示帮助
        do_help
        exit 0
    fi
    
    ACTION="$1"
    shift
    
    case "$ACTION" in
        install)
            check_root
            do_install "$@"
            ;;
        install-pre)
            check_root
            do_install_pre "$@"
            ;;
        update)
            check_root
            do_update "$@"
            ;;
        uninstall)
            check_root
            do_uninstall
            ;;
        status)
            check_root
            do_status
            ;;
        help|--help|-h)
            do_help
            ;;
        *)
            log_error "未知命令: $ACTION"
            echo ""
            do_help
            exit 1
            ;;
    esac
}

# 运行主函数
main "$@"