package main

import (
	"bufio"
	"crypto/rand"
	"database/sql"
	"fmt"
	"io"
	"math/big"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

const (
	installDir = "/opt/gpanel"
	dataDir    = "/var/lib/gpanel"
	logDir     = "/var/log/gpanel"
	dbPath     = "/var/lib/gpanel/gpanel.db"
)

// 服务定义
var services = []string{"gpanel-agent", "gpanel"} // 注意顺序：agent 先启动

// 配置缓存
type ConfigCache struct {
	settings map[string]string
}

var configCache *ConfigCache

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "status":
		handleStatus()
	case "start":
		handleStart()
	case "stop":
		handleStop()
	case "restart":
		handleRestart()
	case "uninstall":
		handleUninstall()
	case "user-info":
		handleUserInfo()
	case "update":
		handleUpdate()
	case "init-security":
		handleInitSecurity()
	case "get-port":
		handleGetPort()
	case "get-entrance":
		handleGetEntrance()
	case "reset":
		handleReset()
	case "restore":
		handleRestore()
	case "listen-ip":
		handleListenIP()
	case "logs":
		handleLogs()
	case "version", "-v", "--version":
		printVersion()
	case "--help", "-h":
		printUsage()
	default:
		fmt.Printf("未知命令: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("GPanel 命令行管理工具")
	fmt.Println()
	fmt.Println("使用方法:")
	fmt.Println("  gpctl [COMMAND] [ARGS...]")
	fmt.Println("  gpctl --help")
	fmt.Println()
	fmt.Println("命令:")
	fmt.Println("  status              检查 GPanel 服务状态")
	fmt.Println("  start               启动 GPanel 服务 (包括 agent)")
	fmt.Println("  stop                停止 GPanel 服务 (包括 agent)")
	fmt.Println("  restart             重启 GPanel 服务 (包括 agent)")
	fmt.Println("  uninstall           卸载 GPanel 服务")
	fmt.Println("  user-info           获取 GPanel 用户信息")
	fmt.Println("  update              更新 GPanel 配置")
	fmt.Println("    update password   更新面板密码")
	fmt.Println("    update port       更新面板端口")
	fmt.Println("    update username   更新面板用户名")
	fmt.Println("  init-security       初始化安全配置（随机端口和安全入口）")
	fmt.Println("    init-security           随机生成端口和安全入口")
	fmt.Println("    init-security --port N  指定端口，随机安全入口")
	fmt.Println("    init-security --show    仅显示当前连接信息")
	fmt.Println("  get-port            获取服务端口")
	fmt.Println("    get-port                显示 Core 和 Agent 端口")
	fmt.Println("    get-port core           仅显示 Core 端口")
	fmt.Println("    get-port agent          仅显示 Agent 端口")
	fmt.Println("    get-port --quiet        静默输出（格式: CORE_PORT AGENT_PORT）")
	fmt.Println("  get-entrance        获取安全入口")
	fmt.Println("    get-entrance             显示安全入口")
	fmt.Println("    get-entrance --quiet     静默输出（仅输出安全入口路径）")
	fmt.Println("  reset               重置 GPanel 配置")
	fmt.Println("    reset domain      取消域名绑定")
	fmt.Println("    reset entrance    取消安全入口 (设置为 /)")
	fmt.Println("    reset password    重置指定用户密码")
	fmt.Println("        reset password              交互式重置 admin 密码")
	fmt.Println("        reset password <username>   重置指定用户密码")
	fmt.Println("        reset password --quiet      静默重置 admin 密码")
	fmt.Println("        reset password <username> --quiet  静默重置指定用户密码")
	fmt.Println("  restore             恢复 GPanel 到上一个稳定版本")
	fmt.Println("  listen-ip           修改监听 IP")
	fmt.Println("    listen-ip ipv4    监听 IPv4 (0.0.0.0)")
	fmt.Println("    listen-ip ipv6    监听 IPv6 ([::])")
	fmt.Println("  logs                查看日志")
	fmt.Println("    logs core         查看 Core 服务日志")
	fmt.Println("    logs agent        查看 Agent 服务日志")
	fmt.Println("  version             显示版本信息")
	fmt.Println("  --help, -h          显示帮助信息")
	fmt.Println()
	fmt.Println("安装目录:")
	fmt.Printf("  程序目录: %s\n", installDir)
	fmt.Printf("  数据目录: %s\n", dataDir)
	fmt.Printf("  日志目录: %s\n", logDir)
	fmt.Printf("  数据库文件: %s\n", dbPath)
}

func printVersion() {
	fmt.Println("GPanel 管理工具 v1.1.0")
}

// loadConfig 从数据库加载配置
func loadConfig() error {
	if configCache != nil {
		return nil
	}

	// 检查数据库文件是否存在
	if !fileExists(dbPath) {
		return fmt.Errorf("数据库文件不存在: %s", dbPath)
	}

	// 连接数据库
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return fmt.Errorf("连接数据库失败: %v", err)
	}
	defer db.Close()

	// 查询所有设置
	rows, err := db.Query("SELECT key, value FROM settings")
	if err != nil {
		return fmt.Errorf("查询设置失败: %v", err)
	}
	defer rows.Close()

	configCache = &ConfigCache{
		settings: make(map[string]string),
	}

	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return fmt.Errorf("读取设置失败: %v", err)
		}
		configCache.settings[key] = value
	}

	return nil
}

