package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

const (
	installDir = "/opt/gpanel"
	dataDir    = "/var/lib/gpanel"
	logDir     = "/var/log/gpanel"
	dbPath     = "/var/lib/gpanel/gpanel.db"
)

// 服务定义
var services = []string{"gpanel-agent", "gpanel"} // 注意顺序：agent 先启动

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
	fmt.Println("GPanel 管理工具 v1.0.0")
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
	fmt.Println("=== GPanel 配置信息 ===")
	fmt.Println()

	// 检查数据库文件
	if fileExists(dbPath) {
		fmt.Printf("数据库文件: %s\n", dbPath)
	} else {
		fmt.Printf("数据库文件不存在: %s\n", dbPath)
	}

	// 服务状态
	fmt.Println()
	fmt.Println("=== 服务状态 ===")
	for _, svc := range services {
		if isServiceRunning(svc) {
			fmt.Printf("✓ %s: 运行中\n", svc)
		} else {
			fmt.Printf("✗ %s: 未运行\n", svc)
		}
	}

	// 尝试从 API 获取信息
	port := "8080"
	url := fmt.Sprintf("http://localhost:%s/api/v1/config", port)
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(url)
	if err == nil && resp.StatusCode == 200 {
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		var config map[string]interface{}
		if err := json.Unmarshal(body, &config); err == nil {
			fmt.Println("\n=== API 配置 ===")
			if data, ok := config["data"].(map[string]interface{}); ok {
				if server, ok := data["server"].(map[string]interface{}); ok {
					if p, ok := server["port"].(string); ok {
						fmt.Printf("  端口: %s\n", p)
					}
					if mode, ok := server["mode"].(string); ok {
						fmt.Printf("  模式: %s\n", mode)
					}
				}
			}
		}
	} else {
		fmt.Printf("\n服务端口: %s (Core), 9998 (Agent)\n", port)
	}

	fmt.Println("\n提示: GPanel 配置存储于数据库，可通过 Web 界面修改")
}

func isRoot() bool {
	return os.Geteuid() == 0
}

func isSystemdService(svc string) bool {
	serviceFile := "/etc/systemd/system/" + svc + ".service"
	return fileExists(serviceFile)
}

func isServiceRunning(svc string) bool {
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
