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

func sshServiceName() string {
	if err := exec.Command("systemctl", "is-active", "--quiet", "sshd").Run(); err == nil {
		return "sshd"
	}
	return "ssh"
}

func (s *SSHService) GetSSHInfo() (dto.SSHInfo, error) {
	info := dto.SSHInfo{
		Port:            22,
		ListenAddress:   "0.0.0.0",
		PasswordAuth:    "yes",
		PubkeyAuth:      "yes",
		PermitRootLogin: "yes",
		UseDNS:          "yes",
	}

	svc := sshServiceName()
	out, _ := exec.Command("systemctl", "is-active", svc).Output()
	info.IsActive = strings.TrimSpace(string(out)) == "active"

	out, _ = exec.Command("systemctl", "is-enabled", svc).Output()
	info.AutoStart = strings.TrimSpace(string(out)) == "enabled"

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
		case "UseDNS":
			info.UseDNS = strings.ToLower(val)
		}
	}
	return info, nil
}

func (s *SSHService) OperateSSH(operation string) error {
	svc := sshServiceName()
	switch operation {
	case "start", "stop", "restart", "enable", "disable":
		return exec.Command("systemctl", operation, svc).Run()
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
	out, err := exec.Command("who", "-u").Output()
	if err != nil {
		return nil, err
	}

	var sessions []dto.SSHSession
	// who -u format: user tty date time idle pid (host)
	re := regexp.MustCompile(`^(\S+)\s+(\S+)\s+(\S+\s+\S+)\s+\S+\s+(\d+)\s+\(([^)]+)\)`)
	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	for scanner.Scan() {
		m := re.FindStringSubmatch(scanner.Text())
		if m == nil {
			continue
		}
		pid, _ := strconv.Atoi(m[4])
		sessions = append(sessions, dto.SSHSession{
			PID:       pid,
			Username:  m[1],
			Terminal:  m[2],
			LoginTime: m[3],
			Host:      m[5],
		})
	}
	return sessions, nil
}

func (s *SSHService) KillSSHSession(pid int) error {
	if pid <= 0 {
		return fmt.Errorf("invalid pid")
	}
	return exec.Command("kill", "-9", strconv.Itoa(pid)).Run()
}

func (s *SSHService) GetSSHLogs(page, pageSize int, status, info string) (dto.SSHLogRes, error) {
	var lines []string

	for _, f := range []string{"/var/log/auth.log.1", "/var/log/auth.log", "/var/log/secure"} {
		data, err := os.ReadFile(f)
		if err == nil {
			lines = append(lines, strings.Split(string(data), "\n")...)
		}
	}

	if len(lines) == 0 {
		out, err := exec.Command("journalctl", "-u", "sshd", "--no-pager", "-n", "10000").Output()
		if err == nil {
			lines = strings.Split(string(out), "\n")
		}
	}

	acceptedRe := regexp.MustCompile(`Accepted (password|publickey|keyboard-interactive) for (\S+) from (\S+) port (\d+)`)
	failedRe := regexp.MustCompile(`Failed (password|publickey|keyboard-interactive) for (?:invalid user )?(\S+) from (\S+) port (\d+)`)
	invalidRe := regexp.MustCompile(`Invalid user (\S+) from (\S+) port (\d+)`)
	dateRe := regexp.MustCompile(`^(\w+\s+\d+\s+\d+:\d+:\d+)`)

	var items []dto.SSHLogItem
	for _, line := range lines {
		var item dto.SSHLogItem
		date := ""
		if m := dateRe.FindStringSubmatch(line); m != nil {
			date = m[1]
		}
		if m := acceptedRe.FindStringSubmatch(line); m != nil {
			item = dto.SSHLogItem{Date: date, AuthMode: m[1], User: m[2], Address: m[3], Port: m[4], Status: "success"}
		} else if m := failedRe.FindStringSubmatch(line); m != nil {
			item = dto.SSHLogItem{Date: date, AuthMode: m[1], User: m[2], Address: m[3], Port: m[4], Status: "failed"}
		} else if m := invalidRe.FindStringSubmatch(line); m != nil {
			item = dto.SSHLogItem{Date: date, AuthMode: "password", User: m[1], Address: m[2], Port: m[3], Status: "failed"}
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
	// reverse
	for i, j := 0, len(items)-1; i < j; i, j = i+1, j-1 {
		items[i], items[j] = items[j], items[i]
	}
	start := (page - 1) * pageSize
	if start > len(items) {
		start = len(items)
	}
	end := start + pageSize
	if end > len(items) {
		end = len(items)
	}
	return dto.SSHLogRes{Total: total, Items: items[start:end]}, nil
}
