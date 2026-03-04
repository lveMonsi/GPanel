package firewall

import (
	"regexp"
	"strconv"
	"strings"
	"sync"
)

// PortStatus 端口状态缓存
type PortStatus struct {
	Port    string
	Used    bool
	Process string
}

// portStatusCache 端口状态缓存
var portStatusCache = make(map[string]PortStatus)
var portStatusMutex sync.RWMutex

// CheckPortUsed 检查端口是否被占用
// 返回: 是否被占用, 进程名称
func CheckPortUsed(port string, protocol string) (bool, string) {
	// 尝试使用 ss 命令（更现代）
	output, err := Exec("ss", "-tulpn")
	if err != nil {
		// 回退到 netstat
		output, err = Exec("netstat", "-tulpn")
		if err != nil {
			return false, ""
		}
	}

	return parsePortStatus(output, port, protocol)
}

// parsePortStatus 解析端口状态
func parsePortStatus(output string, port string, protocol string) (bool, string) {
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// 检查协议匹配
		proto := "tcp"
		if strings.HasPrefix(line, "udp") || strings.HasPrefix(line, "UDP") {
			proto = "udp"
		}

		// 如果指定了协议，则检查协议匹配
		if protocol != "" && protocol != "tcp/udp" && proto != protocol {
			continue
		}

		// 解析端口号
		// ss 格式: LISTEN 0 128 *:22 *:* users:(("sshd",pid=1234,fd=3))
		// netstat 格式: tcp 0 0 0.0.0.0:22 0.0.0.0:* LISTEN 1234/sshd

		// 匹配端口格式 :端口 或 IP:端口
		portPattern := regexp.MustCompile(`[:` + port + `\b]`)
		if strings.Contains(line, ":"+port+" ") || strings.Contains(line, ":"+port+"\t") || portPattern.MatchString(line) {
			// 提取进程名
			process := extractProcessName(line)
			return true, process
		}
	}

	return false, ""
}

// extractProcessName 从输出行中提取进程名
func extractProcessName(line string) string {
	// ss 格式: users:(("sshd",pid=1234,fd=3))
	if strings.Contains(line, "users:((") {
		re := regexp.MustCompile(`"([^"]+)"`)
		matches := re.FindStringSubmatch(line)
		if len(matches) > 1 {
			return matches[1]
		}
	}

	// netstat 格式: 1234/sshd
	fields := strings.Fields(line)
	if len(fields) >= 7 {
		lastField := fields[len(fields)-1]
		if strings.Contains(lastField, "/") {
			parts := strings.Split(lastField, "/")
			if len(parts) >= 2 {
				return parts[1]
			}
		}
	}

	return ""
}

// CheckPortsUsed 批量检查端口占用状态
// 返回: 端口 -> PortStatus 的映射
func CheckPortsUsed(ports []string, protocol string) map[string]PortStatus {
	result := make(map[string]PortStatus)

	// 获取所有端口监听信息
	output, err := Exec("ss", "-tulpn")
	if err != nil {
		output, err = Exec("netstat", "-tulpn")
		if err != nil {
			return result
		}
	}

	// 解析所有端口状态
	parseAllPortStatus(output, result, ports, protocol)

	return result
}

// parseAllPortStatus 解析所有端口状态
func parseAllPortStatus(output string, result map[string]PortStatus, ports []string, defaultProtocol string) {
	lines := strings.Split(output, "\n")
	portSet := make(map[string]bool)
	for _, p := range ports {
		portSet[p] = true
	}

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// 检查协议
		proto := "tcp"
		if strings.HasPrefix(line, "udp") || strings.HasPrefix(line, "UDP") {
			proto = "udp"
		}

		// 提取端口号
		// 匹配 :端口号 格式
		portPattern := regexp.MustCompile(`:(\d+)\s`)
		matches := portPattern.FindStringSubmatch(line)
		if len(matches) > 1 {
			port := matches[1]
			// 检查是否在目标端口列表中
			if portSet[port] {
				process := extractProcessName(line)
				key := port + "_" + proto
				result[key] = PortStatus{
					Port:    port,
					Used:    true,
					Process: process,
				}
			}
		}
	}

	// 标记未占用的端口
	for _, port := range ports {
		tcpKey := port + "_tcp"
		udpKey := port + "_udp"

		// 如果指定了协议
		if defaultProtocol != "" && defaultProtocol != "tcp/udp" {
			key := port + "_" + defaultProtocol
			if _, exists := result[key]; !exists {
				result[key] = PortStatus{
					Port: port,
					Used: false,
				}
			}
		} else {
			// tcp/udp 都检查
			if _, exists := result[tcpKey]; !exists {
				result[tcpKey] = PortStatus{
					Port: port,
					Used: false,
				}
			}
			if _, exists := result[udpKey]; !exists {
				result[udpKey] = PortStatus{
					Port: port,
					Used: false,
				}
			}
		}
	}
}

// ParsePortRange 解析端口范围，返回端口列表
// 支持: "80", "80-100", "80,443,8080"
func ParsePortRange(portStr string) []string {
	var ports []string

	// 处理逗号分隔
	if strings.Contains(portStr, ",") {
		for _, p := range strings.Split(portStr, ",") {
			p = strings.TrimSpace(p)
			if p != "" {
				ports = append(ports, parseSinglePortOrRange(p)...)
			}
		}
		return ports
	}

	return parseSinglePortOrRange(portStr)
}

// parseSinglePortOrRange 解析单个端口或端口范围
func parseSinglePortOrRange(portStr string) []string {
	var ports []string

	// 处理端口范围 80-100
	if strings.Contains(portStr, "-") {
		parts := strings.Split(portStr, "-")
		if len(parts) == 2 {
			start, err1 := strconv.Atoi(strings.TrimSpace(parts[0]))
			end, err2 := strconv.Atoi(strings.TrimSpace(parts[1]))
			if err1 == nil && err2 == nil && start <= end {
				for i := start; i <= end; i++ {
					ports = append(ports, strconv.Itoa(i))
				}
			}
		}
		return ports
	}

	// 单个端口
	portStr = strings.TrimSpace(portStr)
	if portStr != "" {
		ports = append(ports, portStr)
	}

	return ports
}

// IsPortUsedWithCache 使用缓存检查端口状态
func IsPortUsedWithCache(port string, protocol string) (bool, string) {
	key := port + "_" + protocol

	portStatusMutex.RLock()
	status, exists := portStatusCache[key]
	portStatusMutex.RUnlock()

	if exists {
		return status.Used, status.Process
	}

	// 缓存不存在，实时检查
	used, process := CheckPortUsed(port, protocol)

	portStatusMutex.Lock()
	portStatusCache[key] = PortStatus{
		Port:    port,
		Used:    used,
		Process: process,
	}
	portStatusMutex.Unlock()

	return used, process
}

// ClearPortStatusCache 清除端口状态缓存
func ClearPortStatusCache() {
	portStatusMutex.Lock()
	portStatusCache = make(map[string]PortStatus)
	portStatusMutex.Unlock()
}