// updateConfig 更新数据库配置
func updateConfig(key, value string) error {
	// 连接数据库
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return fmt.Errorf("连接数据库失败: %v", err)
	}
	defer db.Close()

	// 更新设置
	_, err = db.Exec("UPDATE settings SET value = ? WHERE key = ?", value, key)
	if err != nil {
		return fmt.Errorf("更新设置失败: %v", err)
	}

	// 更新缓存
	if configCache != nil {
		configCache.settings[key] = value
	}

	return nil
}

// getConfig 获取配置值
func getConfig(key string, defaultValue string) string {
	if configCache == nil {
		return defaultValue
	}
	if value, exists := configCache.settings[key]; exists {
		return value
	}
	return defaultValue
}

func handleStatus() {
	if !isRoot() {
		fmt.Println("错误: 需要 root 权限")
		os.Exit(1)
	}

	fmt.Println("=== GPanel 服务状态 ===")
	fmt.Println()

	allRunning := true
	for _, svc := range services {
		if isSystemdService(svc) {
			if isServiceRunning(svc) {
				fmt.Printf("✓ %s: 运行中\n", svc)
			} else {
				fmt.Printf("✗ %s: 未运行\n", svc)
				allRunning = false
			}
		} else {
			fmt.Printf("✗ %s: 未安装\n", svc)
			allRunning = false
		}
	}

	fmt.Println()

	// 加载配置
	if err := loadConfig(); err == nil {
		fmt.Println("=== 运行信息 ===")

		// 获取端口
		port := getConfig("ServerPort", "8080")
		securityEntrance := getConfig("SecurityEntrance", "/")
		listenAddress := getConfig("ListenAddress", "0.0.0.0")

		// Core 地址
		fmt.Printf("  Core 监听地址: %s:%s\n", listenAddress, port)

		// 安全入口
		if securityEntrance != "/" {
			fmt.Printf("  安全入口: %s\n", securityEntrance)
			fmt.Printf("  访问地址: http(s)://你的IP:%s%s\n", port, securityEntrance)
		} else {
			fmt.Println("  安全入口: 未设置 (/)")
			fmt.Printf("  访问地址: http(s)://你的IP:%s\n", port)
		}

		// Agent 地址
		fmt.Println()
		fmt.Println("  Agent 监听地址: 0.0.0.0:9998 (默认)")
	} else {
		fmt.Printf("无法加载配置: %v\n", err)
	}

	fmt.Println()

	if allRunning {
		fmt.Println("详细信息:")
		for _, svc := range services {
			fmt.Printf("\n--- %s ---\n", svc)
			cmd := exec.Command("systemctl", "status", svc, "--no-pager")
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Run()
		}
	}
}

func handleStart() {
	if !isRoot() {
		fmt.Println("错误: 需要 root 权限")
		os.Exit(1)
	}

	fmt.Println("正在启动 GPanel 服务...")

	for _, svc := range services {
		if !isSystemdService(svc) {
			fmt.Printf("错误: %s 服务未安装\n", svc)
			os.Exit(1)
		}
	}

	for _, svc := range services {
		fmt.Printf("启动 %s...\n", svc)
		cmd := exec.Command("systemctl", "start", svc)
		if err := cmd.Run(); err != nil {
			fmt.Printf("启动 %s 失败: %v\n", svc, err)
			os.Exit(1)
		}
		time.Sleep(1 * time.Second)
	}

	// 等待服务启动
	time.Sleep(2 * time.Second)

	allRunning := true
	for _, svc := range services {
		if !isServiceRunning(svc) {
			fmt.Printf("✗ %s 启动失败\n", svc)
			allRunning = false
		} else {
			fmt.Printf("✓ %s 启动成功\n", svc)
		}
	}

	if allRunning {
		fmt.Println("\n✓ GPanel 服务启动成功")

		// 显示访问信息
		if err := loadConfig(); err == nil {
			port := getConfig("ServerPort", "8080")
			securityEntrance := getConfig("SecurityEntrance", "/")
			if securityEntrance != "/" {
				fmt.Printf("\n访问地址: http(s)://你的IP:%s%s\n", port, securityEntrance)
			} else {
				fmt.Printf("\n访问地址: http(s)://你的IP:%s\n", port)
			}
		}
	} else {
		fmt.Println("\n✗ GPanel 服务启动失败")
		os.Exit(1)
	}
}

func handleStop() {
	if !isRoot() {
		fmt.Println("错误: 需要 root 权限")
		os.Exit(1)
	}

	fmt.Println("正在停止 GPanel 服务...")

	// 反向停止：先停 gpanel，再停 agent
	for i := len(services) - 1; i >= 0; i-- {
		svc := services[i]
		if isSystemdService(svc) {
			fmt.Printf("停止 %s...\n", svc)
			cmd := exec.Command("systemctl", "stop", svc)
			if err := cmd.Run(); err != nil {
				fmt.Printf("停止 %s 失败: %v\n", svc, err)
			}
		}
	}

	// 等待服务停止
	time.Sleep(2 * time.Second)

	allStopped := true
	for _, svc := range services {
		if isSystemdService(svc) && isServiceRunning(svc) {
			fmt.Printf("✗ %s 停止失败\n", svc)
			allStopped = false
		} else {
			fmt.Printf("✓ %s 已停止\n", svc)
		}
	}

	if allStopped {
		fmt.Println("\n✓ GPanel 服务已停止")
	} else {
		fmt.Println("\n✗ GPanel 服务停止异常")
		os.Exit(1)
	}
}

