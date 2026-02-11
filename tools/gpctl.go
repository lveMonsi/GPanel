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
	serviceName = "gpanel"
	binaryPath  = "/opt/gpanel/gpanel"
	configPath  = "/opt/gpanel/config.yaml"
)

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
	fmt.Println("  start               启动 GPanel 服务")
	fmt.Println("  stop                停止 GPanel 服务")
	fmt.Println("  restart             重启 GPanel 服务")
	fmt.Println("  uninstall           卸载 GPanel 服务")
	fmt.Println("  user-info           获取 GPanel 用户信息")
	fmt.Println("  --help, -h          显示帮助信息")
}

func handleStatus() {
	if !isRoot() {
		fmt.Println("错误: 需要 root 权限")
		os.Exit(1)
	}

	// 检查 systemd 服务
	if !isSystemdService() {
		fmt.Println("错误: GPanel 服务未安装")
		os.Exit(1)
	}

	// 获取服务状态
	cmd := exec.Command("systemctl", "status", serviceName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("服务状态异常: %v\n", err)
		os.Exit(1)
	}

	// 检查服务是否正在运行
	if isServiceRunning() {
		fmt.Println("\n✓ GPanel 服务正在运行")
	} else {
		fmt.Println("\n✗ GPanel 服务未运行")
	}
}

func handleStart() {
	if !isRoot() {
		fmt.Println("错误: 需要 root 权限")
		os.Exit(1)
	}

	if !isSystemdService() {
		fmt.Println("错误: GPanel 服务未安装")
		os.Exit(1)
	}

	fmt.Println("正在启动 GPanel 服务...")
	cmd := exec.Command("systemctl", "start", serviceName)
	if err := cmd.Run(); err != nil {
		fmt.Printf("启动失败: %v\n", err)
		os.Exit(1)
	}

	// 等待服务启动
	time.Sleep(2 * time.Second)

	if isServiceRunning() {
		fmt.Println("✓ GPanel 服务启动成功")
	} else {
		fmt.Println("✗ GPanel 服务启动失败")
		os.Exit(1)
	}
}

func handleStop() {
	if !isRoot() {
		fmt.Println("错误: 需要 root 权限")
		os.Exit(1)
	}

	if !isSystemdService() {
		fmt.Println("错误: GPanel 服务未安装")
		os.Exit(1)
	}

	fmt.Println("正在停止 GPanel 服务...")
	cmd := exec.Command("systemctl", "stop", serviceName)
	if err := cmd.Run(); err != nil {
		fmt.Printf("停止失败: %v\n", err)
		os.Exit(1)
	}

	// 等待服务停止
	time.Sleep(2 * time.Second)

	if !isServiceRunning() {
		fmt.Println("✓ GPanel 服务已停止")
	} else {
		fmt.Println("✗ GPanel 服务停止失败")
		os.Exit(1)
	}
}

func handleRestart() {
	if !isRoot() {
		fmt.Println("错误: 需要 root 权限")
		os.Exit(1)
	}

	if !isSystemdService() {
		fmt.Println("错误: GPanel 服务未安装")
		os.Exit(1)
	}

	fmt.Println("正在重启 GPanel 服务...")
	cmd := exec.Command("systemctl", "restart", serviceName)
	if err := cmd.Run(); err != nil {
		fmt.Printf("重启失败: %v\n", err)
		os.Exit(1)
	}

	// 等待服务重启
	time.Sleep(2 * time.Second)

	if isServiceRunning() {
		fmt.Println("✓ GPanel 服务重启成功")
	} else {
		fmt.Println("✗ GPanel 服务重启失败")
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

	// 停止服务
	if isSystemdService() {
		exec.Command("systemctl", "stop", serviceName).Run()
		exec.Command("systemctl", "disable", serviceName).Run()
	}

	// 删除 systemd 服务文件
	serviceFile := "/etc/systemd/system/" + serviceName + ".service"
	if fileExists(serviceFile) {
		os.Remove(serviceFile)
		exec.Command("systemctl", "daemon-reload").Run()
	}

	// 删除二进制文件
	if fileExists(binaryPath) {
		os.Remove(binaryPath)
	}

	// 删除配置文件（可选）
	fmt.Print("是否删除配置文件? (yes/no): ")
	var deleteConfig string
	fmt.Scanln(&deleteConfig)
	if strings.ToLower(deleteConfig) == "yes" {
		if fileExists(configPath) {
			os.Remove(configPath)
		}
		// 删除数据目录
		dataDir := "/opt/gpanel"
		if dirExists(dataDir) {
			os.RemoveAll(dataDir)
		}
		// 删除日志目录
		logDir := "/var/log/gpanel"
		if dirExists(logDir) {
			os.RemoveAll(logDir)
		}
	}

	// 删除 gpctl 本身
	selfPath, _ := os.Executable()
	if fileExists(selfPath) {
		os.Remove(selfPath)
	}

	fmt.Println("✓ GPanel 服务卸载完成")
}

func handleUserInfo() {
	// 尝试从配置文件读取用户信息
	if fileExists(configPath) {
		fmt.Println("配置文件路径:", configPath)
		fmt.Println("请查看配置文件获取用户信息")
	} else {
		fmt.Println("配置文件不存在:", configPath)
	}

	// 尝试从 API 获取用户信息
	port := getServerPort()
	if port == "" {
		port = "8080"
	}

	url := fmt.Sprintf("http://localhost:%s/api/v1/config", port)
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(url)
	if err == nil && resp.StatusCode == 200 {
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		var config map[string]interface{}
		if err := json.Unmarshal(body, &config); err == nil {
			fmt.Println("\n服务配置信息:")
			if data, ok := config["data"].(map[string]interface{}); ok {
				if server, ok := data["server"].(map[string]interface{}); ok {
					if port, ok := server["port"].(string); ok {
						fmt.Printf("  端口: %s\n", port)
					}
					if mode, ok := server["mode"].(string); ok {
						fmt.Printf("  模式: %s\n", mode)
					}
				}
			}
		}
	}

	fmt.Println("\n提示: 默认用户名和密码请查看安装文档或配置文件")
}

func isRoot() bool {
	return os.Geteuid() == 0
}

func isSystemdService() bool {
	serviceFile := "/etc/systemd/system/" + serviceName + ".service"
	return fileExists(serviceFile)
}

func isServiceRunning() bool {
	cmd := exec.Command("systemctl", "is-active", serviceName)
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

func getServerPort() string {
	if fileExists(configPath) {
		content, err := os.ReadFile(configPath)
		if err != nil {
			return ""
		}
		lines := strings.Split(string(content), "\n")
		for _, line := range lines {
			if strings.Contains(line, "port:") {
				parts := strings.Split(line, ":")
				if len(parts) >= 2 {
					return strings.TrimSpace(strings.Trim(parts[1], `"`))
				}
			}
		}
	}
	return ""
}