package service

import (
	"bufio"
	"fmt"
	"gpanel/agent/dto"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

const sshdConfig = "/etc/ssh/sshd_config"

type SSHService struct{}

func NewSSHService() *SSHService {
	return &SSHService{}
}

func (s *SSHService) GetSSHInfo() (dto.SSHInfo, error) {
	info := dto.SSHInfo{
		Port:            22,
		ListenAddress:   "0.0.0.0",
		PasswordAuth:    "yes",
		PubkeyAuth:      "yes",
		PermitRootLogin: "yes",
	}

	// Check service status
	out, _ := exec.Command("systemctl", "is-active", "sshd").Output()
	info.IsActive = strings.TrimSpace(string(out)) == "active"

	// Parse config
	f, err := os.Open(sshdConfig)
	if err != nil {
		return info, nil
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}
		key, val := parts[0], parts[1]
		switch key {
		case "Port":
			if p, err := strconv.Atoi(val); err == nil {
				info.Port = p
			}
		case "ListenAddress":
			info.ListenAddress = val
		case "PasswordAuthentication":
			info.PasswordAuth = strings.ToLower(val)
		case "PubkeyAuthentication":
			info.PubkeyAuth = strings.ToLower(val)
		case "PermitRootLogin":
			info.PermitRootLogin = strings.ToLower(val)
		}
	}
	return info, nil
}

func (s *SSHService) OperateSSH(operation string) error {
	switch operation {
	case "start", "stop", "restart":
		return exec.Command("systemctl", operation, "sshd").Run()
	default:
		return fmt.Errorf("invalid operation: %s", operation)
	}
}

func (s *SSHService) UpdateSSHConfig(key, value string) error {
	data, err := os.ReadFile(sshdConfig)
	if err != nil {
		return err
	}

	lines := strings.Split(string(data), "\n")
	found := false
	re := regexp.MustCompile(`(?i)^#?\s*` + regexp.QuoteMeta(key) + `\s+`)
	for i, line := range lines {
		if re.MatchString(line) {
			lines[i] = key + " " + value
			found = true
			break
		}
	}
	if !found {
		lines = append(lines, key+" "+value)
	}

	return os.WriteFile(sshdConfig, []byte(strings.Join(lines, "\n")), 0644)
}

func (s *SSHService) GetSSHSessions() ([]dto.SSHSession, error) {
	out, err := exec.Command("who").Output()
	if err != nil {
		return nil, err
	}

	var sessions []dto.SSHSession
	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	// who output: username tty date time (host)
	re := regexp.MustCompile(`^(\S+)\s+(\S+)\s+(.+?)\s+\(([^)]+)\)`)
	for scanner.Scan() {
		line := scanner.Text()
		m := re.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		username, tty, loginTime, host := m[1], m[2], m[3], m[4]
		// Only SSH sessions have a host in parentheses that looks like an IP
		pid := s.getSessionPID(tty)
		sessions = append(sessions, dto.SSHSession{
			PID:       pid,
			Username:  username,
			Terminal:  tty,
			Host:      host,
			LoginTime: loginTime,
		})
	}
	return sessions, nil
}

func (s *SSHService) getSessionPID(tty string) int {
	// Get PID of the process owning the tty
	out, err := exec.Command("ps", "-t", tty, "-o", "pid=", "--no-headers").Output()
	if err != nil {
		return 0
	}
	lines := strings.Fields(strings.TrimSpace(string(out)))
	if len(lines) == 0 {
		return 0
	}
	pid, _ := strconv.Atoi(lines[0])
	return pid
}

func (s *SSHService) KillSSHSession(pid int) error {
	if pid <= 0 {
		return fmt.Errorf("invalid pid")
	}
	return exec.Command("kill", "-9", strconv.Itoa(pid)).Run()
}

func (s *SSHService) GetSSHLogs(page, pageSize int, status, info string) (dto.SSHLogRes, error) {
	logFile := "/var/log/auth.log"
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		logFile = "/var/log/secure"
	}

	f, err := os.Open(logFile)
	if err != nil {
		return dto.SSHLogRes{Items: []dto.SSHLogItem{}}, nil
	}
	defer f.Close()

	// Patterns for accepted and failed logins
	acceptedRe := regexp.MustCompile(`(\w+\s+\d+\s+\d+:\d+:\d+).*sshd.*Accepted\s+(\w+)\s+for\s+(\S+)\s+from\s+(\S+)\s+port\s+(\d+)`)
	failedRe := regexp.MustCompile(`(\w+\s+\d+\s+\d+:\d+:\d+).*sshd.*Failed\s+(\w+)\s+for\s+(?:invalid user\s+)?(\S+)\s+from\s+(\S+)\s+port\s+(\d+)`)

	var items []dto.SSHLogItem
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		var item dto.SSHLogItem

		if m := acceptedRe.FindStringSubmatch(line); m != nil {
			item = dto.SSHLogItem{Date: m[1], AuthMode: m[2], User: m[3], Address: m[4], Port: m[5], Status: "success"}
		} else if m := failedRe.FindStringSubmatch(line); m != nil {
			item = dto.SSHLogItem{Date: m[1], AuthMode: m[2], User: m[3], Address: m[4], Port: m[5], Status: "failed"}
		} else {
			continue
		}

		if status != "" && status != "all" && item.Status != status {
			continue
		}
		if info != "" && !strings.Contains(item.Address, info) && !strings.Contains(item.User, info) {
			continue
		}
		items = append(items, item)
	}

	total := int64(len(items))
	start := (page - 1) * pageSize
	end := start + pageSize
	if start > len(items) {
		start = len(items)
	}
	if end > len(items) {
		end = len(items)
	}

	// Return in reverse order (newest first)
	reversed := make([]dto.SSHLogItem, len(items))
	for i, v := range items {
		reversed[len(items)-1-i] = v
	}

	return dto.SSHLogRes{Total: total, Items: reversed[start:end]}, nil
}