func handleRestart() {
	if !isRoot() {
		fmt.Println("错误: 需要 root 权限")
		os.Exit(1)
	}

	fmt.Println("正在重启 GPanel 服务...")

	for _, svc := range services {
		if !isSystemdService(svc) {
			fmt.Printf("错误: %s 服务未安装\n", svc)
			os.Exit(1)
		}
	}

	// 反向停止
	for i := len(services) - 1; i >= 0; i-- {
		svc := services[i]
		fmt.Printf("停止 %s...\n", svc)
		exec.Command("systemctl", "stop", svc).Run()
	}

	time.Sleep(1 * time.Second)

	// 正向启动
	for _, svc := range services {
		fmt.Printf("启动 %s...\n", svc)
		cmd := exec.Command("systemctl", "start", svc)
		if err := cmd.Run(); err != nil {
			fmt.Printf("启动 %s 失败: %v\n", svc, err)
			os.Exit(1)
		}
		time.Sleep(1 * time.Second)
	}

	// 等待服务启动
	time.Sleep(2 * time.Second)

	allRunning := true
	for _, svc := range services {
		if !isServiceRunning(svc) {
			fmt.Printf("✗ %s 重启失败\n", svc)
			allRunning = false
		} else {
			fmt.Printf("✓ %s 重启成功\n", svc)
		}
	}

	if allRunning {
		fmt.Println("\n✓ GPanel 服务重启成功")

		// 显示访问信息
		if err := loadConfig(); err == nil {
			port := getConfig("ServerPort", "8080")
			securityEntrance := getConfig("SecurityEntrance", "/")
			if securityEntrance != "/" {
				fmt.Printf("\n访问地址: http(s)://你的IP:%s%s\n", port, securityEntrance)
			} else {
				fmt.Printf("\n访问地址: http(s)://你的IP:%s\n", port)
			}
		}
	} else {
		fmt.Println("\n✗ GPanel 服务重启失败")
		os.Exit(1)
	}
}

func handleUninstall() {
	if !isRoot() {
		fmt.Println("错误: 需要 root 权限")
		os.Exit(1)
	}

	fmt.Println("警告: 此操作将卸载 GPanel 服务")
	fmt.Print("确认卸载? (yes/no): ")
	var confirm string
	fmt.Scanln(&confirm)

	if strings.ToLower(confirm) != "yes" {
		fmt.Println("取消卸载")
		return
	}

	fmt.Println("正在卸载 GPanel 服务...")

	// 停止并禁用服务
	for _, svc := range services {
		if isSystemdService(svc) {
			fmt.Printf("停止 %s...\n", svc)
			exec.Command("systemctl", "stop", svc).Run()
			exec.Command("systemctl", "disable", svc).Run()
		}
	}

	// 删除 systemd 服务文件
	for _, svc := range services {
		serviceFile := "/etc/systemd/system/" + svc + ".service"
		if fileExists(serviceFile) {
			os.Remove(serviceFile)
		}
	}
	exec.Command("systemctl", "daemon-reload").Run()

	// 删除二进制文件
	binaries := []string{
		installDir + "/gpanel",
		installDir + "/gpanel-agent",
		"/usr/local/bin/gpctl",
	}
	for _, bin := range binaries {
		if fileExists(bin) {
			os.Remove(bin)
			fmt.Printf("删除: %s\n", bin)
		}
	}

	// 询问是否删除数据和配置
	fmt.Print("是否删除数据? (yes/no): ")
	var deleteData string
	fmt.Scanln(&deleteData)

	if strings.ToLower(deleteData) == "yes" {
		// 删除安装目录
		if dirExists(installDir) {
			os.RemoveAll(installDir)
			fmt.Printf("删除目录: %s\n", installDir)
		}

		// 删除数据目录（包含数据库）
		if dirExists(dataDir) {
			os.RemoveAll(dataDir)
			fmt.Printf("删除目录: %s\n", dataDir)
		}

		// 删除日志目录
		if dirExists(logDir) {
			os.RemoveAll(logDir)
			fmt.Printf("删除目录: %s\n", logDir)
		}
	}

	fmt.Println("\n✓ GPanel 服务卸载完成")
}

func handleUserInfo() {
	// 加载配置
	if err := loadConfig(); err != nil {
		fmt.Printf("无法加载配置: %v\n", err)
		os.Exit(1)
	}

	port := getConfig("ServerPort", "8080")
	securityEntrance := getConfig("SecurityEntrance", "/")
	username := getConfig("PanelUser", "admin")
	password := getConfig("PanelPassword", "admin123")

	// 获取服务器 IP
	serverIP := getServerIP()

	// 构建访问地址
	var address string
	if securityEntrance != "/" {
		address = fmt.Sprintf("https://%s:%s%s", serverIP, port, securityEntrance)
	} else {
		address = fmt.Sprintf("https://%s:%s", serverIP, port)
	}

	// 隐藏密码（显示星号）
	maskedPassword := strings.Repeat("*", len(password))

	// 输出用户信息
	fmt.Printf("Panel address: %s\n", address)
	fmt.Printf("Panel user: %s\n", username)
	fmt.Printf("Panel password: %s\n", maskedPassword)
	fmt.Println("Tip: To change the password, you can execute the command: gpctl update password")
}

// handleUpdate 处理 update 命令
func handleUpdate() {
	if !isRoot() {
		fmt.Println("错误: 需要 root 权限")
		os.Exit(1)
	}

	if len(os.Args) < 3 {
		fmt.Println("请指定更新选项:")
		fmt.Println("  gpctl update password  更新面板密码")
		fmt.Println("  gpctl update port      更新面板端口")
		fmt.Println("  gpctl update username  更新面板用户名")
		os.Exit(1)
	}

	subCommand := os.Args[2]

	switch subCommand {
	case "password":
		updatePassword()
	case "port":
		updatePort()
	case "username":
		updateUsername()
	default:
		fmt.Printf("未知的更新选项: %s\n", subCommand)
		fmt.Println("可用选项: password, port, username")
		os.Exit(1)
	}
}

// updatePassword 更新面板密码
func updatePassword() {
	if err := loadConfig(); err != nil {
		fmt.Printf("加载配置失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("=== 更新面板密码 ===")
	fmt.Println()

	// 读取新密码
	fmt.Print("请输入新密码: ")
	reader := bufio.NewReader(os.Stdin)
	newPassword, err := reader.ReadString('\n')
	if err != nil {
		fmt.Printf("读取输入失败: %v\n", err)
		os.Exit(1)
	}
	newPassword = strings.TrimSpace(newPassword)

	if newPassword == "" {
		fmt.Println("错误: 密码不能为空")
		os.Exit(1)
	}

	if len(newPassword) < 6 {
		fmt.Println("错误: 密码长度至少为 6 个字符")
		os.Exit(1)
	}

	// 确认新密码
	fmt.Print("请再次输入新密码: ")
	confirmPassword, err := reader.ReadString('\n')
	if err != nil {
		fmt.Printf("读取输入失败: %v\n", err)
		os.Exit(1)
	}
	confirmPassword = strings.TrimSpace(confirmPassword)

	if newPassword != confirmPassword {
		fmt.Println("错误: 两次输入的密码不一致")
		os.Exit(1)
	}

	// 更新密码
	if err := updateConfig("PanelPassword", newPassword); err != nil {
		fmt.Printf("更新密码失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✓ 密码已更新")
	fmt.Println("提示: 重新登录后生效")
}

// updatePort 更新面板端口
func updatePort() {
	if err := loadConfig(); err != nil {
		fmt.Printf("加载配置失败: %v\n", err)
		os.Exit(1)
	}

	currentPort := getConfig("ServerPort", "8080")

	fmt.Println("=== 更新面板端口 ===")
	fmt.Println()
	fmt.Printf("当前端口: %s\n", currentPort)
	fmt.Println()

	// 读取新端口
	fmt.Print("请输入新端口 (1-65535): ")
	reader := bufio.NewReader(os.Stdin)
	portInput, err := reader.ReadString('\n')
	if err != nil {
		fmt.Printf("读取输入失败: %v\n", err)
		os.Exit(1)
	}
	portInput = strings.TrimSpace(portInput)

	// 验证端口
	var newPort int
	if _, err := fmt.Sscanf(portInput, "%d", &newPort); err != nil {
		fmt.Println("错误: 请输入有效的端口号")
		os.Exit(1)
	}

	if newPort < 1 || newPort > 65535 {
		fmt.Println("错误: 端口号必须在 1-65535 范围内")
		os.Exit(1)
	}

	if portInput == currentPort {
		fmt.Println("新端口与当前端口相同，无需更新")
		return
	}

	// 更新端口
	if err := updateConfig("ServerPort", portInput); err != nil {
		fmt.Printf("更新端口失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ 端口已更新: %s\n", portInput)

	// 重启服务使配置生效
	fmt.Println("正在重启服务...")
	restartServices()

	// 显示新的访问地址
	securityEntrance := getConfig("SecurityEntrance", "/")
	serverIP := getServerIP()
	if securityEntrance != "/" {
		fmt.Printf("\n访问地址: https://%s:%s%s\n", serverIP, portInput, securityEntrance)
	} else {
		fmt.Printf("\n访问地址: https://%s:%s\n", serverIP, portInput)
	}
}

// updateUsername 更新面板用户名
func updateUsername() {
	if err := loadConfig(); err != nil {
		fmt.Printf("加载配置失败: %v\n", err)
		os.Exit(1)
	}

	currentUsername := getConfig("PanelUser", "admin")

	fmt.Println("=== 更新面板用户名 ===")
	fmt.Println()
	fmt.Printf("当前用户名: %s\n", currentUsername)
	fmt.Println()

	// 读取新用户名
	fmt.Print("请输入新用户名: ")
	reader := bufio.NewReader(os.Stdin)
	newUsername, err := reader.ReadString('\n')
	if err != nil {
		fmt.Printf("读取输入失败: %v\n", err)
		os.Exit(1)
	}
	newUsername = strings.TrimSpace(newUsername)

	if newUsername == "" {
		fmt.Println("错误: 用户名不能为空")
		os.Exit(1)
	}

	if len(newUsername) < 3 {
		fmt.Println("错误: 用户名长度至少为 3 个字符")
		os.Exit(1)
	}

	if newUsername == currentUsername {
		fmt.Println("新用户名与当前用户名相同，无需更新")
		return
	}

	// 更新用户名
	if err := updateConfig("PanelUser", newUsername); err != nil {
		fmt.Printf("更新用户名失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ 用户名已更新: %s\n", newUsername)
	fmt.Println("提示: 重新登录后生效")
}

// handleReset 处理 reset 命令
func handleReset() {
	if !isRoot() {
		fmt.Println("错误: 需要 root 权限")
		os.Exit(1)
	}

	if len(os.Args) < 3 {
		fmt.Println("请指定重置选项:")
		fmt.Println("  gpctl reset domain    取消域名绑定")
		fmt.Println("  gpctl reset entrance  取消安全入口 (设置为 /)")
		fmt.Println("  gpctl reset password  重置指定用户密码")
		os.Exit(1)
	}

	subCommand := os.Args[2]

	switch subCommand {
	case "domain":
		resetDomain()
	case "entrance":
		resetEntrance()
	case "password":
		resetPassword()
	default:
		fmt.Printf("未知的重置选项: %s\n", subCommand)
		fmt.Println("可用选项: domain, entrance, password")
		os.Exit(1)
	}
}

// resetDomain 取消域名绑定
func resetDomain() {
	if err := loadConfig(); err != nil {
		fmt.Printf("加载配置失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("正在取消 GPanel 域名绑定...")

	// 清空 ServerAddress
	if err := updateConfig("ServerAddress", ""); err != nil {
		fmt.Printf("取消域名绑定失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✓ 域名绑定已取消")

	// 重启服务使配置生效
	fmt.Println("正在重启服务...")
	restartServices()
}

// resetEntrance 取消安全入口
func resetEntrance() {
	if err := loadConfig(); err != nil {
		fmt.Printf("加载配置失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("正在取消 GPanel 安全入口...")

	// 设置安全入口为 /
	if err := updateConfig("SecurityEntrance", "/"); err != nil {
		fmt.Printf("取消安全入口失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✓ 安全入口已取消")

	// 重启服务使配置生效
	fmt.Println("正在重启服务...")
	restartServices()
}

// handleRestore 处理 restore 命令
func handleRestore() {
	if !isRoot() {
		fmt.Println("错误: 需要 root 权限")
		os.Exit(1)
	}

	fmt.Println("GPanel 将恢复到上一个稳定版本。是否继续? [y/n]: ")

	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))

	if input != "y" && input != "yes" {
		fmt.Println("已取消恢复操作")
		return
	}

	fmt.Println("正在恢复 GPanel...")

	// 查找备份目录
	backupDir := filepath.Join(dataDir, "backups")
	if !dirExists(backupDir) {
		fmt.Println("错误: 未找到备份目录")
		os.Exit(1)
	}

	// 查找最新的备份文件
	files, err := os.ReadDir(backupDir)
	if err != nil {
		fmt.Printf("读取备份目录失败: %v\n", err)
		os.Exit(1)
	}

	var latestBackup string
	var latestTime time.Time
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".db") || strings.HasSuffix(file.Name(), ".sql") {
			info, err := file.Info()
			if err != nil {
				continue
			}
			if info.ModTime().After(latestTime) {
				latestTime = info.ModTime()
				latestBackup = filepath.Join(backupDir, file.Name())
			}
		}
	}

	if latestBackup == "" {
		fmt.Println("错误: 未找到备份文件")
		os.Exit(1)
	}

	fmt.Printf("找到备份文件: %s\n", latestBackup)
	fmt.Printf("备份时间: %s\n", latestTime.Format("2006-01-02 15:04:05"))

	// 停止服务
	fmt.Println("正在停止服务...")
	exec.Command("systemctl", "stop", "gpanel").Run()

	// 备份当前数据库
	currentBackup := dbPath + ".backup." + time.Now().Format("20060102150405")
	if fileExists(dbPath) {
		if err := copyFile(dbPath, currentBackup); err != nil {
			fmt.Printf("备份当前数据库失败: %v\n", err)
		} else {
			fmt.Printf("当前数据库已备份到: %s\n", currentBackup)
		}
	}

	// 恢复数据库
	if strings.HasSuffix(latestBackup, ".db") {
		if err := copyFile(latestBackup, dbPath); err != nil {
			fmt.Printf("恢复数据库失败: %v\n", err)
			os.Exit(1)
		}
	} else if strings.HasSuffix(latestBackup, ".sql") {
		// SQL 文件需要导入
		fmt.Println("正在导入 SQL 备份...")
		cmd := exec.Command("sqlite3", dbPath, ".read "+latestBackup)
		if err := cmd.Run(); err != nil {
			fmt.Printf("导入 SQL 失败: %v\n", err)
			os.Exit(1)
		}
	}

	fmt.Println("✓ 数据库已恢复")

	// 启动服务
	fmt.Println("正在启动服务...")
	exec.Command("systemctl", "start", "gpanel").Run()

	time.Sleep(2 * time.Second)

	if isServiceRunning("gpanel") {
		fmt.Println("✓ GPanel 恢复成功")
	} else {
		fmt.Println("✗ GPanel 启动失败，请检查日志")
		os.Exit(1)
	}
}

// handleListenIP 处理 listen-ip 命令
func handleListenIP() {
	if !isRoot() {
		fmt.Println("错误: 需要 root 权限")
		os.Exit(1)
	}

	if len(os.Args) < 3 {
		fmt.Println("请指定监听 IP 选项:")
		fmt.Println("  gpctl listen-ip ipv4  监听 IPv4 (0.0.0.0)")
		fmt.Println("  gpctl listen-ip ipv6  监听 IPv6 ([::])")
		os.Exit(1)
	}

	subCommand := os.Args[2]

	var listenAddr string
	switch subCommand {
	case "ipv4":
		listenAddr = "0.0.0.0"
		fmt.Println("正在设置监听 IPv4...")
	case "ipv6":
		listenAddr = "::"
		fmt.Println("正在设置监听 IPv6...")
	default:
		fmt.Printf("未知的选项: %s\n", subCommand)
		fmt.Println("可用选项: ipv4, ipv6")
		os.Exit(1)
	}

	if err := loadConfig(); err != nil {
		fmt.Printf("加载配置失败: %v\n", err)
		os.Exit(1)
	}

	// 更新 ListenAddress
	if err := updateConfig("ListenAddress", listenAddr); err != nil {
		fmt.Printf("设置监听地址失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ 监听地址已设置为: %s\n", listenAddr)

	// 重启服务使配置生效
	fmt.Println("正在重启服务...")
	restartServices()
}

// handleLogs 处理 logs 命令
func handleLogs() {
	if !isRoot() {
		fmt.Println("错误: 需要 root 权限")
		os.Exit(1)
	}

	if len(os.Args) < 3 {
		fmt.Println("请指定日志类型:")
		fmt.Println("  gpctl logs core   查看 Core 服务日志")
		fmt.Println("  gpctl logs agent  查看 Agent 服务日志")
		os.Exit(1)
	}

	subCommand := os.Args[2]
	var logFile string
	var serviceName string

	switch subCommand {
	case "core":
		logFile = filepath.Join(logDir, "gpanel.log")
		serviceName = "gpanel"
		fmt.Println("=== GPanel Core 日志 ===")
	case "agent":
		logFile = filepath.Join(logDir, "agent.log")
		serviceName = "gpanel-agent"
		fmt.Println("=== GPanel Agent 日志 ===")
	default:
		fmt.Printf("未知的日志类型: %s\n", subCommand)
		fmt.Println("可用选项: core, agent")
		os.Exit(1)
	}

	// 检查日志文件是否存在
	if !fileExists(logFile) {
		fmt.Printf("日志文件不存在: %s\n", logFile)

		// 尝试使用 journalctl
		fmt.Println("尝试从 systemd 日志读取...")
		cmd := exec.Command("journalctl", "-u", serviceName, "-n", "100", "--no-pager")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Run()
		return
	}

	// 检查是否有 -f 参数（实时跟踪）
	if len(os.Args) > 3 && os.Args[3] == "-f" {
		// 使用 tail -f 实时跟踪
		cmd := exec.Command("tail", "-f", logFile)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		cmd.Run()
		return
	}

	// 读取最后 100 行
	cmd := exec.Command("tail", "-n", "100", logFile)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()

	fmt.Println()
	fmt.Printf("提示: 使用 'gpctl logs %s -f' 实时查看日志\n", subCommand)
}

// restartServices 重启服务
func restartServices() {
	// 反向停止
	for i := len(services) - 1; i >= 0; i-- {
		svc := services[i]
		exec.Command("systemctl", "stop", svc).Run()
	}

	time.Sleep(1 * time.Second)

	// 正向启动
	for _, svc := range services {
		exec.Command("systemctl", "start", svc).Run()
		time.Sleep(1 * time.Second)
	}

	time.Sleep(2 * time.Second)

	allRunning := true
	for _, svc := range services {
		if !isServiceRunning(svc) {
			fmt.Printf("✗ %s 重启失败\n", svc)
			allRunning = false
		} else {
			fmt.Printf("✓ %s 重启成功\n", svc)
		}
	}

	if allRunning {
		// 显示访问信息
		if err := loadConfig(); err == nil {
			port := getConfig("ServerPort", "8080")
			securityEntrance := getConfig("SecurityEntrance", "/")
			if securityEntrance != "/" {
				fmt.Printf("\n访问地址: http(s)://你的IP:%s%s\n", port, securityEntrance)
			} else {
				fmt.Printf("\n访问地址: http(s)://你的IP:%s\n", port)
			}
		}
	}
}

// copyFile 复制文件
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}

	// 复制权限
	sourceInfo, err := os.Stat(src)
	if err != nil {
		return err
	}
	return os.Chmod(dst, sourceInfo.Mode())
}

func isRoot() bool {
	if runtime.GOOS == "windows" {
		return true // Windows 下跳过 root 检查
	}
	return os.Geteuid() == 0
}

func isSystemdService(svc string) bool {
	if runtime.GOOS == "windows" {
		return false
	}
	serviceFile := "/etc/systemd/system/" + svc + ".service"
	return fileExists(serviceFile)
}

func isServiceRunning(svc string) bool {
	if runtime.GOOS == "windows" {
		return false
	}
	cmd := exec.Command("systemctl", "is-active", svc)
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(output)) == "active"
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// handleInitSecurity 处理 init-security 命令
// 功能：初始化安全配置（随机端口和安全入口）
// 参数：
//
//	--port N   指定端口
//	--show     仅显示当前连接信息
func handleInitSecurity() {
	if !isRoot() {
		fmt.Println("错误: 需要 root 权限")
		os.Exit(1)
	}

	// 解析参数
	showOnly := false
	specifiedPort := ""

	for i := 2; i < len(os.Args); i++ {
		arg := os.Args[i]
		if arg == "--show" {
			showOnly = true
		} else if arg == "--port" && i+1 < len(os.Args) {
			specifiedPort = os.Args[i+1]
			i++
		}
	}

	// 加载当前配置
	if err := loadConfig(); err != nil {
		fmt.Printf("加载配置失败: %v\n", err)
		os.Exit(1)
	}

	// 仅显示模式
	if showOnly {
		printConnectionInfo()
		return
	}

	fmt.Println("=== GPanel 安全配置初始化 ===")
	fmt.Println()

	// 生成随机端口（如果未指定）
	var newPort string
	if specifiedPort != "" {
		newPort = specifiedPort
		fmt.Printf("使用指定端口: %s\n", newPort)
	} else {
		newPort = generateRandomPort()
		fmt.Printf("随机生成端口: %s\n", newPort)
	}

	// 生成随机安全入口
	newEntrance := generateRandomEntrance()
	fmt.Printf("随机生成安全入口: %s\n", newEntrance)

	// 更新配置
	fmt.Println()
	fmt.Println("正在更新配置...")

	if err := updateConfig("ServerPort", newPort); err != nil {
		fmt.Printf("更新端口失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✓ 端口已更新: %s\n", newPort)

	if err := updateConfig("SecurityEntrance", newEntrance); err != nil {
		fmt.Printf("更新安全入口失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✓ 安全入口已更新: %s\n", newEntrance)

	// 重启服务
	fmt.Println()
	fmt.Println("正在重启服务...")
	restartServices()

	// 显示连接信息
	fmt.Println()
	printConnectionInfo()
}

// handleGetPort 处理 get-port 命令
// 功能：获取服务端口
// 参数：
//
//	core/agent  指定服务
//	--quiet     静默输出（仅输出端口号）
func handleGetPort() {
	// 解析参数
	quiet := false
	service := ""

	for i := 2; i < len(os.Args); i++ {
		arg := os.Args[i]
		if arg == "--quiet" || arg == "-q" {
			quiet = true
		} else if arg != "" && !strings.HasPrefix(arg, "-") {
			service = arg
		}
	}

	// 获取端口
	corePort := getCorePort()
	agentPort := getAgentPort()

	// 静默模式：只输出端口号
	if quiet {
		switch service {
		case "core":
			fmt.Println(corePort)
		case "agent":
			fmt.Println(agentPort)
		default:
			fmt.Printf("%s %s\n", corePort, agentPort)
		}
		return
	}

	// 格式化输出
	switch service {
	case "core":
		fmt.Println(corePort)
	case "agent":
		fmt.Println(agentPort)
	default:
		fmt.Printf("Core 端口: %s\n", corePort)
		fmt.Printf("Agent 端口: %s\n", agentPort)
	}
}

// handleGetEntrance 处理 get-entrance 命令
// 功能：获取安全入口
// 参数：
//
//	--quiet     静默输出（仅输出安全入口路径）
func handleGetEntrance() {
	// 解析参数
	quiet := false

	for i := 2; i < len(os.Args); i++ {
		arg := os.Args[i]
		if arg == "--quiet" || arg == "-q" {
			quiet = true
		}
	}

	// 获取安全入口
	entrance := getSecurityEntrance()

	// 静默模式
	if quiet {
		fmt.Println(entrance)
		return
	}

	// 格式化输出
	fmt.Println(entrance)
}

// getSecurityEntrance 获取安全入口
func getSecurityEntrance() string {
	// 尝试从数据库加载配置
	if err := loadConfig(); err == nil {
		return getConfig("SecurityEntrance", "/")
	}
	return "/"
}

// getCorePort 获取 Core 服务端口
func getCorePort() string {
	// 尝试从数据库加载配置
	if err := loadConfig(); err == nil {
		return getConfig("ServerPort", "8080")
	}
	return "8080"
}

// getAgentPort 获取 Agent 服务端口
func getAgentPort() string {
	// 1. 尝试从 systemd 服务环境变量读取
	cmd := exec.Command("systemctl", "show", "gpanel-agent", "--property=Environment", "--value")
	output, err := cmd.Output()
	if err == nil {
		env := strings.TrimSpace(string(output))
		// 解析环境变量，查找 GAGENT_LISTEN
		for _, line := range strings.Split(env, "\n") {
			if strings.HasPrefix(line, "GAGENT_LISTEN=") {
				listenAddr := strings.TrimPrefix(line, "GAGENT_LISTEN=")
				// 提取端口（格式: 0.0.0.0:9998 或 [::]:9998）
				if idx := strings.LastIndex(listenAddr, ":"); idx != -1 {
					return listenAddr[idx+1:]
				}
			}
		}
	}

	// 2. 尝试从服务文件中读取 EnvironmentFile
	serviceFile := "/etc/systemd/system/gpanel-agent.service"
	if content, err := os.ReadFile(serviceFile); err == nil {
		contentStr := string(content)
		// 查找 EnvironmentFile
		for _, line := range strings.Split(contentStr, "\n") {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "EnvironmentFile=") {
				envFile := strings.TrimPrefix(line, "EnvironmentFile=")
				envFile = strings.Trim(envFile, "\"'")
				if envContent, err := os.ReadFile(envFile); err == nil {
					for _, envLine := range strings.Split(string(envContent), "\n") {
						envLine = strings.TrimSpace(envLine)
						if strings.HasPrefix(envLine, "GAGENT_LISTEN=") {
							listenAddr := strings.TrimPrefix(envLine, "GAGENT_LISTEN=")
							if idx := strings.LastIndex(listenAddr, ":"); idx != -1 {
								return listenAddr[idx+1:]
							}
						}
					}
				}
			}
		}
	}

	// 3. 返回默认值
	return "9998"
}

// generateRandomPort 生成随机端口（10000-65535）
func generateRandomPort() string {
	// 使用加密随机数生成器
	minPort := int64(10000)
	maxPort := int64(65535)
	rangePort := maxPort - minPort + 1

	n, err := rand.Int(rand.Reader, big.NewInt(rangePort))
	if err != nil {
		// 回退到时间戳方式
		return fmt.Sprintf("%d", minPort+time.Now().UnixNano()%rangePort)
	}

	return fmt.Sprintf("%d", minPort+n.Int64())
}

// generateRandomEntrance 生成随机安全入口（8位小写字母和数字）
func generateRandomEntrance() string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	const length = 8

	b := make([]byte, length)
	for i := range b {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			// 回退到时间戳方式
			b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
			continue
		}
		b[i] = charset[n.Int64()]
	}
	return "/" + string(b)
}

// printConnectionInfo 打印连接信息
func printConnectionInfo() {
	fmt.Println("========================================")
	fmt.Println("  GPanel 连接信息")
	fmt.Println("========================================")
	fmt.Println()

	// 重新加载配置
	if err := loadConfig(); err != nil {
		fmt.Printf("加载配置失败: %v\n", err)
		return
	}

	port := getConfig("ServerPort", "8080")
	securityEntrance := getConfig("SecurityEntrance", "/")
	listenAddress := getConfig("ListenAddress", "0.0.0.0")

	// 获取服务器 IP
	serverIP := getServerIP()

	fmt.Printf("  监听地址: %s:%s\n", listenAddress, port)
	fmt.Println()

	if securityEntrance != "/" {
		fmt.Printf("  安全入口: %s\n", securityEntrance)
		fmt.Println()
		fmt.Printf("  本地访问: http://localhost:%s%s\n", port, securityEntrance)
		fmt.Printf("  外部访问: http://%s:%s%s\n", serverIP, port, securityEntrance)
	} else {
		fmt.Println("  安全入口: 未设置 (/)")
		fmt.Println()
		fmt.Printf("  本地访问: http://localhost:%s\n", port)
		fmt.Printf("  外部访问: http://%s:%s\n", serverIP, port)
	}

	fmt.Println()
	fmt.Println("========================================")
}

// getServerIP 获取服务器 IP 地址
func getServerIP() string {
	// 尝试获取外部 IP
	cmd := exec.Command("hostname", "-I")
	output, err := cmd.Output()
	if err == nil {
		ips := strings.TrimSpace(string(output))
		// 取第一个 IP
		if ips != "" {
			return strings.Fields(ips)[0]
		}
	}

	// 备用方案：从路由获取
	cmd = exec.Command("ip", "route", "get", "1")
	output, err = cmd.Output()
	if err == nil {
		fields := strings.Fields(string(output))
		for i, f := range fields {
			if f == "src" && i+1 < len(fields) {
				return fields[i+1]
			}
		}
	}

	return "your-server-ip"
}

// resetPassword 重置指定用户的密码
func resetPassword() {
	// 解析参数
	quiet := false
	username := "admin"

	// 遍历参数
	for i := 3; i < len(os.Args); i++ {
		arg := os.Args[i]
		if arg == "--quiet" || arg == "-q" {
			quiet = true
		} else if !strings.HasPrefix(arg, "-") {
			username = arg
		}
	}

	// 加载配置
	if err := loadConfig(); err != nil {
		fmt.Printf("加载配置失败: %v\n", err)
		os.Exit(1)
	}

	// 验证用户名是否存在
	currentUsername := getConfig("PanelUser", "admin")
	if username != currentUsername {
		fmt.Printf("错误: 用户 '%s' 不存在\n", username)
		fmt.Printf("当前用户: %s\n", currentUsername)
		os.Exit(1)
	}

	// 生成随机密码
	newPassword := generateRandomPassword()

	// 更新密码
	if err := updateConfig("PanelPassword", newPassword); err != nil {
		fmt.Printf("重置密码失败: %v\n", err)
		os.Exit(1)
	}

	// 静默模式：仅输出用户名和密码
	if quiet {
		fmt.Printf("%s %s\n", username, newPassword)
		return
	}

	// 交互模式：显示详细信息
	fmt.Println("=== 密码重置成功 ===")
	fmt.Println()
	fmt.Printf("  用户名: %s\n", username)
	fmt.Printf("  新密码: %s\n", newPassword)
	fmt.Println()
	fmt.Println("提示: 请妥善保管新密码")
}

// generateRandomPassword 生成随机密码
// 8位长度，包含大小写字母、数字和符号
func generateRandomPassword() string {
	const (
		length      = 8
		lowercase   = "abcdefghijklmnopqrstuvwxyz"
		uppercase   = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
		digits      = "0123456789"
		symbols     = "!@#$%^&*"
		allChars    = lowercase + uppercase + digits + symbols
	)

	password := make([]byte, length)

	// 确保每种字符至少有一个
	charSets := []string{lowercase, uppercase, digits, symbols}
	for i := 0; i < 4; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charSets[i]))))
		if err != nil {
			password[i] = charSets[i][0]
			continue
		}
		password[i] = charSets[i][n.Int64()]
	}

	// 填充剩余位置
	for i := 4; i < length; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(allChars))))
		if err != nil {
			password[i] = allChars[0]
			continue
		}
		password[i] = allChars[n.Int64()]
	}

	// 随机打乱顺序
	for i := len(password) - 1; i > 0; i-- {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
		if err != nil {
			continue
		}
		password[i], password[n.Int64()] = password[n.Int64()], password[i]
	}

	return string(password)
}
